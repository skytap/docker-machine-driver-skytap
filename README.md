# Skytap Driver for Docker Machine (technical preview)
Create docker machines on [Skytap](http://www.skytap.com).

To create machines on [Skytap](http://www.skytap.com), you must supply 3 parameters: your Skytap User Id, your Skytap API Security Token, and the VM Id to use as the source image for the new machine.

## Configuring credentials
Before using the Skytap driver, retrieve your API credentials and ensure the credentials on the VM image are properly setup.

### Skytap API credentials
The driver uses the Skytap RESTful API to provision the VM and control state. You will supply your Skytap login and API key (available from your My Account page).

####Command line flags
You can use the flags `--skytap-user-id` and `--skytap-api-security-token` on the command line:

  $ docker-machine create --driver skytap --skytap-user-id janedoe --skytap-api-security-token 73bc***** ... skytap-machine01

####Environment variables
Alternatively you can use environment variables:

    $ export SKYTAP_USER_ID=janedoe
    $ export SKYTAP_API_SECURITY_TOKEN=73bc*****
    $ docker-machine create --driver skytap ... skytap-machine01

### VM credentials
Your source VM must be pre-configured as follows

####SSH User
Create a user configured for SSH access. This need only be for password-based SSH. The Skytap driver will automatically create an insecure keypair for this user.

You can optionally setup keys for the user ahead of time. To configure SSH access with the insecure keypair, place the public key into the `~/.ssh/authorized_keys` file for the user. Note that OpenSSH is very picky about file permissions. Therefore, make sure that `~/.ssh` has 0700 permissions and the authorized keys file has 0600 permissions.

####Password-less sudo
The Docker installation requires password-less sudo. Configure it (usually using `visudo`) to allow password-less sudo for the SSH user. This can be done with the following line at the end of the configuration file:

  `{user} ALL=(ALL) NOPASSWD: ALL`

VM credentials
The Skytap provider will retrieve the password for the SSH user from the VM metadata. You must store the username and password with the VM. See [VM Credentials](http://help.skytap.com/#VM_Settings_Credentials.html) for more information.

## Options

-   `--skytap-api-security-token`: Your secrect security token
-   `--skytap-env-id`: Id for the environment to add the VM to. Leave blank to create to a new environment
-   `--skytap-ssh-key`:	SSH private key path (if not provided, identities in ssh-agent will be used)
-   `--skytap-ssh-port`: SSH port
-   `--skytap-ssh-user`: SSH user
-   `--skytap-user-id`:	Skytap user id
-   `--skytap-vm-id`:	Id for the VM template to use
-   `--skytap-vpn-id`: VPN id to connect to the environment
-   `--skytap-api-logging-level`: The logging level to use when running api commands.


Environment variables and default values:

| CLI option                               | Environment variable        | Default          |
| ---------------------------------------- | ----------------------------| ---------------- |
| `--skytap-user-id`                       | `SKYTAP_USER_ID`            | -                |
| `--skytap-api-security-token`            | `SKYTAP_API_SECURITY_TOKEN` | -                |
| `--skytap-vm-id`                         | `SKYTAP_VM_ID`              | -                |
| `--skytap-env-id`                        | `SKYTAP_ENV_ID`             | `New`            |
| `--skytap-vpn-id`                        | `SKYTAP_VPN_ID`             | -                |
| `--skytap-ssh-user`                      | `SKYTAP_SSH_USER`           | `root`           |
| `--skytap-ssh-key`                       | `SKYTAP_SSH_KEY`            | -                |
| `--skytap-ssh-port`                      | `SKYTAP_SSH_PORT`           | `22`             |
| `--skytap-api-logging-level`             | -                           | `info`           |

## Logging level
When executing Docker Machine in debug mode with the -D flag, you can specify a logging level for the Skytap API calls.

-  `info`: Default level providing basic information
-  `debug`: Provides detailed information on each Skytap api call
-  `warn`:
-  `error`:

## Building
Run the `./build.sh` scripts to build for Linux, OS X (darwin) and Windows. The appropriate executable for the hardware should be copied to a file called docker-machine-driver-skytap somewhere in the user's PATH, so that the main docker-machine executable can locate it.

## API Tests
You must provide configuration to run the tests. Copy the `api/testdata/config.json.sample` file to `api/testdata/config.json`, and fill out the required fields. Note the the tests were run against a template containing a single Ubuntu 14.04 server VM and a preconfigured NAT based VPN. Other configurations may cause spurious test errors.

Change to a project root directory like `~/work/skytap`

    export GOPATH=`pwd`
    go get -t github.com/skytap/docker-machine-driver-skytap/api
    go test -v github.com/skytap/docker-machine-driver-skytap/api

## License
Apache 2.0; see [LICENSE](LICENSE) for details
