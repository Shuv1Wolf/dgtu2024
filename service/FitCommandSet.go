package service

import (
	"context"
	"hack/data"

	cconv "github.com/pip-services4/pip-services4-go/pip-services4-commons-go/convert"
	exec "github.com/pip-services4/pip-services4-go/pip-services4-components-go/exec"
	cvalid "github.com/pip-services4/pip-services4-go/pip-services4-data-go/validate"
	ccmd "github.com/pip-services4/pip-services4-go/pip-services4-rpc-go/commands"
)

type FitCommandSet struct {
	ccmd.CommandSet
	service      IFitService
	fitConvertor cconv.IJSONEngine[data.FitV1]
}

func NewFitCommandSet(service IFitService) *FitCommandSet {
	c := &FitCommandSet{
		CommandSet:   *ccmd.NewCommandSet(),
		service:      service,
		fitConvertor: cconv.NewDefaultCustomTypeJsonConvertor[data.FitV1](),
	}

	c.AddCommand(c.makeGoogleAuthorizationCommand())
	return c
}

func (c *FitCommandSet) makeGoogleAuthorizationCommand() ccmd.ICommand {
	return ccmd.NewCommand(
		"google_authorization",
		cvalid.NewObjectSchema().
			WithRequiredProperty("mail", cconv.String),
		func(ctx context.Context, args *exec.Parameters) (result any, err error) {
			return c.service.GoogleAuthorization(ctx, args.GetAsString("mail"))
		})
}
