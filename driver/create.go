package driver

import (
	"fmt"
	"reflect"
	"time"

	"app/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func Create[T any](db *GremlinDriver, value *T) error {
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

func structToMap(value any) (string, map[any]any) {
	mapValue := make(map[any]any)

	// Get the reflection value
	rv := reflect.ValueOf(value)

	// Check if it's a pointer and get the underlying value
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Get the type information
	rt := rv.Type()

	// Loop through all fields
	for i := range rv.NumField() {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Get the gremlin tag
		gremlinTag := field.Tag.Get("gremlin")

		// Skip if no gremlin tag or if field is not exported
		if gremlinTag == "" || gremlinTag == "-" || !fieldValue.CanInterface() {
			continue
		}

		// Get the field value
		fieldInterface := fieldValue.Interface()

		// Use the gremlin tag as the property name
		mapValue[gremlinTag] = fieldInterface
	}

	return rv.Type().Name(), mapValue
}

func validateStructPointerWithAnonymousVertex(value any) error {
	rv := reflect.ValueOf(value)

	// Check if it's a pointer
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("value must be a pointer")
	}

	// Check if it's a nil pointer
	if rv.IsNil() {
		return fmt.Errorf("value cannot be nil")
	}

	// Check if it points to a struct
	if rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("value must point to a struct")
	}

	// Get the struct type
	rt := rv.Elem().Type()

	// Check for anonymous Vertex field
	for i := 0; i < rv.Elem().NumField(); i++ {
		field := rt.Field(i)

		if field.Anonymous && field.Type == reflect.TypeOf(types.Vertex{}) {
			return nil
		}
	}

	return fmt.Errorf("struct must contain anonymous types.Vertex field")
}
