package db

import (
	"fmt"
	"os"
	"go.mongodb.org/mongo-driver/mongo"
)








func ConnectMgDB(applicationName string, cred map[string]string) (conn *pgx.ConnPool) {
	var runtimeParams map[string]string
	runtimeParams = make(map[string]string)
	runtimeParams["application_name"] = applicationName
	connConfig := pgx.ConnConfig{
		User:              cred["user"],
		Password:          cred["password"],
		Host:              cred["host"],
		Port:              5432,
		Database:          cred["user"],
		TLSConfig:         nil,
		UseFallbackTLS:    false,
		FallbackTLSConfig: nil,
		RuntimeParams:     runtimeParams,
	}
	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: 30,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to establish connection pool: %v\n", err)
		os.Exit(1)
	}
	return pool
}
