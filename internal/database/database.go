package database

import (
	"github.com/go-pg/pg"
	"github.com/reconquest/karma-go"
)

const (
	ERR_CODE_TABLE_ALREADY_EXISTS = "#42P07"
)

type Database struct {
	name     string
	user     string
	password string
	client   *pg.DB
}

type Ticks struct {
	Timestamp int64
	Symbol    string
	AskPrice  float64
	BidPrice  float64
}

func NewDatabase(
	name, user, password string,
) *Database {
	database := &Database{
		name:     name,
		user:     user,
		password: password,
	}

	database.connect()
	return database
}

func (database *Database) connect() {
	database.client = pg.Connect(
		&pg.Options{
			Database: database.name,
			User:     database.user,
			Password: database.password,
		})
}

func (database *Database) Close() error {
	err := database.client.Close()
	if err != nil {
		return karma.Format(
			err,
			"unable to close connection to the database",
		)
	}

	return nil
}
