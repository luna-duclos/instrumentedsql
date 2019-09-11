// +build go1.10

package instrumentedsql

import (
	"context"
	"database/sql/driver"
)

type wrappedConnector struct {
	parent    driver.Connector
	driverRef *wrappedDriver
}

func (d wrappedDriver) OpenConnector(name string) (driver.Connector, error) {
	driver, ok := d.parent.(driver.DriverContext)
	if !ok {
		return wrappedConnector{
			parent:    dsnConnector{dsn: name, driver: &d},
			driverRef: &d,
		}, nil
	}
	conn, err := driver.OpenConnector(name)
	if err != nil {
		return nil, err
	}

	return wrappedConnector{parent: conn, driverRef: &d}, nil
}

func (c wrappedConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.parent.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return wrappedConn{opts: c.driverRef.opts, parent: conn}, nil
}

func (c wrappedConnector) Driver() driver.Driver {
	return c.driverRef
}

func (c wrappedConn) ResetSession(ctx context.Context) error {
	conn, ok := c.parent.(driver.SessionResetter)
	if !ok {
		return nil
	}

	return conn.ResetSession(ctx)
}

// dsnConnector is a fallback connector placed in position of wrappedConnector.parent
// when given Driver does not comply with DriverContext interface.
type dsnConnector struct {
	dsn    string
	driver driver.Driver
}

func (t dsnConnector) Connect(_ context.Context) (driver.Conn, error) {
	return t.driver.Open(t.dsn)
}

func (t dsnConnector) Driver() driver.Driver {
	return t.driver
}
