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

	now := time.Now().UTC()

	structName, mapValue, err := structToMap(value)
	if err != nil {
		return err
	}
	mapValue["lastModified"] = now
	mapValue["createdAt"] = now

	query := db.g.AddV(structName)
	for key, value := range mapValue {
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice:
			sliceLen := rv.Len()
			for i := range sliceLen {
				query.Property(gremlingo.Cardinality.Set, key, rv.Index(i).Interface())
			}
		case reflect.Map:
			if mapVal, ok := rv.Interface().(map[string]any); ok {
				for k := range mapVal {
					query.Property(gremlingo.Cardinality.Set, key, k)
				}
			}
		default:
			query.Property(gremlingo.Cardinality.Single, key, value)
		}
	}
	id, err := query.Id().Next()
	if err != nil {
		return err
	}
	reflect.ValueOf(value).Elem().FieldByName("Id").Set(reflect.ValueOf(id.GetInterface()))
	reflectNow := reflect.ValueOf(now)
	reflect.ValueOf(value).Elem().FieldByName("LastModified").Set(reflectNow)
	reflect.ValueOf(value).Elem().FieldByName("CreatedAt").Set(reflectNow)
	return nil
}
