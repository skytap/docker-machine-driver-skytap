# Skytap Driver for Docker Machine
##Create docker machines on [Skytap](http://www.skytap.com).

To create machines on [Skytap](http://www.skytap.com), you must supply 3 parameters: your Skytap User ID, your Skytap API Security Token, and the VM ID to use as the source image for the new machine.

## Configuring credentials
Before using the Skytap driver, retrieve your API credentials and ensure the credentials on the VM image are properly configured.

### Skytap API credentials
The driver uses the Skytap RESTful API to provision the VM and control state. You will supply your Skytap login and API key (logon to your Skytap account and go to the My Account page).

####Command line flags
You can use the flags `--skytap-user-id` and `--skytap-api-security-token` on the command line:

  $ docker-machine create --driver skytap --skytap-user-id janedoe --skytap-api-security-token 73bc***** {other flags} skytap-machine01

####Environment variables
Alternatively you can use environment variables:

    $ export SKYTAP_USER_ID=janedoe
    $ export SKYTAP_API_SECURITY_TOKEN=73bc*****
    $ docker-machine create --driver skytap {other flags} skytap-machine01

### VM credentials
Your source VM must be pre-configured as follows.

####SSH User
Create a user configured for SSH access. This need only be for password-based SSH. The Skytap driver will automatically create an insecure keypair for this user.

You can optionally setup keys for the user ahead of time. To configure SSH access with the insecure keypair, place the public key into the `~/.ssh/authorized_keys` file for the user. Note that OpenSSH is sensitive about file permissions. Therefore, make sure that `~/.ssh` has 0700 permissions and the authorized keys file has 0600 permissions.

####Password-less sudo
The Docker installation process requires password-less sudo. Configure it (usually using `visudo`) to allow password-less sudo for the SSH user. This can be done by adding the following line at the end of the configuration file:

  `{user} ALL=(ALL) NOPASSWD: ALL`

####VM credentials
The Skytap provider will retrieve the password for the SSH user from the VM metadata. You must store the username and password with the VM. See [VM Credentials](http://help.skytap.com/#VM_Settings_Credentials.html) for more information.

## Options

Command line flags, environment variables and default values:

| CLI flag                                 | Environment variable        | Default          | Description
| ---------------------------------------- | ----------------------------| ---------------- | -----------
| `--skytap-api-security-token`            | `SKYTAP_API_SECURITY_TOKEN` | -                | Your secret security token.
| `--skytap-env-id`                        | `SKYTAP_ENV_ID`             | `New`            | ID for the environment to add the VM to. Leave blank to create to a new environment.
| `--skytap-ssh-key`                       | `SKYTAP_SSH_KEY`            | -                | SSH private key path (if not provided, identities in ssh-agent will be used).
| `--skytap-ssh-port`                      | `SKYTAP_SSH_PORT`           | `22`             | SSH port.
| `--skytap-ssh-user`                      | `SKYTAP_SSH_USER`           | `docker`         | SSH user.
| `--skytap-user-id`                       | `SKYTAP_USER_ID`            | -                | Skytap user ID.
| `--skytap-vm-cpus`                       | `SKYTAP_VM_CPUS`            | -                | The number of CPUs for the VM. The default is what’s configured for the source VM.
| `--skytap-vm-cpuspersocket`              | `SKYTAP_VM_CPUSPERSOCKET`   | -                | Specifies how the total number of CPUs should be distributed across virtual sockets. The default is what’s configured for the source VM.
| `--skytap-vm-id`                         | `SKYTAP_VM_ID`              | -                | ID of the source VM to use.
| `--skytap-vm-ram`                        | `SKYTAP_VM_RAM`             | -                | The amount of ram, in megabytes, allocated to the VM. The default is what’s configured for the source VM.
| `--skytap-vpn-id`                        | `SKYTAP_VPN_ID`             | -                | VPN ID to connect to the environment.
| `--skytap-api-logging-level`             | `SKYTAP_API_LOGGING_LEVEL`  | `info`           | The logging level to use when running api commands.

## VM Hardware Settings
By default the new Docker host will inherit the CPU and memory settings of the source VM. These can be overridden and configured during the create process.

Using `--skytap-vm-ram` you can provide an integer representing the amount of RAM, in megabytes, to allocated to the VM. See the [RAM](http://help.skytap.com/Hardware.html?#RAMMB) section for Hardware Settings in the Skytap help for more information.

Using `--skytap-vm-cpus` and `--skytap-vm-cpuspersocket` you can configure the number of CPUs to allocate to the VM and how those CPUs should be distributed across virtual sockets. The number of CPUs (cores) per socket must be uniformly distributed across the number of sockets. If you provide a number of CPUs but don't specify the number of CPUs per socket, each CPU will be allocated to one socket - one core per virtual socket. If you provide the number of CPUs per socket you must also provide the number of CPUs. For more information on CPUs and CPUs per socket see the [CPU](http://help.skytap.com/Hardware.html?#CPUs) section for the Hardware Settings in the Skytap help.

## Logging level
When executing Docker Machine in debug mode with the -D flag, you can specify a logging level for the Skytap API calls.

-  `info`: Default level providing basic information
-  `debug`: Provides detailed information on each Skytap api call
-  `warn`: Outputs warning messages
-  `error`: Outputs error messages

## Building
Run the `./build.sh` scripts to build for Linux, OS X (darwin) and Windows. The appropriate executable for the hardware should be copied to a file called docker-machine-driver-skytap somewhere in the user's PATH, so that the main docker-machine executable can locate it.

## License
Apache 2.0; see [LICENSE](LICENSE) for details
