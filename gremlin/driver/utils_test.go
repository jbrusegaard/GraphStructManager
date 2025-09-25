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
	Sort     int      `json:"sort"     gremlin:"sort"`
}

type testVertexWithNumSlice struct {
	types.Vertex
	ListInts []int `json:"listInts" gremlin:"listInts"`
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
	unloadGremlinResultIntoStructTests := []struct {
		testName  string
		result    *gremlingo.Result
		v         any
		shouldErr bool
	}{
		{
			result:    &gremlingo.Result{},
			v:         &testVertexForUtils{},
			shouldErr: true,
			testName:  "UnloadGremlinResultIntoStructWithError",
		},
		{
			result: &gremlingo.Result{
				Data: map[any]any{
					"id":           "1",
					"lastModified": 1,
					"name":         "test",
					"listTest":     []string{"test1", "test2"},
				},
			},
			v:         &testVertexForUtils{},
			shouldErr: false,
			testName:  "UnloadGremlinResultIntoStructTest",
		},
		{
			result: &gremlingo.Result{
				Data: map[any]any{
					"id":           "1",
					"lastModified": 1,
					"name":         "test",
					"listTest":     []any{"test1", "test2"},
				},
			},
			v:         testVertexForUtils{},
			shouldErr: true,
			testName:  "UnloadGremlinResultIntoStructTestWithoutPointer",
		},
		{
			testName: "UnloadGremlinResultIntoStructTestWithSlice",
			result: &gremlingo.Result{
				Data: map[any]any{
					"id":           "1",
					"lastModified": 1,
					"listInts":     []any{1.0, 2.0, 3.0},
				},
			},
			shouldErr: false,
			v:         &testVertexWithNumSlice{},
		},
	}

	for _, tt := range unloadGremlinResultIntoStructTests {
		t.Run(
			tt.testName, func(t *testing.T) {
				t.Parallel()
				err := unloadGremlinResultIntoStruct(tt.v, tt.result)
				if (err != nil) != tt.shouldErr {
					t.Errorf(
						"unloadGremlinResultIntoStruct() error = %v, shouldErr %v",
						err,
						tt.shouldErr,
					)
				}
			},
		)
	}

	t.Run(
		"TestStructToMap", func(t *testing.T) {
			t.Parallel()
			v := testVertexForUtils{
				Name: "test",
			}
			name, mapValue, err := structToMap(v)
			if err != nil {
				t.Errorf("Error getting struct name: %v", err)
			}
			if name != "testVertexForUtils" {
				t.Errorf("Struct name should be testVertexForUtils, got %s", name)
			}
			if mapValue["name"] != "test" {
				t.Errorf("Struct name should be test, got %s", mapValue["name"])
			}
		},
	)
	t.Run(
		"TestStructToMapPointer", func(t *testing.T) {
			t.Parallel()
			v := testVertexForUtils{
				Name: "test",
			}
			name, mapValue, err := structToMap(&v)
			if err != nil {
				t.Errorf("Error getting struct name: %v", err)
			}
			if name != "testVertexForUtils" {
				t.Errorf("Struct name should be testVertexForUtils, got %s", name)
			}
			if mapValue["name"] != "test" {
				t.Errorf("Struct name should be test, got %s", mapValue["name"])
			}
		},
	)
	t.Run(
		"TestStructToMapPointerError", func(t *testing.T) {
			t.Parallel()
			_, _, err := structToMap(1)
			if err == nil {
				t.Errorf("No error struct to map: %v", err)
			}
		},
	)
	var i *int
	testsForValidatingStructPointer := []struct {
		testName  string
		v         any
		shouldErr bool
	}{
		{testName: "testNil", v: nil, shouldErr: true},
		{testName: "testStruct", v: gremlingo.Result{}, shouldErr: true},
		{testName: "testStructPointer", v: &gremlingo.Result{}, shouldErr: true},
		{testName: "testStructPointerPointer", v: &testVertexForUtils{}, shouldErr: false},
		{testName: "testStructPointerPointerError", v: i, shouldErr: true},
		{testName: "testStructPointerPointerErrorPointer", v: &i, shouldErr: true},
	}
	for _, tt := range testsForValidatingStructPointer {
		t.Run(
			tt.testName, func(t *testing.T) {
				t.Parallel()
				err := validateStructPointerWithAnonymousVertex(tt.v)
				if (err != nil) != tt.shouldErr {
					t.Errorf(
						"validateStructPointerWithAnonymousVertex() error = %v, shouldErr %v",
						err,
						tt.shouldErr,
					)
				}
			},
		)
	}

}
