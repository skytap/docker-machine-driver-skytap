package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/skytap/docker-machine-driver-skytap/docker/driver"
)

func main() {
	plugin.RegisterDriver(driver.NewDriver("", ""))
}
