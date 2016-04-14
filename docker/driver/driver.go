package driver

import (
	"fmt"

	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	dockerSsh "github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/skytap/docker-machine-driver-skytap/api"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"time"
)

const (
	defaultEnvironmentId = "New"
	defaultVPNId         = ""
	driverName           = "skytap"
)

// Driver is the driver used when no driver is selected. It is used to
// connect to existing Docker hosts by specifying the URL of the host as
// an option.
type Driver struct {
	*drivers.BaseDriver
	ClientCredentials api.SkytapCredentials
	DeviceConfig      deviceConfig
	Vm                api.VirtualMachine
	LogLevel          logrus.Level
	LastState         state.State
}

type deviceConfig struct {
	SourceVMId    string
	EnvironmentId string
	VPNId         string
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		ClientCredentials: api.SkytapCredentials{},
		DeviceConfig:      deviceConfig{},
		Vm:                api.VirtualMachine{},
		LogLevel:          logrus.WarnLevel,
		LastState: state.None,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "skytap-user-id",
			Usage:  "Skytap user id",
			EnvVar: "SKYTAP_USER_ID",
		},
		mcnflag.StringFlag{
			Name:   "skytap-api-security-token",
			Usage:  "Your secrect security token",
			EnvVar: "SKYTAP_API_SECURITY_TOKEN",
		},
		mcnflag.StringFlag{
			Name:   "skytap-vm-id",
			Usage:  "Id for the VM template to use",
			EnvVar: "SKYTAP_VM_ID",
		},
		mcnflag.StringFlag{
			Name:   "skytap-env-id",
			Usage:  "Id for the environment to add the VM to. Leave blank to create to a new environment",
			Value:  defaultEnvironmentId,
			EnvVar: "SKYTAP_ENV_ID",
		},
		mcnflag.StringFlag{
			Name:   "skytap-vpn-id",
			Usage:  "VPN id to connect to the environment",
			Value:  defaultVPNId,
			EnvVar: "SKYTAP_VPN_ID",
		},
		mcnflag.StringFlag{
			Name:   "skytap-ssh-user",
			Usage:  "SSH user",
			Value:  drivers.DefaultSSHUser,
			EnvVar: "SKYTAP_SSH_USER",
		},
		mcnflag.StringFlag{
			Name:   "skytap-ssh-key",
			Usage:  "SSH private key path (if not provided, identities in ssh-agent will be used)",
			Value:  "",
			EnvVar: "SKYTAP_SSH_KEY",
		},
		mcnflag.IntFlag{
			Name:   "skytap-ssh-port",
			Usage:  "SSH port",
			Value:  drivers.DefaultSSHPort,
			EnvVar: "SKYTAP_SSH_PORT",
		},
		mcnflag.StringFlag{
			Name:  "skytap-api-logging-level",
			Usage: "The logging level to use when running api commands.",
			Value: "info",
		},
	}
}

func (d *Driver) Create() error {
	d.SetLogLevel()
	log.Info("Creating docker machine in Skytap")
	log.Debug("Skytap client auth: %s", d.ClientCredentials)

	client := *api.NewSkytapClientFromCredentials(d.ClientCredentials)

	var env *api.Environment = nil
	var err error = nil
	if d.DeviceConfig.EnvironmentId == defaultEnvironmentId {
		vm, err := api.GetVirtualMachine(client, d.DeviceConfig.SourceVMId)
		if err != nil {
			return err
		}

		template, err := vm.GetTemplate(client)
		if err != nil {
			return err
		}
		if template == nil {
			return errors.New(fmt.Sprintf("Specified VM %s is not associated with a template", d.DeviceConfig.SourceVMId))
		}

		env, err = api.CreateNewEnvironment(client, template.Id)
		if err != nil {
			return err
		}
		d.DeviceConfig.EnvironmentId = env.Id
		env, err = env.WaitUntilReady(client)
		if err != nil {
			return err
		}
	} else {
		env, err = api.GetEnvironment(client, d.DeviceConfig.EnvironmentId)
		if err != nil {
			return err
		}
		env, err = env.WaitUntilReady(client)
		if err != nil {
			return err
		}
		env, err = env.AddVirtualMachine(client, d.DeviceConfig.SourceVMId)
		if err != nil {
			return err
		}
	}

	//TODO: Multiple networks?
	vpnId := d.DeviceConfig.VPNId
	if vpnId != "" {
		attached := false
		for _, network := range env.Networks {
			// Look to see if there is an attached VPN that we simply need to connect
			for _, attachment := range network.VpnAttachments {
				if attachment.Vpn.Id == vpnId {
					attached = true
					if !attachment.Connected {
						if err = network.ConnectToVpn(client, env.Id, vpnId); err != nil {
							return err
						}
					}
					break
				}
			}
		}
		if !attached {
			_, err := env.Networks[0].AttachToVpn(client, env.Id, vpnId)
			if err != nil {
				return err
			}
			if err = env.Networks[0].ConnectToVpn(client, env.Id, vpnId); err != nil {
				return err
			}
		}
		env, err = env.WaitUntilReady(client)
		if err != nil {
			return err
		}
	}

	vm := env.Vms[len(env.Vms)-1]

	// Rename interface to match name of machine from docker-machine's perspective.
	_, err = vm.RenameNetworkInterface(client, env.Id, vm.Interfaces[0].Id, d.MachineName)
	if err != nil {
		return err
	}

	// Just added a VM so pick the last one
	started, err := vm.Start(client)
	if err != nil {
		return err
	}

	renamed, err := started.WaitUntilReady(client)
	if err != nil {
		return err
	}

	d.Vm = *renamed

	if err = d.refreshIpAddress(); err != nil {
		return err
	}

	err = d.GenerateSshKeyAndCopy()
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) refreshVm() error {
	client := *api.NewSkytapClientFromCredentials(d.ClientCredentials)
	vm, err := api.GetVirtualMachine(client, d.Vm.Id)
	if err != nil {
		return err
	}
	d.Vm = *vm
	return nil
}

func (d *Driver) refreshIpAddress() error {
	//TODO: What about multiple interfaces?
	if d.DeviceConfig.VPNId != defaultVPNId {
		var correctNat api.VpnNatAddress
		for _, a := range d.Vm.Interfaces[0].NatAddresses.VpnNatAddresses {
			if a.VpnId == d.DeviceConfig.VPNId {
				correctNat = a
			}
		}
		if correctNat.IpAddress == "" {
			return errors.New(fmt.Sprintf("Unable to find network NAT address for correct VPN in VM %s", d.Vm.Id))
		}
		d.IPAddress = correctNat.IpAddress
	} else {
		d.IPAddress = d.Vm.Interfaces[0].Ip
	}
	return nil
}

/*
 Generates a new SSH keypair, uses password auth to create the .ssh/authorized_keys file for later docker-machine access.
*/
func (d *Driver) GenerateSshKeyAndCopy() error {
	d.SetLogLevel()
	client := *api.NewSkytapClientFromCredentials(d.ClientCredentials)
	creds, err := d.Vm.GetCredentials(client)
	if err != nil {
		return err
	}
	var foundCred *api.VmCredential
	for _, c := range creds {
		user, err := c.Username()
		if err != nil {
			return err
		}
		if user == d.SSHUser {
			foundCred = &c
			break
		}
	}
	if foundCred == nil {
		return fmt.Errorf("Virtual machine does not have credentials stored for specified SSH user %s", d.SSHUser)
	}

	password, err := foundCred.Password()
	if err != nil {
		return err
	}

	success := false
	for i := 0; i < 5 && !success; i++ {
		sleepTime := 10 * time.Second
		log.Infof("Sleeping for %s, so that SSH services can come up properly", sleepTime)
		time.Sleep(sleepTime)

		err = d.DoSshCopy(client, password)
		if err != nil {
			log.Warnf("Error attempting to connect to SSH, will retry %d times: %s", 5-i, err)
		} else {
			success = true
		}
	}
	if !success {
		log.Errorf("Unable to SSH to target machine to copy public key credentials after retries: %s", err)
		return err
	}
	return nil
}

func (d *Driver) DoSshCopy(client api.SkytapClient, password string) error {

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", d.IPAddress, d.SSHPort), &ssh.ClientConfig{
		User: d.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	})

	if err != nil {
		return err
	}

	if err = runRemoteBashCommand(sshClient, "mkdir -p ~/.ssh"); err != nil {
		return err
	}

	if err = dockerSsh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	scpSession, err := sshClient.NewSession()
	if err != nil {
		return err
	}

	pubKeyFile := d.GetSSHKeyPath() + ".pub"
	destFile := "docker-machine-id_rsa.pub"
	if err = scp.CopyPath(pubKeyFile, destFile, scpSession); err != nil {
		return err
	}

	if err = runRemoteBashCommand(sshClient, fmt.Sprintf("cat %s >> ~/.ssh/authorized_keys; chmod 600 ~/.ssh/authorized_keys", destFile)); err != nil {
		return err
	}

	return nil
}

func runRemoteBashCommand(sshClient *ssh.Client, cmd string) error {
	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	err = session.Run(cmd)
	return err
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", errors.New("No IP address available in Skytap driver")
	}
	log.Debugf("IP Address: %s", d.IPAddress)
	return d.IPAddress, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) GetURL() (string, error) {
	// Driver code will only get current state if we return a blank string here, so
	// only return a valid URL if we believe we are running
	if d.LastState == state.Running {
		d.SetLogLevel()
		d.refreshVm()
		d.refreshIpAddress()
		ip, err := d.GetIP()
		if err != nil {
			return "", err
		}
		if d.Vm.Runstate == api.RunStateStart {
			return fmt.Sprintf("tcp://%s:2376", ip), nil
		} else {
			return "", nil
		}
	} else {
		return "", nil
	}
}

func (d *Driver) GetState() (state.State, error) {
	d.SetLogLevel()
	client := *api.NewSkytapClientFromCredentials(d.ClientCredentials)
	vm, err := api.GetVirtualMachine(client, d.Vm.Id)
	if err != nil {
		return state.None, err
	}
	switch vm.Runstate {
	case api.RunStateBusy:
		d.LastState = state.None
		return state.None, errors.New("VM is busy, wait and try again")
	case api.RunStateStop:
		d.LastState = state.Stopped
		return state.Stopped, nil
	case api.RunStateStart:
		d.LastState = state.Running
		return state.Running, nil
	case api.RunStatePause:
		d.LastState = state.Paused
		return state.Paused, nil
	default:
		d.LastState = state.None
		return state.None, errors.New("Unhandled VM state: " + vm.Runstate)
	}
}

func (d *Driver) Kill() error {
	d.SetLogLevel()
	client := *api.NewSkytapClientFromCredentials(d.ClientCredentials)

	_, err := d.Vm.Kill(client)
	return err
}

func (d *Driver) Remove() error {
	d.SetLogLevel()
	client := *api.NewSkytapClientFromCredentials(d.ClientCredentials)
	err := api.DeleteVirtualMachine(client, d.Vm.Id)
	return err
}

func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	user := flags.String("skytap-user-id")
	key := flags.String("skytap-api-security-token")
	d.ClientCredentials = api.SkytapCredentials{user, key}

	d.SetSwarmConfigFromFlags(flags)
	d.SSHUser = flags.String("skytap-ssh-user")
	d.SSHPort = flags.Int("skytap-ssh-port")

	envId := flags.String("skytap-env-id")
	if envId == "" {
		envId = defaultEnvironmentId
	}
	d.DeviceConfig = deviceConfig{
		SourceVMId:    flags.String("skytap-vm-id"),
		EnvironmentId: envId,
		VPNId:         flags.String("skytap-vpn-id"),
	}

	if err := validateDeviceConfig(d.DeviceConfig); err != nil {
		return err
	}

	log.Debugf("Skytap driver configuration: %+v", d)
	logLvlStr := flags.String("skytap-api-logging-level")
	logLevel, err := logrus.ParseLevel(logLvlStr)
	if err != nil {
		log.Errorf("Unable to parse log level as specified '%s'", logLvlStr)
	}
	d.LogLevel = logLevel
	d.SetLogLevel()
	return nil
}

func validateDeviceConfig(deviceConfig deviceConfig) error {
	if deviceConfig.SourceVMId == "" {
		return errors.New("No source VPN specified")
	}
	return nil
}

func (d *Driver) Start() error {
	d.SetLogLevel()
	client := *api.NewSkytapClientFromCredentials(d.ClientCredentials)

	d.LastState = state.Starting
	_, err := d.Vm.Start(client)
	if err != nil {
		d.LastState = state.Error
		return err
	}
	d.LastState = state.Running

	return nil
}

func (d *Driver) Stop() error {
	d.SetLogLevel()
	client := *api.NewSkytapClientFromCredentials(d.ClientCredentials)
	d.LastState = state.Stopping
	_, err := d.Vm.Stop(client)
	if err != nil {
		d.LastState = state.Error
		return err
	}
	d.LastState = state.Stopped
	return err
}

func (d *Driver) SetLogLevel() {
	logrus.SetLevel(d.LogLevel)
}
