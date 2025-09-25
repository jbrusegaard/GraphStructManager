package driver

import (
	"testing"

	"app/comparator"
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

}
