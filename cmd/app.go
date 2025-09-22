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

	test2, err := GSM.Model[VertexTesting](db).Where("testString", comparator.EQ, "test").First()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info(test2)

	// err = GSM.Create(db, &test1)
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	// var test2 TestVertex
	// // benchmark speed of first
	// start := time.Now()
	// test2, err = GSM.First[TestVertex](db, GSM.QueryOpts{Id: test1.Id})
	// if err != nil {
	// 	return
	// }
	// elapsed := time.Since(start)
	// logger.Infof("First time: %s", elapsed)
	// var test3 TestVertex

	// test3, err = GSM.First[TestVertex](
	// 	db,
	// 	GSM.QueryOpts{Where: gremlingo.T__.Has("test2", "test2")},
	// )
	// if err != nil {
	// 	return
	// }
	// logger.Info(test1)
	// logger.Info(test2)
	// logger.Info(test3)

	// allV, err := GSM.Find[TestVertex](db, GSM.QueryOpts{})
	// if err != nil {
	// 	return
	// }
	// logger.Info("looking through all vertices")
	// for k, v := range allV {
	// 	logger.Info(k, v)
	// }
}
