package mappers

import (
	"omar-kada/air-compose/api"
)

// Mapper is a generic interface for mapping between types T and V.
type Mapper[T any, V any] interface {
	// Map converts a single source type T to a target type V.
	Map(T) V
}

// MapperUnmapper is a generic interface that combines mapping and unmapping capabilities.
type MapperUnmapper[T any, V any] interface {
	Mapper[T, V]
	UnMap(V) T
}

// PageMapper is a specialized Mapper interface that adds page information mapping capabilities.
type PageMapper[T any, V any] interface {
	Mapper[T, V]
	// Map converts a single source type T to a target type V.
	MapToPageInfo(arr []T, limit int) api.PageInfo
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
