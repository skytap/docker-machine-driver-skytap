# Skytap Go Support

This repository holds support for the Skytap REST API using the Go programming language, as well as docker-machine driver
support based on this API client.

## Building

Run the ./build.sh scripts to build for Linux, OS X (darwin) and Windows. The appropriate executable for the hardware should
be copied to a file called docker-machine-driver-skytap somewhere in the user's PATH, so that the main docker-machine executable
can locate it.

## API Tests

You must provide configuration to run the tests. Copy the api/testdata/config.json.sample file to api/testdata/config.json,
and fill out the required fields. Note the the tests were run against a template containing a single Ubunto 14 server VM and
a preconfigured NAT based VPN. Other configurations may cause spurious test errors.

Change to a project root directory like ~/work/skytap
    
    export GOPATH=`pwd`
    go get -t github.com/skytap/docker-machine-driver-skytap/api
    go test -v github.com/skytap/docker-machine-driver-skytap/api
     
    
