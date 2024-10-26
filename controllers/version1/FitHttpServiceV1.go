package controller

import (
	"context"

	cref "github.com/pip-services4/pip-services4-go/pip-services4-components-go/refer"
	cservices "github.com/pip-services4/pip-services4-go/pip-services4-http-go/controllers"
)

type FitHttpControllerV1 struct {
	cservices.CommandableHttpController
}

func NewFitHttpControllerV1() *FitHttpControllerV1 {
	c := &FitHttpControllerV1{}
	c.CommandableHttpController = *cservices.InheritCommandableHttpController(c, "v1/fit")
	c.DependencyResolver.Put(context.Background(), "service", cref.NewDescriptor("fit", "service", "*", "*", "1.0"))
	return c
}
