package main

import (
	"time"

	"app/driver"
	"app/log"
	"app/types"
)

type TestVertex struct {
	types.Vertex
	Test        string   `json:"test"`
	Test2       string   `json:"test2"`
	OtherField  int      `json:"otherField"`
	OtherField2 string   `json:"otherField2"`
	OtherFields []string `json:"otherFields"`
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
	// benchmark speed of first
	start := time.Now()
	err = db.First(&test2, test1.Id)
	if err != nil {
		return
	}
	elapsed := time.Since(start)
	logger.Infof("First time: %s", elapsed)
	logger.Info(test1)
	logger.Info(test2)
}
