package main

import (
	"app/driver"
	"app/log"
	"app/types"
)

type TestVertex struct {
	types.Vertex
	Test        string   `gremlin:"test"`
	Test2       string   `gremlin:"test2"`
	OtherField  int      `gremlin:"otherField"`
	OtherField2 string   `gremlin:"otherField2"`
	OtherFields []string `gremlin:"otherFields"`
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
	}
	err = db.Create(&test1)
	if err != nil {
		logger.Fatal(err)
	}

	var test2 TestVertex

	err = db.First(&test2, test1.Id)
	if err != nil {
		return
	}
	logger.Info(test1)
}
