package driver

import (
	"testing"

	"app/comparator"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func seedData(db *GremlinDriver, data []testVertexForUtils) error {
	for _, d := range data {
		err := Create(db, &d)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanDB() {
	db, _ := Open(DbUrl)
	<-db.g.V().Drop().Iterate()
}

func TestQuery(t *testing.T) {
	db, err := Open(DbUrl)
	if err != nil {
		t.Fatal(err)
	}
	seededData := []testVertexForUtils{
		{
			Name: "first",
			Sort: 1,
		},
		{
			Name: "second",
			Sort: 2,
		},
		{
			Name: "third",
			Sort: 3,
		},
	}

	t.Run(
		"TestFindWhereFirst", func(t *testing.T) {
			t.Cleanup(cleanDB)
			err = seedData(db, seededData)
			if err != nil {
				t.Error(err)
			}
			model := Model[testVertexForUtils](db)
			results, err := model.Where("name", comparator.EQ, "first").Find()
			if err != nil {
				t.Error(err)
			}
			if len(results) != 1 {
				t.Errorf("Expected 1 result, got %d", len(results))
			}
			if results[0].Name != "first" {
				t.Errorf("Expected first result, got %s", results[0].Name)
			}
		},
	)
	orderTests := []struct {
		Name  string
		Order GremlinOrder
	}{
		{Name: "TestFindNoWhereOrderAsc", Order: Asc},
		{Name: "TestFindWhereOrderDesc", Order: Desc},
	}
	for _, orderTest := range orderTests {
		t.Run(
			orderTest.Name, func(t *testing.T) {
				t.Cleanup(cleanDB)
				err = seedData(db, seededData)
				if err != nil {
					t.Error(err)
				}
				model := Model[testVertexForUtils](db)
				results, err := model.OrderBy("sort", orderTest.Order).Find()
				if err != nil {
					t.Error(err)
				}
				if len(results) != len(seededData) {
					t.Errorf("Expected %d results, got %d", len(seededData), len(results))
				}
				for i, item := range results {
					var idx int
					switch orderTest.Order {
					case Asc:
						idx = i
					case Desc:
						idx = len(results) - i - 1
					}
					if item.Name != seededData[idx].Name {
						t.Errorf("Expected %s result, got %s", seededData[i].Name, item.Name)
					}
				}
			},
		)
	}
	t.Run(
		"TestQueryWhereTraversal", func(t *testing.T) {
			t.Cleanup(cleanDB)
			err = seedData(db, seededData)
			if err != nil {
				t.Error(err)
			}
			model := Model[testVertexForUtils](
				db,
			).WhereTraversal(gremlingo.T__.Has("name", "second"))
			result, err := model.Take()
			if err != nil {
				t.Error(err)
			}
			if result.Name != "second" {
				t.Errorf("Expected second result, got %s", result.Name)
			}
		},
	)

	t.Run(
		"TestDelete", func(t *testing.T) {
			t.Cleanup(cleanDB)
			err = seedData(db, seededData)
			if err != nil {
				t.Error(err)
			}
			err := Model[testVertexForUtils](db).Limit(1).Delete()
			if err != nil {
				t.Error(err)
			}
			count, err := Model[testVertexForUtils](db).Count()
			if err != nil {
				t.Error(err)
			}
			if count != len(seededData)-1 {
				t.Errorf("Expected %d results, got %d", len(seededData)-1, count)
			}
		},
	)

	t.Run(
		"TestQueryById", func(t *testing.T) {
			t.Cleanup(cleanDB)
			err = seedData(db, seededData)
			if err != nil {
				t.Error(err)
			}
			model, err := Model[testVertexForUtils](db).Take()
			if err != nil {
				t.Error(err)
			}
			result, err := Model[testVertexForUtils](db).Id(model.Id)
			if err != nil {
				t.Error(err)
			}
			if result.Name != model.Name {
				t.Errorf("Expected %s result, got %s", model.Name, result.Name)
			}
			if result.Id != model.Id {
				t.Errorf("Expected %s result, got %s", model.Id, result.Id)
			}
			if result.Sort != model.Sort {
				t.Errorf("Expected %b result, got %b", model.Sort, result.Sort)
			}
		},
	)

	t.Run(
		"TestQueryUpdateBadInput", func(t *testing.T) {
			t.Cleanup(cleanDB)
			err = seedData(db, seededData)
			if err != nil {
				t.Error(err)
			}
			err = Model[testVertexForUtils](db).Update("badField", "badValue")
			if err == nil {
				t.Error("Expected error")
			}
		},
	)
	t.Run(
		"TestQueryUpdateSingleCardinality", func(t *testing.T) {
			t.Cleanup(cleanDB)
			err = seedData(db, seededData)
			if err != nil {
				t.Error(err)
			}
			preUpdateModel, err := Model[testVertexForUtils](
				db,
			).Where("name", comparator.EQ, "first").
				Take()
			if err != nil {
				t.Error(err)
			}
			err = Model[testVertexForUtils](
				db,
			).Where("name", comparator.EQ, "first").
				Update("name", "fourth")
			if err != nil {
				t.Error(err)
			}
			model, err := Model[testVertexForUtils](
				db,
			).Where("name", comparator.EQ, "fourth").
				Take()
			if err != nil {
				t.Error(err)
			}
			if model.Name != "fourth" {
				t.Errorf("Expected %s result, got %s", "fourth", model.Name)
			}
			if preUpdateModel.LastModified.Equal(model.LastModified) {
				t.Error("Expected last modified time to be updated")
			}
			if preUpdateModel.LastModified.Equal(model.LastModified) {
				t.Error("Expected last modified time to be updated")
			}
		},
	)
}
