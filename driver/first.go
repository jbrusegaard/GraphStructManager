package driver

import (
	"errors"
	"reflect"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func First[T any](db *GremlinDriver, id any) (T, error) {
	var v T
	structName, err := getStructName[T]()
	if err != nil {
		return v, err
	}
	result, err := GremlinBaseQueryByIdOrTrav(db.g, id, structName).ElementMap().Next()
	if err != nil {
		return v, err
	}
	err = unloadGremlinResultIntoStruct(&v, result)
	return v, err
}

func GremlinBaseQueryByIdOrTrav(
	g *gremlingo.GraphTraversalSource,
	id any,
	label string,
) *gremlingo.GraphTraversal {
	switch id.(type) {
	case gremlingo.AnonymousTraversal, *gremlingo.GraphTraversal:
		return g.V().HasLabel(label).Where(id)
	default:
		return g.V(id)
	}
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
