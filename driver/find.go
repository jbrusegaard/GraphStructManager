package driver

import (
	"fmt"
	"reflect"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func Find[T any](db *GremlinDriver, whereClause *gremlingo.GraphTraversal) ([]T, error) {
	label, err := getStructName[T]()
	if err != nil {
		return nil, err
	}
	query := db.g.V().HasLabel(label)
	if whereClause != nil {
		query.Where(whereClause)
	}
	queryResults, err := query.ElementMap().ToList()
	if err != nil {
		return nil, err
	}
	results := make([]T, 0, len(queryResults))
	for _, result := range queryResults {
		var v T
		err = unloadGremlinResultIntoStruct(&v, result)
		if err != nil {
			return nil, err
		}
		results = append(results, v)
	}
	return results, nil
}

// getStructName takes a generic type T, confirms it's a struct, and returns its name
func getStructName[T any]() (string, error) {
	var zero T
	t := reflect.TypeOf(zero)
	// Handle pointer types by getting the underlying type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Check if T is a struct type
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("type %s is not a struct, it's a %s", t.Name(), t.Kind())
	}
	return t.Name(), nil
}
