package main

// import (
//
//	"github.com/jbrusegaard/graph-struct-manager/comparator"
//	"github.com/jbrusegaard/graph-struct-manager/gremlin/driver"
//	"github.com/jbrusegaard/graph-struct-manager/log"
//	"github.com/jbrusegaard/graph-struct-manager/types"
//
// )
//
//	type VertexTesting struct {
//		types.Vertex
//		TestString string   `json:"testString" gremlin:"testString"`
//		TestInt    int      `json:"testInt"    gremlin:"testInt"`
//		TestList   []string `json:"testList"   gremlin:"testList"`
//	}
func main() {
}

// 	logger := log.InitializeLogger()
// 	logger.Info("Logger initialized")
// 	db, err := driver.Open("ws://localhost:8182")
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
// 	defer db.Close()
//
// 	test1 := VertexTesting{
// 		TestString: "test",
// 		TestInt:    1,
// 		TestList:   []string{"otherField1", "otherField2"},
// 		// MapField:    map[string]string{"mapField1": "mapField1", "mapField2": "mapField2"},
// 		// Nest: Nested{
// 		// 	NestedField: "nested",
// 		// },
// 	}
//
// 	err = driver.Create(db, &test1)
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
//
// 	test2, err := driver.Model[VertexTesting](db).Where("testString", comparator.EQ, "test").Take()
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
// 	logger.Info(test2)
// }
