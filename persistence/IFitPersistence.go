package persistence

import (
	"context"
	"hack/data"

	cquery "github.com/pip-services4/pip-services4-go/pip-services4-data-go/query"
)

type IFitPersistence interface {
	GetPage(ctx context.Context) (cquery.DataPage[data.FitV1], error)

	GetOneById(ctx context.Context, id string) (data.FitV1, error)

	Create(ctx context.Context, item data.FitV1) (data.FitV1, error)

	Update(ctx context.Context, item data.FitV1) (data.FitV1, error)

	DeleteById(ctx context.Context, id string) (data.FitV1, error)
}
