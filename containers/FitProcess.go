package containers

import (
	"hack/build"

	cproc "github.com/pip-services4/pip-services4-go/pip-services4-container-go/container"
	rbuild "github.com/pip-services4/pip-services4-go/pip-services4-http-go/build"
)

type FitProcess struct {
	cproc.ProcessContainer
}

func NewFitProcess() *FitProcess {
	c := &FitProcess{
		ProcessContainer: *cproc.NewProcessContainer("fit", "Fit microservice"),
	}

	c.AddFactory(build.NewFitServiceFactory())
	c.AddFactory(rbuild.NewDefaultHttpFactory())

	return c
}
