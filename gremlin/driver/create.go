package driver

import (
	"reflect"
	"time"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func Create[T VertexType](db *GremlinDriver, value *T) error {
	return createOrUpdate(db, value)
}

func Update[T VertexType](db *GremlinDriver, value *T) error {
	return createOrUpdate(db, value)
}

func createOrUpdate[T VertexType](db *GremlinDriver, value *T) error {
	err := validateStructPointerWithAnonymousVertex(value)
	if err != nil {
		db.logger.Errorf("Validation failed: %v", err)
		return err
	}
	now := time.Now().UTC()
	label, mapValue, err := structToMap(value)
	if err != nil {
		return err
	}
	id := mapValue["id"]
	delete(mapValue, "id")
	mapValue["lastModified"] = now
	var query *gremlingo.GraphTraversal
	newMap := make(map[any]any, len(mapValue))
	if id == nil {
		mapValue["createdAt"] = now
		for k, v := range mapValue {
			newMap[k] = v
		}
		newMap[gremlingo.T.Label] = label
		query = db.g.MergeV(newMap)
	} else {
		query = db.g.MergeV(map[any]any{gremlingo.T.Id: id})
	}
	query.Option(gremlingo.Merge.OnMatch, mapValue)
	vertexID, err := query.Id().Next()
	if err != nil {
		return err
	}
	reflect.ValueOf(value).Elem().FieldByName("ID").Set(reflect.ValueOf(vertexID.GetInterface()))
	reflectNow := reflect.ValueOf(now)
	reflect.ValueOf(value).Elem().FieldByName("LastModified").Set(reflectNow)
	reflect.ValueOf(value).Elem().FieldByName("CreatedAt").Set(reflectNow)
	return nil
}
