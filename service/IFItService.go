package service

import (
	"context"
)

type IFitService interface {
	GoogleAuthorization(ctx context.Context, mail string) (string, error)
}
