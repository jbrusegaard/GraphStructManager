package driver

import (
	"encoding/json"
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
	resultJson, err := json.Marshal(stringMap)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resultJson, v)
	if err != nil {
		return err
	}
	return nil
}
