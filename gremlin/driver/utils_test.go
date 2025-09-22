package driver

import (
	"testing"

	"app/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type testVertexForUtils struct {
	types.Vertex
	Name     string   `json:"name"     gremlin:"name"`
	Ignore   string   `json:"-"        gremlin:"-"`
	ListTest []string `json:"listTest" gremlin:"listTest"`
	Unmapped int      `json:"unmapped" gremlin:"unmapped"`
}

func TestUtils(t *testing.T) {
	t.Parallel()
	t.Run(
		"GetStructName", func(t *testing.T) {
			t.Parallel()
			name, err := getStructName[testVertexForUtils]()
			if err != nil {
				t.Errorf("Error getting struct name: %v", err)
			}
			if name != "testVertexForUtils" {
				t.Errorf("Struct name should be testVertexForUtils, got %s", name)
			}
		},
	)
	t.Run(
		"GetStructNameErr", func(t *testing.T) {
			t.Parallel()
			_, err := getStructName[int]()
			if err == nil {
				t.Errorf("No error getting struct name: %v", err)
			}
		},
	)
	t.Run(
		"TestUnloadGremlinResultIntoStruct", func(t *testing.T) {
			t.Parallel()
			var v testVertexForUtils
			err := unloadGremlinResultIntoStruct(
				&v, &gremlingo.Result{
					Data: map[any]any{
						"id":           "1",
						"lastModified": 1,
						"name":         "test",
						"listTest":     []string{"test1", "test2"},
					},
				},
			)
			if err != nil {
				t.Errorf("Error unloading gremlin result into struct: %v", err)
			}
			if v.Id != "1" {
				t.Errorf("Vertex ID should be 1, got %s", v.Id)
			}
			if v.LastModified != 1 {
				t.Errorf("Vertex LastModified should be 1, got %d", v.LastModified)
			}
			if v.Name != "test" {
				t.Errorf("Vertex Name should be test, got %s", v.Name)
			}
		},
	)
}
