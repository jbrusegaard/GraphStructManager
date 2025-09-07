package main

import (
	"fmt"

	"app/driver"
	"app/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func main() {
	logger := log.InitializeLogger()
	conn, err := gremlingo.NewDriverRemoteConnection("ws://localhost:8182/gremlin")
	if err != nil {
		logger.Fatalf("Failed to create driver remote connection: %v", err)
	}
	defer conn.Close()
	g := driver.G(conn)
	fmt.Println(g.V().ToList())
}
