package driver

import (
	"errors"
	"fmt"
	"reflect"

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
	GetVertexLastModified() int64
}

func toMapTraversal(query *gremlingo.GraphTraversal, args ...any) *gremlingo.GraphTraversal {
	return query.ValueMap(args...).By(
		__.Choose(
			__.Count(Scope.Local).Is(P.Eq(1)),
			__.Unfold(),
			__.Identity(),
		),
	)
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
	return recursivelyUnloadIntoStruct(v, stringMap)
}

func recursivelyUnloadIntoStruct(v any, stringMap map[string]any) error {
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()

	for i := range rv.NumField() {
		field := rv.Field(i)
		fieldType := rt.Field(i)
		// handle anonymous Vertex field
		if fieldType.Anonymous {
			err := recursivelyUnloadIntoStruct(field.Addr().Interface(), stringMap)
			if err != nil {
				return err
			}
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
