package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
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
	client   *pgxpool.Pool
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

func (database *Database) connect() (*pgxpool.Pool, error) {
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", database.user, database.password, database.host, database.port, database.name)
	// connection, err := pgx.Connect(context.Background(), databaseUrl)
	connection, err := pgxpool.Connect(context.Background(), databaseUrl)
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
	database.client.Close()
	return nil
}

func (database *Database) CreateTables() error {
	log.Infof(
		karma.Describe("database", database.name),
		"create tables in database",
	)

	_, err := database.client.Query(
		context.Background(),
		SQL_CREATE_TABLE_UPCOMING_EVENTS)
	if err != nil {
		return karma.Format(
			err,
			"unable to create upcoming_events table in the database",
		)
	}

	return nil
}

func (database *Database) InsertEventsForToday(events []requester.EventWithOdds) error {
	log.Infof(
		karma.Describe("database", database.name),
		"inserting events in database",
	)

	for _, event := range events {
		_, err := database.client.Query(
			context.Background(),
			SQL_INSERT_EVENTS_FOR_TODAY,
			event.EventID,
			event.HumanTime,
			event.League.ID,
			event.League.Name,
			event.Favorite,
			event.HomeOdd,
			event.AwayOdd,
		)
		if err != nil {
			if strings.Contains(err.Error(), "ERROR: duplicate key value violates unique constraint") {
				continue
			}

			return karma.Format(
				err,
				"unable to add event to the database,"+
					" event: %v, event_id: %s",
				event, event.EventID,
			)
		}
	}

	log.Info("events successfully added")
	return nil
}

// func (database *Database) InsertEventsForToday(events []requester.EventWithOdds) error {
// 	log.Infof(
// 		karma.Describe("database", database.name),
// 		"inserting events in database",
// 	)

// 	transaction, err := database.client.Begin(context.Background())
// 	if err != nil {
// 		return karma.Format(
// 			err,
// 			"unable to start a sql transaction",
// 		)
// 	}

// 	for _, event := range events {
// 		_, err = transaction.Exec(
// 			context.Background(),
// 			SQL_INSERT_EVENTS_FOR_TODAY,
// 			event.EventID,
// 			event.HumanTime,
// 			event.League.ID,
// 			event.League.Name,
// 			event.Favorite,
// 			event.HomeOdd,
// 			event.AwayOdd,
// 		)
// 		if err != nil {
// 			errRollback := transaction.Rollback(context.Background())
// 			if errRollback != nil {
// 				return karma.Format(
// 					errRollback,
// 					"unable to rollback transaction!",
// 				)
// 			}

// 			if strings.Contains(err.Error(), "ERROR: duplicate key value violates unique constraint") {
// 				continue
// 			}

// 			return karma.Format(
// 				err,
// 				"unable to add event to the database,"+
// 					" event: %v, event_id: %s",
// 				event, event.EventID,
// 			)
// 		}
// 	}

// 	err = transaction.Commit(context.Background())
// 	if err != nil {
// 		return karma.Format(
// 			err,
// 			"unable to commit transaction for adding events",
// 		)
// 	}

// 	log.Info("events successfully added")
// 	return nil
// }
