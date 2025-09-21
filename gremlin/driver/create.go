package driver

import (
	"reflect"
	"time"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func Create[T VertexType](db *GremlinDriver, value *T) error {
	// Validate it's a struct pointer with anonymous Vertex
	err := validateStructPointerWithAnonymousVertex(value)
	if err != nil {
		db.logger.Errorf("Validation failed: %v", err)
		return err
	}

	now := time.Now().Unix()

	structName, mapValue := structToMap(value)
	mapValue["lastModified"] = now

	query := db.g.AddV(structName)
	for key, value := range mapValue {
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Map:
			query.Property(gremlingo.Cardinality.Set, key, value)
		default:
			query.Property(gremlingo.Cardinality.Single, key, value)
		}
	}
	id, err := query.Id().Next()
	if err != nil {
		return err
	}
	reflect.ValueOf(value).Elem().FieldByName("Id").Set(reflect.ValueOf(id.GetInterface()))
	reflect.ValueOf(value).Elem().FieldByName("LastModified").SetInt(now)
	return nil
}
