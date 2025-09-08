package driver

import (
	"errors"
	"reflect"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func (driver *GremlinDriver) First(v any, id any) error {
	err := validateStructPointerWithAnonymousVertex(v)
	if err != nil {
		return err
	}

	reflect.ValueOf(v).Elem().FieldByName("Id").Set(reflect.ValueOf(id))

	result, err := driver.g.V(id).ElementMap().Next()
	if err != nil {
		return err
	}

	err = unloadResultIntoStruct(v, result)
	if err != nil {
		return err
	}
	return nil
}

func unloadResultIntoStruct(v any, result *gremlingo.Result) error {
	mapResult := result.GetInterface().(map[any]any)

	// make string map
	stringMap := make(map[string]any)
	for key, value := range mapResult {
		stringMap[key.(string)] = value
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return errors.New("v must be a pointer")
	}
	rv = rv.Elem()

	rt := rv.Type()

	for i := range rv.NumField() {
		field := rv.Field(i)

		gremlinTag := rt.Field(i).Tag.Get("gremlin")
		if gremlinTag == "" || gremlinTag == "-" || !field.CanInterface() || !field.CanSet() {
			continue
		}
		if _, ok := stringMap[gremlinTag]; !ok {
			continue
		}
		gType := reflect.TypeOf(stringMap[gremlinTag])

		// check if field is of type gType
		if field.Type() != gType {
			continue
		}

		field.Set(reflect.ValueOf(stringMap[gremlinTag]))
	}
	return nil
}
