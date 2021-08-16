package instrumentedsql

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionStrings(t *testing.T) {
	const (
		host     = "localhost"
		port     = 5432
		user     = "username"
		password = "VerySecure"
		dbname   = "MyDb"
	)

	for _, tc := range []struct {
		uc  string
		dsn string
		verify func(*testing.T, dbConnDetails)
	}{
		{
			uc: "Full MySQL DSN",
			dsn: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?foo=bar", user, password, host, port, dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Equal(t, host + ":" + strconv.Itoa(port), details.address)
				assert.Equal(t, user, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "mysql", details.dbSystem)
				assert.NotContains(t, details.rawString, password)
				assert.Contains(t, details.rawString, "foo=bar")
			},
		},
		{
			uc: "Minimal MySQL DSN",
			dsn: "/",
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Empty(t, details.address)
				assert.Empty(t, details.user)
				assert.Empty(t, details.dbName)
				assert.Equal(t, "mysql", details.dbSystem)
				assert.Equal(t, "/", details.rawString)
			},
		},
		{
			uc: "MySQL DSN with DB name only",
			dsn: fmt.Sprintf("/%s", dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Empty(t, details.address)
				assert.Empty(t, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "mysql", details.dbSystem)
				assert.Equal(t, "/" + dbname, details.rawString)
			},
		},
		{
			uc: "MySQL DSN without password",
			dsn: fmt.Sprintf("%s@tcp(%s:%d)/%s?foo=bar", user, host, port, dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Equal(t, host + ":" + strconv.Itoa(port), details.address)
				assert.Equal(t, user, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "mysql", details.dbSystem)
				assert.Contains(t, details.rawString, "foo=bar")
			},
		},
		{
			uc: "MySQL DSN with default host and port",
			dsn: fmt.Sprintf("%s:%s@tcp/%s?foo=bar", user, password, dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Empty(t, details.address)
				assert.Equal(t, user, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "mysql", details.dbSystem)
				assert.NotContains(t, details.rawString, password)
				assert.Contains(t, details.rawString, "foo=bar")
			},
		},
		{
			uc: "MySQL DSN with Google Cloud SQL on App Engine",
			dsn: fmt.Sprintf("%s:%s@unix(/cloudsql/project-id:region-name:instance-name)/%s?foo=bar", user, password, dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Equal(t, "/cloudsql/project-id:region-name:instance-name", details.address)
				assert.Equal(t, user, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "mysql", details.dbSystem)
				assert.NotContains(t, details.rawString, password)
				assert.Contains(t, details.rawString, "foo=bar")
			},
		},
		{
			uc: "Full PostgreSQL DSN",
			dsn: fmt.Sprintf("postgres://%s:%s@%s:%d/%s?foo=bar", user, password, host, port, dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Equal(t, host, details.host)
				assert.Equal(t, strconv.Itoa(port), details.port)
				assert.Equal(t, host + ":" + strconv.Itoa(port), details.address)
				assert.Equal(t, user, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "postgres", details.dbSystem)
				assert.NotContains(t, details.rawString, password)
				assert.Contains(t, details.rawString, "foo=bar")
			},
		},
		{
			uc: "Minimal PostgreSQL DSN",
			dsn: "postgres://",
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Empty(t, details.address)
				assert.Empty(t, details.user)
				assert.Empty(t, details.dbName)
				assert.Equal(t, "postgres", details.dbSystem)
				assert.Equal(t, "postgres://", details.rawString)
			},
		},
		{
			uc: "Minimal PostgreSQL DSN with param spec",
			dsn: "postgres://?foo=bar",
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Empty(t, details.address)
				assert.Empty(t, details.user)
				assert.Empty(t, details.dbName)
				assert.Equal(t, "postgres", details.dbSystem)
				assert.Equal(t, "postgres://?foo=bar", details.rawString)
			},
		},
		{
			uc: "PostgreSQL DSN with DB name only",
			dsn: fmt.Sprintf("postgres:///%s", dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Empty(t, details.address)
				assert.Empty(t, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "postgres", details.dbSystem)
				assert.Equal(t, "postgres:///" + dbname, details.rawString)
			},
		},
		{
			uc: "PostgreSQL DSN with host part only",
			dsn: fmt.Sprintf("postgres://%s", host),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Equal(t, host, details.host)
				assert.Empty(t, details.port)
				assert.Equal(t, host, details.address)
				assert.Empty(t, details.user)
				assert.Empty(t, details.dbName)
				assert.Equal(t, "postgres", details.dbSystem)
				assert.Equal(t, "postgres://" + host, details.rawString)
			},
		},
		{
			uc: "PostgreSQL DSN with user and DB name only",
			dsn: fmt.Sprintf("postgres://%s@/%s", user, dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Empty(t, details.address)
				assert.Equal(t, user, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "postgres", details.dbSystem)
				assert.Equal(t, fmt.Sprintf("postgres://%s@/%s", user, dbname), details.rawString)
			},
		},
		{
			uc: "PostgreSQL KV DSN",
			dsn: fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s foo=bar", user, password, host, port, dbname),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Equal(t, host, details.host)
				assert.Equal(t, strconv.Itoa(port), details.port)
				assert.Equal(t, details.host + ":" + details.port, details.address)
				assert.Equal(t, user, details.user)
				assert.Equal(t, dbname, details.dbName)
				assert.Equal(t, "postgres", details.dbSystem)
				assert.NotContains(t, details.rawString, password)
				assert.Contains(t, details.rawString, "foo=bar")
			},
		},
		{
			uc: "SQLite DSN 1",
			dsn: ":memory:",
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Equal(t, ":memory:", details.address)
				assert.Empty(t, details.user)
				assert.Empty(t, details.dbName)
				assert.Equal(t, "sqlite", details.dbSystem)
				assert.Contains(t, details.rawString, ":memory:")
			},
		},
		{
			uc: "SQLite DSN 2",
			dsn: "file:memory:",
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Equal(t, "file:memory:", details.address)
				assert.Empty(t, details.user)
				assert.Empty(t, details.dbName)
				assert.Equal(t, "sqlite", details.dbSystem)
				assert.Contains(t, details.rawString, "file:memory:")
			},
		},
		{
			uc: "SQLite DSN 3",
			dsn: fmt.Sprintf("file:%s?_auth&_auth_user=%s&_auth_pass=%s",dbname, user, password),
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Equal(t, "file:" + dbname, details.address)
				assert.Equal(t, user, details.user)
				assert.Empty(t, details.dbName)
				assert.Equal(t, "sqlite", details.dbSystem)
				assert.NotContains(t, details.rawString, password)
				assert.Contains(t, details.rawString, "file:" + dbname)
			},
		},
		{
			uc: "DSN Reference",
			dsn: "DSN=FooBar",
			verify: func(t *testing.T, details dbConnDetails) {
				assert.Empty(t, details.host)
				assert.Empty(t, details.port)
				assert.Empty(t, details.address)
				assert.Empty(t, details.user)
				assert.Empty(t, details.dbName)
				assert.Empty(t, details.dbSystem)
				assert.Contains(t, details.rawString, "DSN=FooBar")
			},
		},
	} {
		t.Run("case="+tc.uc, func(t *testing.T) {
			cd := newDBConnDetails(tc.dsn)
			tc.verify(t, cd)
		})
	}
}
