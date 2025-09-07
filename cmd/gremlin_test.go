package cmd

import (
	"app/driver"
	"fmt"
	"testing"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func TestGremlin(t *testing.T) {
	conn, err := gremlingo.NewDriverRemoteConnection("ws://localhost:8182/gremlin")
	if err != nil {
		t.Fatalf("Failed to create driver remote connection: %v", err)
	}
	defer conn.Close()
	g := driver.G(conn)
	fmt.Println(g.V())

}
