# Skytap Go Support

This repository holds support for the Skytap REST API using the Go programming language, as well as docker-machine driver
support based on this API client.

## API Tests

Change to a project root directory like ~/work/skytap
    
    export GOPATH=`pwd`
    go get -t github.com/skytap/go-skytap
    go test -v github.com/skytap/go-skytap/api
     
    
    