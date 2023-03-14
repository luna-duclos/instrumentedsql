module github.com/luna-duclos/instrumentedsql/google

go 1.14

require (
	cloud.google.com/go v0.45.1
	cloud.google.com/go/pubsub v1.0.1 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/luna-duclos/instrumentedsql v1.1.3
	go.opencensus.io v0.22.1 // indirect
)

replace (
	github.com/luna-duclos/instrumentedsql => ../
)
