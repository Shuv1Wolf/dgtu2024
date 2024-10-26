package persistence

import (
	"context"
	"hack/data"

	cquery "github.com/pip-services4/pip-services4-go/pip-services4-data-go/query"
	cpersist "github.com/pip-services4/pip-services4-go/pip-services4-persistence-go/persistence"
)

type FitMemoryPersistence struct {
	cpersist.IdentifiableMemoryPersistence[data.FitV1, string]
}

func NewFitMemoryPersistence() *FitMemoryPersistence {
	c := &FitMemoryPersistence{
		IdentifiableMemoryPersistence: *cpersist.NewIdentifiableMemoryPersistence[data.FitV1, string](),
	}
	c.IdentifiableMemoryPersistence.MaxPageSize = 1000
	return c
}

func (c *FitMemoryPersistence) GetPage(ctx context.Context) (cquery.DataPage[data.FitV1], error) {
	paging := cquery.NewEmptyPagingParams()

	return c.IdentifiableMemoryPersistence.GetPageByFilter(ctx, nil, *paging, nil, nil)
}

func ContainsStr(arr []string, substr string) bool {
	for _, _str := range arr {
		if _str == substr {
			return true
		}
	}
	return false
}
