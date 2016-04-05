package driver

import (
	"fmt"

	"errors"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/go-skytap/api"
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
	base         *drivers.BaseDriver
	client       api.SkytapClient
	deviceConfig deviceConfig
	vm           api.VirtualMachine
}

type deviceConfig struct {
	SourceVMId    string
	EnvironmentId string
	VPNId         string
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{}
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
			Name:   "skytap-icnr-env-id",
			Usage:  "The id of the environment containing the network to connect to",
			Value:  "",
			EnvVar: "SKYTAP_SSH_KEY",
		},
		mcnflag.StringFlag{
			Name:   "skytap-icnr-network-id",
			Usage:  "The id of the network within the environment to connect to.",
			Value:  "",
			EnvVar: "SKYTAP_SSH_PORT",
		},
	}
}

func (d *Driver) Create() error {
	var env *api.Environment = nil
	var err error = nil
	if d.deviceConfig.EnvironmentId == defaultEnvironmentId {
		vm, err := api.GetVirtualMachine(d.client, d.deviceConfig.SourceVMId)
		if err != nil {
			return err
		}

		template, err := vm.GetTemplate(d.client)
		if err != nil {
			return err
		}
		if template == nil {
			return errors.New(fmt.Sprintf("Specified VM %s is not associated with a template", d.deviceConfig.SourceVMId))
		}

		env, err = api.CreateNewEnvironment(d.client, template.Id)
		if err != nil {
			return err
		}
		env, err = env.WaitUntilReady(d.client)
		if err != nil {
			return err
		}

		//TODO: Multiple networks?
		if d.deviceConfig.VPNId != nil {
			_, err := env.Networks[0].AttachToVpn(d.client, env.Id, d.deviceConfig.VPNId)
			if err != nil {
				return err
			}
			err = env.Networks[0].ConnectToVpn(d.client, env.Id, d.deviceConfig.VPNId)
			if err != nil {
				return err
			}
		}

	} else {
		env, err = api.GetEnvironment(d.client, d.deviceConfig.EnvironmentId)
		if err != nil {
			return err
		}
	}

	env, err = env.WaitUntilReady(d.client)
	if err != nil {
		return err
	}

	env, err = env.AddVirtualMachine(d.client, d.deviceConfig.SourceVMId)
	if err != nil {
		return err
	}
	env, err = env.WaitUntilReady(d.client)
	if err != nil {
		return err
	}

	// Just added a VM so pick the last one
	vm, err := env.Vms[len(env.Vms)-1].Start(d.client)
	if err != nil {
		return err
	}
	vm.WaitUntilReady(d.client)

	d.vm = vm
	//TODO: What about multiple interfaces?
	if d.deviceConfig.VPNId {
		var correctNat api.VpnNatAddress = nil
		for _, a := range vm.Interfaces[0].NatAddresses.VpnNatAddresses {
			if a.VpnId == d.deviceConfig.VPNId {
				correctNat = a
			}
		}
		if correctNat == nil {
			return errors.New(fmt.Sprintf("Unable to find network NAT address for correct VPN in VM %s", vm.Id))
		}
		d.base.IPAddress = correctNat.IpAddress
	} else {
		d.base.IPAddress = vm.Interfaces[0].Ip
	}

	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetIP() (string, error) {
	return d.base.GetIP(), nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.base.GetIP(), nil
}

func (d *Driver) GetSSHKeyPath() string {
	return d.base.GetSSHKeyPath()
}

func (d *Driver) GetSSHPort() (int, error) {
	return d.base.GetSSHPort(), nil
}

func (d *Driver) GetSSHUsername() string {
	return d.base.GetSSHUsername()
}

func (d *Driver) GetURL() (string, error) {
	return fmt.Sprintf("tcp://%s:2376", d.GetIP()), nil
}

func (d *Driver) GetState() (state.State, error) {
	return state.Running, nil
}

func (d *Driver) Kill() error {
	return fmt.Errorf("hosts without a driver cannot be killed")
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Restart() error {
	return fmt.Errorf("hosts without a driver cannot be restarted")
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {

	user := flags.String("skytap-user-id")
	key := flags.String("skytap-api-security-token")
	d.client = api.NewSkytapClient(user, key)

	d.base.SetSwarmConfigFromFlags(flags)
	d.base.SSHUser = "root"
	d.base.SSHPort = 22

	d.deviceConfig = &deviceConfig{
		SourceVMId:    flags.String("skytap-vm-id"),
		EnvironmentId: flags.String("skytap-env-id"),
		VPNId:         flags.String("skytap-vpn-id"),
	}

	if err := validateDeviceConfig(d.deviceConfig); err != nil {
		return err
	}

	return nil
}

func validateDeviceConfig(deviceConfig deviceConfig) error {
	return nil
}

func (d *Driver) Start() error {
	_, err := d.vm.Start(d.client)
	return err
}

func (d *Driver) Stop() error {
	_, err := d.vm.Stop(d.client)
	return err
}
