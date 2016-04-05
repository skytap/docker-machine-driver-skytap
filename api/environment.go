package api

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/dghubble/sling"
)

const (
	EnvironmentPath = "configurations"
)

/**
 Skytap Environment resource.
 */
type Environment struct {
	Id          string            `json:"id"`
	Url         string            `json:"url"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Error       []string          `json:"errors"`
	Runstate    string            `json:"runstate"`
	Vms         []*VirtualMachine `json:"vms"`
	Networks    []Network         `json:"networks"`
}

/*
 Request body for create commands.
 */
type CreateEnvironmentBody struct {
	TemplateId string `json:"template_id"`
}

/*
 Request body for merge commands.
 */
type MergeTemplateBody struct {
	TemplateId string   `json:"template_id"`
	VmIds      []string `json:"vm_ids"`
}

/*
 Request body for merge commands.
 */
type MergeEnvironmentBody struct {
	EnvironmentId string   `json:"merge_configuration"`
	VmIds         []string `json:"vm_ids"`
}

func environmentIdV1Path(envId string) string { return EnvironmentPath + "/" + envId }
func environmentIdPath(envId string) string   { return EnvironmentPath + "/" + envId + ".json" }

/*
 Adds a VM to an existing environment.
 */
func (e *Environment) AddVirtualMachine(client SkytapClient, vmId string) (*Environment, error) {
	log.WithFields(log.Fields{"vmId": vmId, "envId": e.Id}).Info("Adding virtual machine")

	vm, err := GetVirtualMachine(client, vmId)
	if err != nil {
		return e, err
	}

	template, err := vm.GetTemplate(client)
	if err != nil {
		return e, err
	}
	if template != nil {
		return e.MergeTemplateVirtualMachine(client, template.Id, vmId)
	}

	sourceEnv, err := vm.GetEnvironment(client)
	if err != nil {
		return e, err
	}
	if sourceEnv != nil {
		return e.MergeEnvironmentVirtualMachine(client, sourceEnv.Id, vmId)
	}

	return e, errors.New("Unable to determine source of VM, no environment or template url found")
}

func (e *Environment) RunstateStr() (string) { return e.Runstate }

func (e *Environment) Refresh(client SkytapClient) (RunstateAwareResource, error) {
	return GetEnvironment(client, e.Id)
}

func (e *Environment) WaitUntilInState(client SkytapClient, desiredStates []string) (*Environment, error) {
	r, err := WaitUntilInState(client, desiredStates, e)
	newEnv := r.(*Environment)
	return newEnv, err
}

func (e *Environment) WaitUntilReady(client SkytapClient) (*Environment, error) {
	return e.WaitUntilInState(client, []string{RunStateStop, RunStateStart})
}

/*
 Merge an environment based VM into this environment (the VM must be in an existing environment).
 */
func (e *Environment) MergeEnvironmentVirtualMachine(client SkytapClient, envId string, vmId string) (*Environment, error) {
	return e.MergeVirtualMachine(client, &MergeEnvironmentBody{EnvironmentId: envId, VmIds: []string{vmId}})
}

/*
 Merge a template based VM into this environment (the VM must be in an existing template).
 */
func (e *Environment) MergeTemplateVirtualMachine(client SkytapClient, templateId string, vmId string) (*Environment, error) {
	return e.MergeVirtualMachine(client, &MergeTemplateBody{TemplateId: templateId, VmIds: []string{vmId}})
}

/*
 Merge arbitrary VM into this environment.

 mergeBody - The correct representation of the request body, see the MergeEnvironmentVirtualMachine and MergeTemplateVirtualMachine methods.
 */
func (e *Environment) MergeVirtualMachine(client SkytapClient, mergeBody interface{}) (*Environment, error) {

	log.WithFields(log.Fields{"mergeBody": mergeBody, "envId": e.Id}).Info("Merging a VM into environment")

	merge := func(s *sling.Sling) *sling.Sling {
		return s.Put(environmentIdV1Path(e.Id)).BodyJSON(mergeBody)
	}

	newEnv := &Environment{}
	_, err := RunSkytapRequest(client, false, newEnv, merge)
	if err != nil {
		log.Errorf("Unable to add VM to environment (%s), requestBody: %s, cause: %s", e.Id, mergeBody, err)
		return nil, err
	}
	return newEnv, nil
}

/*
 Return an existing environment by id.
 */
func GetEnvironment(client SkytapClient, envId string) (*Environment, error) {
	env := &Environment{}

	getEnv := func(s *sling.Sling) *sling.Sling {
		return s.Get(environmentIdPath(envId))
	}

	_, err := RunSkytapRequest(client, true, env, getEnv)
	return env, err
}

/*
 Create a new environment from a template.
 */
func CreateNewEnvironment(client SkytapClient, templateId string) (*Environment, error) {
	log.WithFields(log.Fields{"templateId": templateId}).Info("Creating environment from template")

	env := &Environment{}

	createEnv := func(s *sling.Sling) *sling.Sling {
		return s.Post(EnvironmentPath + ".json").BodyJSON(&CreateEnvironmentBody{TemplateId: templateId})
	}

	_, err := RunSkytapRequest(client, false, env, createEnv)
	return env, err
}

/*
 Delete an environment by id.
 */
func DeleteEnvironment(client SkytapClient, envId string) error {
	log.WithFields(log.Fields{"envId": envId}).Info("Deleting environment")

	deleteEnv := func(s *sling.Sling) *sling.Sling {
		return s.Delete(EnvironmentPath + "/" + envId)
	}

	_, err := RunSkytapRequest(client, false, nil, deleteEnv)
	return err
}
