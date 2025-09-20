package main

import (
	"app/driver"
	"app/log"
	"app/types"
)

type TestVertex struct {
	types.Vertex
	Test        string   `json:"test"        gremlin:"test"`
	Test2       string   `json:"test2"       gremlin:"test2"`
	OtherField  int      `json:"otherField"  gremlin:"otherField"`
	OtherField2 string   `json:"otherField2" gremlin:"otherField2"`
	OtherFields []string `json:"otherFields" gremlin:"otherFields"`
}

func main() {
	logger := log.InitializeLogger()
	logger.Info("Logger initialized")
	db, err := driver.Open("ws://localhost:8182")
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	test1 := TestVertex{
		Test:        "test",
		Test2:       "test2",
		OtherField:  1,
		OtherField2: "otherField2",
		OtherFields: []string{"otherField1", "otherField2"},
		// MapField:    map[string]string{"mapField1": "mapField1", "mapField2": "mapField2"},
		// Nest: Nested{
		// 	NestedField: "nested",
		// },
	}

	err = driver.Create(db, &test1)
	if err != nil {
		logger.Fatal(err)
	}

	test2, err := driver.Model[TestVertex](db).Where("test", "=", "test").First()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info(test2)

	// err = driver.Create(db, &test1)
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	// var test2 TestVertex
	// // benchmark speed of first
	// start := time.Now()
	// test2, err = driver.First[TestVertex](db, driver.QueryOpts{Id: test1.Id})
	// if err != nil {
	// 	return
	// }
	// elapsed := time.Since(start)
	// logger.Infof("First time: %s", elapsed)
	// var test3 TestVertex

	// test3, err = driver.First[TestVertex](
	// 	db,
	// 	driver.QueryOpts{Where: gremlingo.T__.Has("test2", "test2")},
	// )
	// if err != nil {
	// 	return
	// }
	// logger.Info(test1)
	// logger.Info(test2)
	// logger.Info(test3)

	// allV, err := driver.Find[TestVertex](db, driver.QueryOpts{})
	// if err != nil {
	// 	return
	// }
	// logger.Info("looking through all vertices")
	// for k, v := range allV {
	// 	logger.Info(k, v)
	// }
}
