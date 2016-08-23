# Skytap Driver for Docker Machine
##Create docker machines on [Skytap](http://www.skytap.com).
To create machines on [Skytap](http://cloud.skytap.com), you must supply 3 parameters: your Skytap User ID, your Skytap API Security Token, and the VM ID to use as the source image for the new machine.

##Installation
Visit the [releases page](https://github.com/skytap/docker-machine-driver-skytap/releases) for instructions on downloading and installing the Skytap driver.

##Help
For detailed help using the driver visit the Skytap help topic [Using the Skytap driver for Docker Machine](http://help.skytap.com/docker-machine-driver.html).

##Options
Command line flags, environment variables and default values. These flags are available during the `create` process.

| CLI flag                                 | Environment variable        | Default          | Description
| ---------------------------------------- | ----------------------------| ---------------- | -----------
| `--skytap-api-security-token`            | `SKYTAP_API_SECURITY_TOKEN` | -                | Your secret security token.
| `--skytap-container-host`                | `SKYTAP_CONTAINER_HOST`     | `false`          | Configures the VM as a container host. 
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

##Building
Run the `./build.sh` scripts to build for Linux, OS X (darwin) and Windows. The appropriate executable for the hardware should be copied to a file called docker-machine-driver-skytap somewhere in the user's PATH, so that the main docker-machine executable can locate it.

##License
Apache 2.0; see [LICENSE](LICENSE) for details
