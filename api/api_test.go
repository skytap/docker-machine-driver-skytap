package api

import (
	"net/http"
	"testing"

	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"time"
	"fmt"
)

type testConfig struct {
	Username   string `json:"username"`
	ApiKey     string `json:"apiKey"`
	TemplateId string `json:"templateId"`
	VmId       string `json:"vmId"`
	VpnId      string `json:"vpnId"`
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func skytapClient(t *testing.T) SkytapClient {
	c := getTestConfig(t)
	fmt.Printf("c: %s, user: %s", c, c.Username)
	client := &http.Client{}
	return SkytapClient{
		HttpClient:  client,
		Credentials: SkytapCredentials{Username: c.Username, ApiKey: c.ApiKey},
	}
}

func getTestConfig(t *testing.T) *testConfig {
	configFile, err := os.Open("testdata/config.json")
	require.NoError(t, err, "Error reading config.json")

	jsonParser := json.NewDecoder(configFile)
	c := &testConfig{}
	err = jsonParser.Decode(c)
	require.NoError(t, err, "Error parsing config.json")
	return c
}

func getEnvironment(t *testing.T, client SkytapClient, envId string) error {

	env, err := GetEnvironment(client, envId)

	if err != nil {
		t.Error(err)
	}
	if env.Id != envId {
		t.Error("Id didn't match")
	}

	return err
}

func TestCreateEnvironment(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	require.NoError(t, err, "Error creating environment")

	err = getEnvironment(t, client, env.Id)
	require.NoError(t, err, "Error retrieving environment")

	err = DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error deleting environment")

}

func TestAddDeleteVirtualMachine(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	require.NoError(t, err, "Error creating environment")

	defer DeleteEnvironment(client, env.Id)

	// Add from template
	env2, err := env.AddVirtualMachine(client, c.VmId)
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, len(env.Vms)+1, len(env2.Vms))

	// Add from environment
	sourceEnv, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, sourceEnv.Id)
	require.NoError(t, err, "Error creating second environment")

	env3, err := env.AddVirtualMachine(client, sourceEnv.Vms[0].Id)
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, len(env2.Vms)+1, len(env3.Vms))

	// Delete last one
	DeleteVirtualMachine(client, env3.Vms[len(env3.Vms)-1].Id)
	env4, err := GetEnvironment(client, env.Id)
	require.NoError(t, err, "Error getting environment")
	require.Equal(t, len(env2.Vms), len(env4.Vms))

	// Add from same environment
	env5, err := env.AddVirtualMachine(client, env.Vms[0].Id)
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, len(env2.Vms)+1, len(env5.Vms))

	for _, value := range env5.Vms {
		require.Equal(t, "Ubuntu Server 14.04 - 64-bit", value.Name)
	}
}

func TestVmCredentials(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	require.NoError(t, err, "Error creating environment")

	defer DeleteEnvironment(client, env.Id)

	creds, err := env.Vms[0].GetCredentials(client)
	require.NoError(t, err, "Error getting VM credentials")

	user, err := creds[0].Username()
	require.NoError(t, err, "Error username")
	require.Equal(t, "root", user)

	pass, err := creds[0].Password()
	require.NoError(t, err, "Error getting password")
	require.Equal(t, "ChangeMe!", pass)

}

func TestManipulateVmRunstate(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error creating environment")

	vm, err := GetVirtualMachine(client, env.Vms[0].Id)
	require.NoError(t, err, "Error creating vm")

	stopped, err := vm.WaitUntilReady(client)
	require.NoError(t, err, "Error waiting on vm")
	require.Equal(t, RunStateStop, stopped.Runstate, "Should be stopped after waiting")

	started, err := stopped.Start(client)
	require.NoError(t, err, "Error starting VM")
	require.Equal(t, RunStateStart, started.Runstate, "Should be started")

	time.Sleep(10 * time.Second)

	// Can't get the VM to stop, waiting for a dialog
	stopped, err = started.Stop(client)
	require.NoError(t, err, "Error stopping VM")
	require.Equal(t, RunStateStop, stopped.Runstate, "Should be stopped")

	started, err = stopped.Start(client)
	require.NoError(t, err, "Error starting VM")
	require.Equal(t, RunStateStart, started.Runstate, "Should be started")

	killed, err := started.Kill(client)
	require.NoError(t, err, "Error stopping VM")
	require.Equal(t, RunStateStop, killed.Runstate, "Should be stopped/killed")

}

func TestAttachVpn(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error creating environment")

	result, err := env.Networks[0].AttachToVpn(client, env.Id, c.VpnId)
	require.NoError(t, err, "Error attaching VPN")
	log.Println(result)

	err = env.Networks[0].ConnectToVpn(client, env.Id, c.VpnId)
	require.NoError(t, err, "Error connecting VPN")

	// Now start VM and make sure it gets a good address from VPN NAT
	started, err := env.Vms[0].Start(client)
	require.NoError(t, err, "Error starting VM")
	require.Equal(t, c.VpnId, started.Interfaces[0].NatAddresses.VpnNatAddresses[0].VpnId, "Should have correct VPN id")

	err = env.Networks[0].DisconnectFromVpn(client, env.Id, c.VpnId)
	require.NoError(t, err, "Error disconnecting VPN")

	err = env.Networks[0].DetachFromVpn(client, env.Id, c.VpnId)
	require.NoError(t, err, "Error detaching VPN")

	env, err = env.WaitUntilReady(client)
	require.NoError(t, err, "Error waiting for environment")
	log.Println(env.Networks[0].VpnAttachments)
}

func TestMergeBody(t *testing.T) {
	vmId := "9285760"
	c := getTestConfig(t)

	b := &MergeTemplateBody{TemplateId: c.TemplateId, VmIds: []string{vmId}}
	j, _ := json.Marshal(b)
	println(string(j))

	j, _ = json.Marshal(&MergeEnvironmentBody{EnvironmentId: c.TemplateId, VmIds: []string{vmId}})
	println(string(j))

}
