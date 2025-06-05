package flag

import (
	flags "github.com/jessevdk/go-flags"
)

type ResourceType string

const (
	AppResourceType             ResourceType = "app"
	BuildpackResourceType       ResourceType = "buildpack"
	DomainResourceType          ResourceType = "domain"
	OrgResourceType             ResourceType = "org"
	RouteResourceType           ResourceType = "route"
	SpaceResourceType           ResourceType = "space"
	StackResourceType           ResourceType = "stack"
	ServiceBrokerResourceType   ResourceType = "service-broker"
	ServiceInstanceResourceType ResourceType = "service-instance"
	ServiceOfferingResourceType ResourceType = "service-offering"
	ServicePlanResourceType     ResourceType = "service-plan"
)

var allResourceTypes = []ResourceType{
	AppResourceType,
	BuildpackResourceType,
	DomainResourceType,
	OrgResourceType,
	RouteResourceType,
	SpaceResourceType,
	StackResourceType,
	ServiceBrokerResourceType,
	ServiceInstanceResourceType,
	ServiceOfferingResourceType,
	ServicePlanResourceType,
}

func GetAllResourceTypes() []string {
	var resourceTypes []string
	for _, resourceType := range allResourceTypes {
		resourceTypes = append(resourceTypes, string(resourceType))
	}
	return resourceTypes
}

func (ResourceType) Complete(prefix string) []flags.Completion {
	return completions(GetAllResourceTypes(), prefix, false)
}
