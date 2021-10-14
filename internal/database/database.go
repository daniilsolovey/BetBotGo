package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/tools"
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

type StatisticResultOfPreviousDay struct {
	EventID           string
	PlayerIsWin       string
	Score             string
	WinnerInSecondSet string
	CreatedAt         time.Time
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

	log.Info("creating events_volleyball table")
	_, err := database.client.Query(
		context.Background(),
		SQL_CREATE_TABLE_UPCOMING_EVENTS,
	)
	if err != nil {
		return karma.Format(
			err,
			"unable to create events_volleyball table in the database",
		)
	}

	log.Info("events_volleyball table successfully created")

	log.Info("creating live_events_results table")
	_, err = database.client.Query(
		context.Background(),
		SQL_CREATE_TABLE_LIVE_EVENTS_RESULTS,
	)
	if err != nil {
		return karma.Format(
			err,
			"unable to create live_events_results table in the database",
		)
	}

	log.Info("creating statistic_on_current_day table")
	_, err = database.client.Query(
		context.Background(),
		SQL_CREATE_TABLE_STATISTIC_ON_PREVIOUS_DAY,
	)
	if err != nil {
		return karma.Format(
			err,
			"unable to create statistic_on_current_day table in the database",
		)
	}

	log.Info("statistic_on_current_day table successfully created")
	return nil
}

func (database *Database) InsertEventsForToday(events []requester.EventWithOdds) error {
	if len(events) != 0 {
		log.Infof(
			karma.Describe("database", database.name),
			"inserting events in database",
		)
	} else {
		log.Infof(
			karma.Describe("database", database.name),
			"empty list with events for today",
		)
		return nil
	}

	for _, event := range events {
		log.Infof(nil, "inserting event: %v", event)
		rows, err := database.client.Query(
			context.Background(),
			SQL_INSERT_EVENTS_FOR_TODAY,
			event.EventID,
			event.EventStartTime,
			event.League.ID,
			event.League.Name,
			event.Favorite,
			event.HomeCommandName,
			event.AwayCommandName,
			event.HomeOdd,
			event.AwayOdd,
		)
		if rows.Err() != nil {
			return karma.Format(
				err,
				"error with rows unable to add event to the database,"+
					" event: %v",
				event,
			)
		}

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
		rows.Close()
	}

	log.Info("events successfully added")
	return nil
}

func (database *Database) InsertEventsResultsToStatistic(events []requester.LiveEventResult) error {
	if len(events) != 0 {
		log.Infof(
			karma.Describe("database", database.name),
			"inserting events to statistic in database",
		)
	} else {
		log.Infof(
			karma.Describe("database", database.name),
			"empty list with events for previous day when inserting in statistic",
		)
		return nil
	}

	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return karma.Format(
			err,
			"unable to get current moscow time before inserting events results to statistic",
		)
	}

	for _, event := range events {
		var playerIsWin string // can be "true"/"false" only
		if event.WinnerInSecondSet == event.Favorite {
			playerIsWin = "true"
		} else {
			playerIsWin = "false"
		}
		rows, err := database.client.Query(
			context.Background(),
			SQL_INSERT_STATISTIC_ON_PREVIOUS_DAY,
			event.EventID,
			playerIsWin,
			event.Score,
			event.WinnerInSecondSet,
			timeNow,
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

		rows.Close()
	}

	log.Info("events successfully added to statistic")
	return nil
}

func (database *Database) UpdateLiveEventsResultsScoreAndWinnerFields(eventID, setData, winner string) error {
	log.Infof(
		karma.Describe("database", database.name),
		"update live event result score and winner of second set in database",
	)

	rows, err := database.client.Query(
		context.Background(),
		SQL_UPDATE_LIVE_EVENTS_RESULTS_SCORE_AND_WINNER,
		setData,
		winner,
		eventID,
	)
	defer func() {
		rows.Close()
	}()
	if err != nil {
		return karma.Format(
			err,
			"unable to update live event in the database,"+
				" event_id: %s",
			eventID,
		)
	}

	log.Infof(nil, "live event with final results successfully updated, event_id: %s", eventID)
	return nil
}

func (database *Database) InsertLiveEventResult(event requester.EventWithOdds) error {
	log.Infof(
		karma.Describe("database", database.name),
		"inserting live event result in database",
	)

	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return karma.Format(
			err,
			"unable to get current time before inserting data to live_events_results",
		)
	}

	rows, err := database.client.Query(
		context.Background(),
		SQL_INSERT_LIVE_EVENTS_RESULTS,
		event.EventID,
		event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd,
		event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd,
		event.ResultEventWithOdds.Odds.Odds91_1[0].SS,
		event.WinnerInSecondSet,
		event.Favorite,
		timeNow,
	)
	defer func() {
		rows.Close()
	}()
	if err != nil {
		return karma.Format(
			err,
			"unable to add live event to the database,"+
				" event: %v, event_id: %s",
			event, event.EventID,
		)
	}

	log.Info("live event successfully added")
	return nil
}

func (database *Database) GetLiveEventsResultsOnPreviousDate() ([]requester.LiveEventResult, error) {
	log.Info("receiving live events results on previous date before inserting it to statistic")
	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get current moscow time for receive events results on previous date",
		)
	}
	rows, err := database.client.Query(
		context.Background(),
		SQL_SELECT_LIVE_EVENTS_AT_END_OF_DAY,
		timeNow.Add(-24*time.Hour),
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, karma.Format(
			err,
			"unable to get live events on previous date from the database",
		)
	}

	defer func() {
		rows.Close()
		log.Error(err)
	}()

	var result []requester.LiveEventResult
	for rows.Next() {
		var (
			liveEvent   requester.LiveEventResult
			lastHomeOdd string
			lastAwayOdd string
			id          int
		)
		err := rows.Scan(
			&id,
			&liveEvent.EventID,
			&lastHomeOdd,
			&lastAwayOdd,
			&liveEvent.Score,
			&liveEvent.WinnerInSecondSet,
			&liveEvent.Favorite,
			&liveEvent.CreatedAt,
		)
		if err != nil {
			return nil, karma.Format(
				err,
				"error during scaning live event results from database rows",
			)
		}

		convertedLastHomeOdd, err := strconv.ParseFloat(lastHomeOdd, 64)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to parse home odd",
			)
		}

		convertedLastAwayOdd, err := strconv.ParseFloat(lastAwayOdd, 64)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to parse home odd",
			)
		}

		liveEvent.LastHomeOdd = convertedLastHomeOdd
		liveEvent.LastAwayOdd = convertedLastAwayOdd
		result = append(result, liveEvent)
	}

	return result, nil
}

func (database *Database) GetStatisticOnPreviousWeek() ([]StatisticResultOfPreviousDay, error) {
	log.Info("receiving live events results on previous week")
	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get current moscow time for receive statistic on previous week",
		)
	}

	rows, err := database.client.Query(
		context.Background(),
		SQL_SELECT_STATISTICS_OF_PREVIOUS_WEEK,
		timeNow.Add(-24*time.Hour*7),
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, karma.Format(
			err,
			"unable to get statistic on previous week from the database",
		)
	}

	defer func() {
		rows.Close()
		log.Error(err)
	}()

	var results []StatisticResultOfPreviousDay
	for rows.Next() {
		var (
			id     int
			result StatisticResultOfPreviousDay
		)
		err := rows.Scan(
			&id,
			&result.EventID,
			&result.PlayerIsWin,
			&result.Score,
			&result.WinnerInSecondSet,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, karma.Format(
				err,
				"error during scaning results of previous week from database rows",
			)
		}

		results = append(results, result)
	}

	return results, nil
}

func (database *Database) GetUpcomingEventsForToday() ([]requester.EventWithOdds, error) {
	log.Info("receiving upcoming events for today for viewing in handler")
	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get current moscow time for receive upcoming_events_handler for today",
		)
	}

	rows, err := database.client.Query(
		context.Background(),
		SQL_SELECT_UPCOMING_EVENTS_FOR_CURRENT_DAY,
		"2021-10-10",
	)

	log.Warning("timeNOw ", timeNow)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, karma.Format(
			err,
			"unable to get upcoming_events_handler for current day from the database",
		)
	}

	defer func() {
		rows.Close()
	}()

	var results []requester.EventWithOdds
	for rows.Next() {
		var (
			event requester.EventWithOdds
		)
		err := rows.Scan(
			&event.EventID,
			&event.EventStartTime,
			&event.League.ID,
			&event.League.Name,
			&event.Favorite,
			&event.HomeCommandName,
			&event.AwayCommandName,
			&event.HomeOdd,
			&event.AwayOdd,
		)
		if err != nil {
			return nil, karma.Format(
				err,
				"error during scaning upcoming_events_handler for current day from database rows",
			)
		}

		results = append(results, event)
	}

	log.Info("upcoming_events_handler successfully received")

	return results, nil
}
