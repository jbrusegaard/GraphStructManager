package driver

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"time"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/gobeam/stringy"
	"github.com/jbrusegaard/graph-struct-manager/gsmtypes"
)

var (
	anonymousTraversal = gremlingo.T__
	P                  = gremlingo.P
	Order              = gremlingo.Order
	Scope              = gremlingo.Scope
)

type GremlinOrder int

const (
	Asc GremlinOrder = iota
	Desc
)

type VertexType interface {
	GetVertexID() any
	GetVertexLastModified() time.Time
	GetVertexCreatedAt() time.Time
	Label() string
}

type EdgeType interface {
	GetEdgeID() any
	GetEdgeLastModified() string
	GetEdgeCreatedAt() int64
	Label() string
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
		keyStr, keyOk := key.(string)
		if !keyOk {
			return errors.New("gremlin key is not a string")
		}
		stringMap[keyStr] = value
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
			strSlice := stringMap[gremlinTag].([]any) //nolint:errcheck // we already validated via reflect type check
			slice := reflect.MakeSlice(
				field.Type(), len(strSlice), len(strSlice),
			)
			for i, v := range strSlice {
				slice.Index(i).Set(reflect.ValueOf(v).Convert(field.Type().Elem()))
			}
			field.Set(slice)
		}
	}
}

// getLabelFromValue gets the label from a value, using the Label() method if it returns a non-empty string,
// otherwise falling back to struct name normalization
// Supports both pointer and value receivers for the Label() method
func getLabelFromValue( //nolint:gocognit // this is a complex function but it's necessary to support both pointer and value receivers
	value any,
) (string, error) {
	rv := reflect.ValueOf(value)
	originalRv := rv

	// Get the underlying struct type
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return "", errors.New("value is a nil pointer")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return "", errors.New("value is not a struct")
	}
	rt := rv.Type()

	// Try to call Label() method using reflection to support both pointer and value receivers
	var labelMethod reflect.Value

	// First, try on the original value (which might be a pointer)
	if originalRv.Kind() == reflect.Ptr {
		labelMethod = originalRv.MethodByName("Label")
	}

	// If not found on pointer, try on the value itself (for value receivers)
	if !labelMethod.IsValid() {
		labelMethod = rv.MethodByName("Label")
	}

	// If still not found and we have a value (not pointer), try getting a pointer to it
	// This handles the case where Label() has a pointer receiver but we received a value
	if !labelMethod.IsValid() && originalRv.Kind() != reflect.Ptr {
		// Check if the value is addressable (can take its address)
		if rv.CanAddr() {
			labelMethod = rv.Addr().MethodByName("Label")
		} else {
			// If not addressable, create a new pointer with the value copied
			ptrRv := reflect.New(rt)
			ptrRv.Elem().Set(rv)
			labelMethod = ptrRv.MethodByName("Label")
		}
	}

	if labelMethod.IsValid() {
		// Call the Label() method
		results := labelMethod.Call(nil)
		if len(results) > 0 {
			label := results[0].String()
			// If Label() returns empty string, use struct name normalization
			if label == "" {
				return stringy.New(rt.Name()).SnakeCase().ToLower(), nil
			}
			return label, nil
		}
	}

	// Fallback: try interface assertion (for value receivers)
	if vertexType, ok := value.(VertexType); ok {
		label := vertexType.Label()
		if label == "" {
			return stringy.New(rt.Name()).SnakeCase().ToLower(), nil
		}
		return label, nil
	}
	if edgeType, ok := value.(EdgeType); ok {
		label := edgeType.Label()
		if label == "" {
			return stringy.New(rt.Name()).SnakeCase().ToLower(), nil
		}
		return label, nil
	}

	// Fallback to struct name normalization
	return stringy.New(rt.Name()).SnakeCase().ToLower(), nil
}

// structToMap converts a struct to a map[string]any and returns the label and the map
// the label is determined by calling Label() method if available, otherwise the name of the struct converted to snake case
// the map is the map of the struct
// the error is the error if any
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

	// Get the label using the helper function
	label, err := getLabelFromValue(value)
	if err != nil {
		return "", nil, err
	}

	// Get the type information
	rt := rv.Type()

	// Loop through all fields
	for i := range rv.NumField() {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			// Recursively process the anonymous struct
			_, anonymousMap, structMapErr := structToMap(fieldValue.Interface())
			if structMapErr != nil {
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

	return label, mapValue, nil
}

func validateStructPointerWithAnonymousVertex(value any) error {
	rv := reflect.ValueOf(value)

	// Check if it's a pointer
	if rv.Kind() != reflect.Ptr {
		return errors.New("value must be a pointer")
	}

	// Check if it's a nil pointer
	if rv.IsNil() {
		return errors.New("value cannot be nil")
	}

	// Check if it points to a struct
	if rv.Elem().Kind() != reflect.Struct {
		return errors.New("value must point to a struct")
	}

	// Get the struct type
	rt := rv.Elem().Type()

	// Check for anonymous Vertex field
	for i := range rv.Elem().NumField() {
		field := rt.Field(i)

		if field.Anonymous && field.Type == reflect.TypeFor[gsmtypes.Vertex]() {
			return nil
		}
	}

	return errors.New("struct must contain anonymous types.Vertex field")
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
	return "", nil, errors.New("field not found")
}
