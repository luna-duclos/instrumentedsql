// +build go1.10

package instrumentedsql

import "database/sql/driver"

var _ driver.DriverContext = WrappedDriver{}

func (d WrappedDriver) OpenConnector(name string) (driver.Connector, error) {
	var (
		conn driver.Connector
		err  error
	)

	if driver, ok := d.parent.(driver.DriverContext); ok {
		conn, err = driver.OpenConnector(name)
		if err != nil {
			return nil, err
		}
	} else {
		conn = dsnConnector{dsn: name, driver: d.parent}
	}

	var details dbConnDetails
	if !d.omitDbConnectionTags {
		details = newDBConnDetails(name)
	}

	return wrappedConnector{
		Logger: d.Logger,
		childSpanFactory: childSpanFactoryImpl{
			opts:          d.opts,
			dbConnDetails: details,
		},
		parent:    conn,
		driverRef: &d,
	}, nil
}
