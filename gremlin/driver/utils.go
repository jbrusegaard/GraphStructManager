package driver

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"time"

	"app/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

var __ = gremlingo.T__
var P = gremlingo.P
var Order = gremlingo.Order
var Scope = gremlingo.Scope

type GremlinOrder int

const (
	Asc GremlinOrder = iota
	Desc
)

type VertexType interface {
	GetVertexId() any
	GetVertexLastModified() time.Time
	GetVertexCreatedAt() time.Time
}

// getStructName takes a generic type T, confirms it's a struct, and returns its name
func getStructName[T any]() (string, error) {
	var s T
	t := reflect.TypeOf(s)
	// Check if T is a struct type
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("type %s is not a struct, it's a %s", t.Name(), t.Kind())
	}
	return t.Name(), nil
}

func unloadGremlinResultIntoStruct(v any, result *gremlingo.Result) error {
	mapResult, ok := result.GetInterface().(map[any]any)
	if !ok {
		return errors.New("result is not a map")
	}
	// make string map
	stringMap := make(map[string]any)
	for key, value := range mapResult {
		stringMap[key.(string)] = value
	}
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return errors.New("v must be a pointer")
	}
	recursivelyUnloadIntoStruct(v, stringMap)
	return nil
}

func recursivelyUnloadIntoStruct(v any, stringMap map[string]any) {
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()

	for i := range rv.NumField() {
		field := rv.Field(i)
		fieldType := rt.Field(i)
		// handle anonymous Vertex field
		if fieldType.Anonymous {
			recursivelyUnloadIntoStruct(field.Addr().Interface(), stringMap)
		}

		gremlinTag := rt.Field(i).Tag.Get("gremlin")
		if gremlinTag == "" || gremlinTag == "-" || !field.CanInterface() || !field.CanSet() {
			continue
		}
		if _, ok := stringMap[gremlinTag]; !ok {
			continue
		}
		gType := reflect.TypeOf(stringMap[gremlinTag])

		if gType.ConvertibleTo(field.Type()) {
			field.Set(reflect.ValueOf(stringMap[gremlinTag]).Convert(field.Type()))
		} else if gType.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(
				field.Type(), len(stringMap[gremlinTag].([]any)), len(stringMap[gremlinTag].([]any)),
			)
			for i, v := range stringMap[gremlinTag].([]any) {
				slice.Index(i).Set(reflect.ValueOf(v).Convert(field.Type().Elem()))
			}
			field.Set(slice)
		}
	}
}

func structToMap(value any) (string, map[string]any, error) {
	mapValue := make(map[string]any)

	// Get the reflection value
	rv := reflect.ValueOf(value)

	// Check if it's a pointer and get the underlying value
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return "", nil, errors.New("value is not a struct")
	}

	// Get the type information
	rt := rv.Type()

	// Loop through all fields
	for i := range rv.NumField() {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			// Recursively process the anonymous struct
			_, anonymousMap, err := structToMap(fieldValue.Interface())
			if err != nil {
				return "", nil, fmt.Errorf(
					"error processing anonymous field %s: %w",
					field.Name,
					err,
				)
			}
			maps.Copy(mapValue, anonymousMap)
			continue
		}

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

	return rv.Type().Name(), mapValue, nil
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

func getStructFieldNameAndType[T any](tag string) (string, reflect.Type, error) {
	var s T
	rt := reflect.TypeOf(s)
	for i := range rt.NumField() {
		field := rt.Field(i)
		if field.Tag.Get("gremlin") == tag {
			return field.Name, field.Type, nil
		}
	}
	return "", nil, fmt.Errorf("field not found")
}
