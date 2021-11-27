package service

import (
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
)

// ServiceProviderBackend is implemented by the backends supported by odo
// It is used in "odo service create" and "odo service delete"
type ServiceProviderBackend interface {
	CompleteServiceCreate(options *CreateOptions, args []string) error
	ValidateServiceCreate(options *CreateOptions) error
	RunServiceCreate(options *CreateOptions) error

	ServiceDefined(context *genericclioptions.Context, name string) (bool, error)
	DeleteService(options *DeleteOptions, serviceName, app string) error
	DescribeService(options *DescribeOptions, serviceName, app string) error
}
