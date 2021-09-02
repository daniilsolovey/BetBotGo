package database

const (
	SQL_CREATE_TABLE_UPCOMING_EVENTS = `
CREATE TABLE IF NOT EXISTS
	events_volleyball(
		id SERIAL NOT NULL PRIMARY KEY,
		event_id VARCHAR(50),
		event_time TIMESTAMP,
		league_id VARCHAR(50),
		league_name VARCHAR(50),
		favorite_name VARCHAR(50),
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
		odd_home,
		odd_away
	)
	VALUES($1, $2, $3, $4, $5, $6, $7);
`
)

// SQL_CREATE_TABLE_UPCOMING_EVENTS = `
// CREATE TABLE IF NOT EXISTS events_volleyball (
// 	id INT NOT NULL PRIMARY KEY,
// 	event_id INT,
// 	sport_id INT,
// 	event_time INT,
// 	time_status INT,
// 	league_name VARCHAR(50),
// 	league_id INT,
// 	home_id INT,
// 	home_name VARCHAR(50),
// 	away_id INT,
// 	away_id_name VARCHAR(50),
// 	odd_home DECIMAL,
// 	odd_away DECIMAL,
// 	odd_add_time INT

// );
// `
