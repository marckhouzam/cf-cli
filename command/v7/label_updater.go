package v7

import (
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/types"
	"code.cloudfoundry.org/cli/v8/util/configv3"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SetLabelActor

type SetLabelActor interface {
	GetCurrentUser() (configv3.User, error)
	UpdateApplicationLabelsByApplicationName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateBuildpackLabelsByBuildpackNameAndStackAndLifecycle(string, string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateDomainLabelsByDomainName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateOrganizationLabelsByOrganizationName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateRouteLabels(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateSpaceLabelsBySpaceName(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateStackLabelsByStackName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateServiceInstanceLabels(string, string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateServiceBrokerLabelsByServiceBrokerName(string, map[string]types.NullString) (v7action.Warnings, error)
	UpdateServiceOfferingLabels(serviceOfferingName string, serviceBrokerName string, labels map[string]types.NullString) (v7action.Warnings, error)
	UpdateServicePlanLabels(servicePlanName string, serviceOfferingName string, serviceBrokerName string, labels map[string]types.NullString) (v7action.Warnings, error)
}

type ActionType string

const (
	Unset ActionType = "Removing"
	Set   ActionType = "Setting"
)

type TargetResource struct {
	ResourceType       string
	ResourceName       string
	BuildpackStack     string
	BuildpackLifecycle string
	ServiceBroker      string
	ServiceOffering    string
}

type LabelUpdater struct {
	targetResource TargetResource
	labels         map[string]types.NullString

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SetLabelActor

	Username string
	Action   ActionType
}

func (cmd *LabelUpdater) Execute(targetResource TargetResource, labels map[string]types.NullString) error {
	cmd.targetResource = targetResource
	cmd.labels = labels
	cmd.targetResource.ResourceType = strings.ToLower(cmd.targetResource.ResourceType)

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.Username = user.Name

	if err := cmd.validateFlags(); err != nil {
		return err
	}

	if err := cmd.checkTarget(); err != nil {
		return err
	}

	var warnings v7action.Warnings
	switch flag.ResourceType(cmd.targetResource.ResourceType) {
	case flag.AppResourceType:
		cmd.displayMessageWithOrgAndSpace()
		warnings, err = cmd.Actor.UpdateApplicationLabelsByApplicationName(cmd.targetResource.ResourceName, cmd.Config.TargetedSpace().GUID, cmd.labels)
	case flag.BuildpackResourceType:
		cmd.displayMessageWithStackAndLifecycle()
		warnings, err = cmd.Actor.UpdateBuildpackLabelsByBuildpackNameAndStackAndLifecycle(cmd.targetResource.ResourceName, cmd.targetResource.BuildpackStack, cmd.targetResource.BuildpackLifecycle, cmd.labels)
	case flag.DomainResourceType:
		cmd.displayMessageDefault()
		warnings, err = cmd.Actor.UpdateDomainLabelsByDomainName(cmd.targetResource.ResourceName, cmd.labels)
	case flag.OrgResourceType:
		cmd.displayMessageDefault()
		warnings, err = cmd.Actor.UpdateOrganizationLabelsByOrganizationName(cmd.targetResource.ResourceName, cmd.labels)
	case flag.RouteResourceType:
		cmd.displayMessageWithOrgAndSpace()
		warnings, err = cmd.Actor.UpdateRouteLabels(cmd.targetResource.ResourceName, cmd.Config.TargetedSpace().GUID, cmd.labels)
	case flag.ServiceBrokerResourceType:
		cmd.displayMessageDefault()
		warnings, err = cmd.Actor.UpdateServiceBrokerLabelsByServiceBrokerName(cmd.targetResource.ResourceName, cmd.labels)
	case flag.ServiceInstanceResourceType:
		cmd.displayMessageWithOrgAndSpace()
		warnings, err = cmd.Actor.UpdateServiceInstanceLabels(cmd.targetResource.ResourceName, cmd.Config.TargetedSpace().GUID, cmd.labels)
	case flag.ServiceOfferingResourceType:
		cmd.displayMessageForServiceCommands()
		warnings, err = cmd.Actor.UpdateServiceOfferingLabels(cmd.targetResource.ResourceName, cmd.targetResource.ServiceBroker, cmd.labels)
	case flag.ServicePlanResourceType:
		cmd.displayMessageForServiceCommands()
		warnings, err = cmd.Actor.UpdateServicePlanLabels(cmd.targetResource.ResourceName, cmd.targetResource.ServiceOffering, cmd.targetResource.ServiceBroker, cmd.labels)
	case flag.SpaceResourceType:
		cmd.displayMessageWithOrg()
		warnings, err = cmd.Actor.UpdateSpaceLabelsBySpaceName(cmd.targetResource.ResourceName, cmd.Config.TargetedOrganization().GUID, cmd.labels)
	case flag.StackResourceType:
		cmd.displayMessageDefault()
		warnings, err = cmd.Actor.UpdateStackLabelsByStackName(cmd.targetResource.ResourceName, cmd.labels)
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd *LabelUpdater) checkTarget() error {
	switch flag.ResourceType(cmd.targetResource.ResourceType) {
	case flag.AppResourceType, flag.ServiceInstanceResourceType, flag.RouteResourceType:
		return cmd.SharedActor.CheckTarget(true, true)
	case flag.SpaceResourceType:
		return cmd.SharedActor.CheckTarget(true, false)
	default:
		return cmd.SharedActor.CheckTarget(false, false)
	}
}

func (cmd *LabelUpdater) validateFlags() error {
	resourceType := flag.ResourceType(cmd.targetResource.ResourceType)
	switch resourceType {
	case flag.AppResourceType, flag.BuildpackResourceType, flag.DomainResourceType, flag.OrgResourceType, flag.RouteResourceType, flag.ServiceBrokerResourceType, flag.ServiceInstanceResourceType, flag.ServiceOfferingResourceType, flag.ServicePlanResourceType, flag.SpaceResourceType, flag.StackResourceType:
	default:
		return errors.New(cmd.UI.TranslateText("Unsupported resource type of '{{.ResourceType}}'", map[string]interface{}{"ResourceType": cmd.targetResource.ResourceType}))
	}

	if cmd.targetResource.BuildpackStack != "" && resourceType != flag.BuildpackResourceType {
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				cmd.targetResource.ResourceType, "--stack, -s",
			},
		}
	}

	if cmd.targetResource.ServiceBroker != "" && !(resourceType == flag.ServiceOfferingResourceType || resourceType == flag.ServicePlanResourceType) {
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				cmd.targetResource.ResourceType, "--broker, -b",
			},
		}
	}

	if cmd.targetResource.ServiceOffering != "" && resourceType != flag.ServicePlanResourceType {
		return translatableerror.ArgumentCombinationError{
			Args: []string{
				cmd.targetResource.ResourceType, "--offering, -o",
			},
		}
	}

	return nil
}

func actionForResourceString(action string, resourceType string) string {
	return fmt.Sprintf("%s label(s) for %s", action, resourceType)
}

func (cmd *LabelUpdater) displayMessageDefault() {
	cmd.UI.DisplayTextWithFlavor(actionForResourceString(string(cmd.Action), cmd.targetResource.ResourceType)+" {{.ResourceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.targetResource.ResourceName,
		"User":         cmd.Username,
	})
}

func (cmd *LabelUpdater) displayMessageWithOrgAndSpace() {
	cmd.UI.DisplayTextWithFlavor(actionForResourceString(string(cmd.Action), cmd.targetResource.ResourceType)+" {{.ResourceName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.targetResource.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"SpaceName":    cmd.Config.TargetedSpace().Name,
		"User":         cmd.Username,
	})
}

func (cmd *LabelUpdater) displayMessageWithStackAndLifecycle() {
	template := actionForResourceString(string(cmd.Action), cmd.targetResource.ResourceType) + " {{.ResourceName}}"

	if cmd.targetResource.BuildpackStack != "" {
		template += " with stack {{.StackName}}"
	}

	if cmd.targetResource.BuildpackLifecycle != "" {
		template += " with lifecycle {{.Lifecycle}}"
	}

	template += " as {{.User}}..."

	cmd.UI.DisplayTextWithFlavor(template, map[string]interface{}{
		"ResourceName": cmd.targetResource.ResourceName,
		"StackName":    cmd.targetResource.BuildpackStack,
		"Lifecycle":    cmd.targetResource.BuildpackLifecycle,
		"User":         cmd.Username,
	})
}

func (cmd *LabelUpdater) displayMessageForServiceCommands() {
	template := actionForResourceString(string(cmd.Action), cmd.targetResource.ResourceType) + " {{.ResourceName}}"

	if cmd.targetResource.ServiceOffering != "" || cmd.targetResource.ServiceBroker != "" {
		template += " from"

		if cmd.targetResource.ServiceOffering != "" {
			template += " service offering {{.ServiceOffering}}"
			if cmd.targetResource.ServiceBroker != "" {
				template += " /"
			}
		}

		if cmd.targetResource.ServiceBroker != "" {
			template += " service broker {{.ServiceBroker}}"
		}
	}

	template += " as {{.User}}..."
	cmd.UI.DisplayTextWithFlavor(template, map[string]interface{}{
		"ResourceName":    cmd.targetResource.ResourceName,
		"ServiceBroker":   cmd.targetResource.ServiceBroker,
		"ServiceOffering": cmd.targetResource.ServiceOffering,
		"User":            cmd.Username,
	})
}

func (cmd *LabelUpdater) displayMessageWithOrg() {
	cmd.UI.DisplayTextWithFlavor(actionForResourceString(string(cmd.Action), cmd.targetResource.ResourceType)+" {{.ResourceName}} in org {{.OrgName}} as {{.User}}...", map[string]interface{}{
		"ResourceName": cmd.targetResource.ResourceName,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"User":         cmd.Username,
	})
}
