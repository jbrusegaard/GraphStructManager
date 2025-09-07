package cmd

import (
	"testing"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func TestGremlin(t *testing.T) {
	_, err := gremlingo.NewDriverRemoteConnection("http://localhost:8182/gremlin")
	if err != nil {
		t.Fatalf("Failed to create driver remote connection: %v", err)
	}

}
