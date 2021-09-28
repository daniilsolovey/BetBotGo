package database

const (
	SQL_CREATE_TABLE_UPCOMING_EVENTS = `
	CREATE TABLE IF NOT EXISTS
	events_volleyball(
		event_id VARCHAR(50) UNIQUE NOT NULL PRIMARY KEY,
		event_time TIMESTAMP,
		league_id VARCHAR(50),
		league_name VARCHAR(50),
		favorite_name VARCHAR(50),
		home_command_name VARCHAR(50),
		away_command_name VARCHAR(50),
		odd_home DECIMAL,
		odd_away DECIMAL
	);
`
	SQL_INSERT_EVENTS_FOR_TODAY = `
	INSERT INTO
	events_volleyball(
		event_id,
		event_time,
		league_id,
		league_name,
		favorite_name,
		home_command_name,
		away_command_name,
		odd_home,
		odd_away
	)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);
`
	SQL_CREATE_TABLE_LIVE_EVENTS_RESULTS = `
	CREATE TABLE IF NOT EXISTS
		live_events_results(
			id serial PRIMARY KEY,
    		event_id VARCHAR(50),
			last_odd_home VARCHAR(50),
			last_odd_away VARCHAR(50),
			score VARCHAR(50),
			winner_in_second_set VARCHAR(20),
			favorite VARCHAR(20),
    		created_at TIMESTAMP,
    		FOREIGN KEY (event_id) REFERENCES events_volleyball (event_id)
	);
`
	SQL_CREATE_TABLE_STATISTIC_ON_PREVIOUS_DAY = `
	CREATE TABLE IF NOT EXISTS
	statistic_on_previous_day(
			id serial PRIMARY KEY,
			event_id VARCHAR(50) UNIQUE NOT NULL,
			player_is_win VARCHAR(20),
			score VARCHAR(50),
			winner_in_second_set VARCHAR(20),
    		created_at TIMESTAMP,
    		FOREIGN KEY (event_id) REFERENCES events_volleyball (event_id)
	);
`
	//player_is_win should contain only true/false

	SQL_INSERT_STATISTIC_ON_PREVIOUS_DAY = `
	INSERT INTO
		statistic_on_previous_day(
			event_id,
			player_is_win,
			score,
			winner_in_second_set,
    		created_at
	)
	VALUES($1, $2, $3, $4, $5);
`
	SQL_UPDATE_LIVE_EVENTS_RESULTS_SCORE_AND_WINNER = `
	UPDATE live_events_results
		SET
			score = $1,
			winner_in_second_set = $2
	WHERE live_events_results.event_id = $3;
`

	SQL_INSERT_LIVE_EVENTS_RESULTS = `
	INSERT INTO
	live_events_results(
		event_id,
		last_odd_home,
		last_odd_away,
		score,
		winner_in_second_set,
		favorite,
		created_at
	)
	VALUES($1, $2, $3, $4, $5, $6, $7);
`

	SQL_CREATE_TABLE_TELEGRAM_SUBSCRIBERS = `
	CREATE TABLE IF NOT EXISTS
	telegram_subscribers(
		id serial PRIMARY KEY,
		secret_key VARCHAR(50),
		secret_key_expired_at TIMESTAMP,
		created_at TIMESTAMP
	);
`

	SQL_SELECT_LIVE_EVENTS_AT_END_OF_DAY = `
	SELECT * FROM live_events_results
	WHERE CAST($1 AS Date) = CAST(live_events_results.created_at AS Date);
`

	SQL_SELECT_STATISTICS_OF_PREVIOUS_WEEK = `
	SELECT * FROM statistic_on_previous_day
	WHERE (created_at >= CAST($1 AS Date));
	`
)
