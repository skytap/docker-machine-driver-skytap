package api

import (
	"fmt"
	"github.com/dghubble/sling"
	log "github.com/Sirupsen/logrus"
)

const (
	VmPath = "vms"
)

/*
 Skytap VM resource.
 */
type VirtualMachine struct {
	Id             string              `json:"id"`
	Name           string              `json:"name"`
	Runstate       string              `json:"runstate"`
	Error          interface{}         `json:"error"`
	TemplateUrl    string              `json:"template_url,omitempty"`
	EnvironmentUrl string              `json:"configuration_url,omitempty"`
	Interfaces     []*NetworkInterface `json:"interfaces`
}

// Paths for VMs.
func vmIdInEnvironmentPath(envId string, vmId string) string {
	return fmt.Sprintf("%s/%s/%s/%s.json", EnvironmentPath, envId, VmPath, vmId)
}
func vmIdInTemplatePath(templateId string, vmId string) string {
	return fmt.Sprintf("%s/%s/%s/%s.json", TemplatePath, templateId, VmPath, vmId)
}
func vmIdPath(vmId string) string { return fmt.Sprintf("%s/%s", VmPath, vmId) }

/*
 If VM is in a template, returns the template, otherwise nil.
 */
func (vm *VirtualMachine) GetTemplate(client SkytapClient) (*Template, error) {
	if vm.TemplateUrl == "" {
		return nil, nil
	}
	template := &Template{}
	_, err := GetSkytapResource(client, vm.TemplateUrl, template)
	return template, err
}

/*
 If a VM is in an environment, returns the environment, otherwise nil.
 */
func (vm *VirtualMachine) GetEnvironment(client SkytapClient) (*Environment, error) {
	if vm.EnvironmentUrl == "" {
		return nil, nil
	}
	env := &Environment{}
	_, err := GetSkytapResource(client, vm.EnvironmentUrl, env)
	return env, err
}

/*
 Fetch fresh representation.
 */
func (vm *VirtualMachine) Refresh(client SkytapClient) (RunstateAwareResource, error) {
	return GetVirtualMachine(client, vm.Id)
}

func (vm *VirtualMachine) RunstateStr() string { return vm.Runstate }


/*
 Waits until VM is either stopped or started.
 */
func (vm *VirtualMachine) WaitUntilReady(client SkytapClient) (*VirtualMachine, error) {
	return vm.WaitUntilInState(client, []string{RunStateStop, RunStateStart})
}

/*
  Wait until the VM is in one of the desired states.
 */
func (vm *VirtualMachine) WaitUntilInState(client SkytapClient, desiredStates []string) (*VirtualMachine, error) {
	r, err := WaitUntilInState(client, desiredStates, vm)
	v := r.(*VirtualMachine)
	return v, err
}

/*
 Starts a VM.
 */
func (vm *VirtualMachine) Start(client SkytapClient) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"vmId": vm.Id}).Info("Starting VM")

	return vm.ChangeRunstate(client, RunStateStart, RunStateStart)
}

/*
 Stops a VM. Note that some VMs may require user input and cannot be stopped with the method.
 */
func (vm *VirtualMachine) Stop(client SkytapClient) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"vmId": vm.Id}).Info("Stopping VM")

	return vm.ChangeRunstate(client, RunStateStop, RunStateStop)
}

/*
 Kills a VM forcefully.
 */
func (vm *VirtualMachine) Kill(client SkytapClient) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"vmId": vm.Id}).Info("Killing VM")

	return vm.ChangeRunstate(client, RunStateKill, RunStateStop)
}

/*
 Changes the runstate of the VM to the specified state and waits until the VM is in the desired state.
 */
func (vm *VirtualMachine) ChangeRunstate(client SkytapClient, runstate string, desiredRunstate string) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"changeState": runstate, "targetState": desiredRunstate, "vmId": vm.Id}).Info("Changing VM runstate")

	ready, err := vm.WaitUntilReady(client)
	if err != nil {
		return ready, err
	}
	changeState := func(s *sling.Sling) *sling.Sling {
		return s.Put(vmIdPath(vm.Id)).BodyJSON(&RunstateBody{Runstate: runstate})
	}
	_, err = RunSkytapRequest(client, false, nil, changeState)

	if err != nil {
		return vm, err
	}
	return vm.WaitUntilInState(client, []string{desiredRunstate})
}

/*
 Get a VM from an existing environment.
 */
func GetVirtualMachineInEnvironment(client SkytapClient, envId string, vmId string) (*VirtualMachine, error) {
	vm := &VirtualMachine{}

	getVm := func(s *sling.Sling) *sling.Sling {
		return s.Get(vmIdInEnvironmentPath(envId, vmId))
	}

	_, err := RunSkytapRequest(client, true, vm, getVm)
	return vm, err
}

/*
 Get a VM from an existing template.
 */
func GetVirtualMachineInTemplate(client SkytapClient, templateId string, vmId string) (*VirtualMachine, error) {
	vm := &VirtualMachine{}

	getVm := func(s *sling.Sling) *sling.Sling {
		return s.Get(vmIdInTemplatePath(templateId, vmId))
	}

	_, err := RunSkytapRequest(client, true, vm, getVm)
	return vm, err
}

/*
 Get a VM without reference to environment or template. The result object should contain information on its source.
 */
func GetVirtualMachine(client SkytapClient, vmId string) (*VirtualMachine, error) {
	vm := &VirtualMachine{}

	getVm := func(s *sling.Sling) *sling.Sling {
		return s.Get(vmIdPath(vmId))
	}

	_, err := RunSkytapRequest(client, false, vm, getVm)
	return vm, err
}

/*
 Delete a VM.
 */
func DeleteVirtualMachine(client SkytapClient, vmId string) error {
	log.WithFields(log.Fields{"vmId": vmId}).Info("Deleting VM")

	deleteVm := func(s *sling.Sling) *sling.Sling { return s.Delete(vmIdPath(vmId)) }
	_, err := RunSkytapRequest(client, false, nil, deleteVm)
	return err
}
