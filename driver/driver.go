package driver

import (
	"fmt"

	appLogger "app/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/charmbracelet/log"
)

type GremlinDriver struct {
	remoteConn *gremlingo.DriverRemoteConnection
	g          *gremlingo.GraphTraversalSource
	logger     *log.Logger
}

type QueryOpts struct {
	Id    any
	Where *gremlingo.GraphTraversal
}

func g(remoteConnection *gremlingo.DriverRemoteConnection) *gremlingo.GraphTraversalSource {
	return gremlingo.Traversal_().WithRemote(remoteConnection)
}

func Open(url string) (*GremlinDriver, error) {
	driverLogger := appLogger.InitializeLogger()
	driverLogger.Infof("Opening driver with url: %s/gremlin", url)
	remote, err := gremlingo.NewDriverRemoteConnection(fmt.Sprintf("%s/gremlin", url))
	if err != nil {
		return nil, err
	}

	driver := &GremlinDriver{
		g:          g(remote),
		remoteConn: remote,
		logger:     driverLogger,
	}
	return driver, nil
}

func (driver *GremlinDriver) Close() {
	driver.remoteConn.Close()
}
