package driver

import gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"

func G(remoteConnection *gremlingo.DriverRemoteConnection) *gremlingo.GraphTraversalSource {
	return gremlingo.Traversal_().WithRemote(remoteConnection)
}
