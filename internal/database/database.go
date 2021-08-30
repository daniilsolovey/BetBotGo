package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/reconquest/karma-go"
)

const (
	ERR_CODE_TABLE_ALREADY_EXISTS = "#42P07"
)

type Database struct {
	name     string
	host     string
	port     string
	user     string
	password string
	client   *pgx.Conn
}

func NewDatabase(
	name, host, port, user, password string,
) *Database {
	database := &Database{
		name:     name,
		host:     host,
		user:     user,
		password: password,
	}

	connection, err := database.connect()
	if err != nil {
		log.Fatal(err)
	}

	database.client = connection

	return database
}

func (database *Database) connect() (*pgx.Conn, error) {
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", database.user, database.password, database.host, database.port, database.name)
	connection, err := pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to connect to the database: %s",
			database.name,
		)
	}

	return connection, nil
}

func (database *Database) Close() error {
	err := database.client.Close(context.Background())
	if err != nil {
		return karma.Format(
			err,
			"unable to close connection to the database",
		)
	}

	return nil
}

func (database *Database) QueryRow() error {
	// err := database.client.QueryRow()
	// if err != nil {
	// 	return karma.Format(
	// 		err,
	// 		"unable to close connection to the database",
	// 	)
	// }

	return nil
}

// defer conn.Close(context.Background())
// var name string
// var weight int64
// err = conn.QueryRow(context.Background(), "select name, weight from widgets where id=$1", 42).Scan(&name, &weight)
// if err != nil {
// 	fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
// 	os.Exit(1)
// }
