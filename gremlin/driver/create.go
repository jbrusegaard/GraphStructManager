package driver

import (
	"reflect"
	"time"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/jbrusegaard/graph-struct-manager/gsmtypes"
)

func Create[T gsmtypes.VertexType](db *GremlinDriver, value *T) error {
	return createOrUpdate(db, value)
}

func Update[T gsmtypes.VertexType](db *GremlinDriver, value *T) error {
	return createOrUpdate(db, value)
}

func createOrUpdate[T gsmtypes.VertexType](db *GremlinDriver, value *T) error {
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
	mapValue[gsmtypes.LastModified] = now
	var query *gremlingo.GraphTraversal

	if id == nil {
		mapValue[gsmtypes.CreatedAt] = now
		query = db.g.AddV(label)
		for k, v := range mapValue {
			if reflect.ValueOf(v).Kind() == reflect.Slice {
				sliceValue := reflect.ValueOf(v)
				if db.dbDriver == Neptune {
					for i := range sliceValue.Len() {
						query.Property(gremlingo.Cardinality.Set, k, sliceValue.Index(i))
					}
				} else {
					for i := range sliceValue.Len() {
						query.Property(gremlingo.Cardinality.List, k, sliceValue.Index(i))
					}
				}
			} else {
				query.Property(gremlingo.Cardinality.Single, k, v)
			}
		}
		result, err := query.Next()
		if err != nil {
			return err
		}
		resultElement, err := result.GetElement()
		if err != nil {
			return err
		}
		reflect.ValueOf(value).Elem().FieldByName("ID").Set(reflect.ValueOf(resultElement.Id))
	} else {
		query = db.g.MergeV(map[any]any{gremlingo.T.Id: id})
		query.Option(gremlingo.Merge.OnMatch, mapValue)
		id, err := query.Id().Next()
		if err != nil {
			return err
		}
		reflect.ValueOf(value).Elem().FieldByName("ID").Set(reflect.ValueOf(id.GetInterface()))
	}
	reflectNow := reflect.ValueOf(now)
	reflect.ValueOf(value).Elem().FieldByName("LastModified").Set(reflectNow)
	reflect.ValueOf(value).Elem().FieldByName("CreatedAt").Set(reflectNow)
	return nil
}
