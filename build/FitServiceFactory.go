package build

import (
	controller "hack/controllers/version1"
	"hack/persistence"
	"hack/service"

	cbuild "github.com/pip-services4/pip-services4-go/pip-services4-components-go/build"
	cref "github.com/pip-services4/pip-services4-go/pip-services4-components-go/refer"
)

type FitServiceFactory struct {
	cbuild.Factory
}

func NewFitServiceFactory() *FitServiceFactory {
	c := &FitServiceFactory{
		Factory: *cbuild.NewFactory(),
	}

	memoryPersistenceDescriptor := cref.NewDescriptor("fit", "persistence", "memory", "*", "1.0")
	serviceDescriptor := cref.NewDescriptor("fit", "service", "default", "*", "1.0")
	httpcontrollerV1Descriptor := cref.NewDescriptor("fit", "controller", "http", "*", "1.0")

	c.RegisterType(memoryPersistenceDescriptor, persistence.NewFitMemoryPersistence)
	c.RegisterType(serviceDescriptor, service.NewFitService)
	c.RegisterType(httpcontrollerV1Descriptor, controller.NewFitHttpControllerV1)

	return c
}
