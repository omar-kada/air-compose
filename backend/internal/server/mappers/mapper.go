package mappers

import (
	"omar-kada/air-compose/api"
)

// Mapper is a generic interface for mapping between types T and V.
// It provides methods to convert single instances and slices of types.
type Mapper[T any, V any] interface {
	// Map converts a single source type T to a target type V.
	Map(T) V
}

// MapToPageInfo maps a slice of T to an api.PageInfo, determining if there are more items
// and providing the end cursor for pagination.
func MapToPageInfo[T any](objs []T, limit int, getEndCursor func(obj T) string) api.PageInfo {
	endCursor := ""
	if len(objs) > 0 {
		last := objs[len(objs)-1]
		endCursor = getEndCursor(last)
	}
	return api.PageInfo{
		HasNextPage: len(objs) == limit,
		EndCursor:   endCursor,
	}
}
