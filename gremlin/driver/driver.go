package driver

import (
	"fmt"

	"app/comparator"
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

// Label returns a query builder for a specific label
func (driver *GremlinDriver) Label(label string) *RawQuery {
	return &RawQuery{
		db:    driver,
		label: label,
	}
}
func Save[T VertexType](driver *GremlinDriver, v *T) error {
	if (*v).GetVertexId() == nil {
		return Create(driver, v)
	}
	return Update(driver, v)
}

// Package-level generic functions

// Model returns a new query builder for the specified type
func Model[T VertexType](driver *GremlinDriver) *Query[T] {
	return NewQuery[T](driver)
}

// Where is a convenience method that creates a new query with a condition
func Where[T VertexType](
	driver *GremlinDriver,
	field string,
	operator comparator.Comparator,
	value any,
) *Query[T] {
	return NewQuery[T](driver).Where(field, operator, value)
}
