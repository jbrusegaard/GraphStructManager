package main

import (
	"app/comparator"
	GSM "app/gremlin/driver"
	"app/log"
	"app/types"
)

type VertexTesting struct {
	types.Vertex
	TestString string   `json:"testString" gremlin:"testString"`
	TestInt    int      `json:"testInt"    gremlin:"testInt"`
	TestList   []string `json:"testList"   gremlin:"testList"`
}

func main() {
	logger := log.InitializeLogger()
	logger.Info("Logger initialized")
	db, err := GSM.Open("ws://localhost:8182")
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	test1 := VertexTesting{
		TestString: "test",
		TestInt:    1,
		TestList:   []string{"otherField1", "otherField2"},
		// MapField:    map[string]string{"mapField1": "mapField1", "mapField2": "mapField2"},
		// Nest: Nested{
		// 	NestedField: "nested",
		// },
	}

	err = GSM.Create(db, &test1)
	if err != nil {
		logger.Fatal(err)
	}

	test2, err := GSM.Model[VertexTesting](db).Where("testString", comparator.EQ, "test").Take()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info(test2)

}
