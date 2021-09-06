package requester

import "time"

type UpcomingEvents struct {
	Results []Result `json:"results"`
}

type Result struct {
	ID        string `json:"id"`
	SportID   string `json:"sport_id"`
	Time      string `json:"time"`
	HumanTime time.Time
	League    League `json:"league"`
	Home      Home   `json:"home"`
	Away      Away   `json:"away"`
}

type League struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CC   string `json:"cc"`
}

type Home struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CC   string `json:"cc"`
}

type Away struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CC   string `json:"cc"`
}

type EventWithOdds struct {
	EventID             string
	ResultEventWithOdds ResultEventWithOdds `json:"results"`
	League              struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		CC   string `json:"cc"`
	}

	EventStartTime  time.Time
	Favorite        string
	HomeCommandName string
	AwayCommandName string
	HomeOdd         float64
	AwayOdd         float64
}

type ResultEventWithOdds struct {
	Stats struct {
		OddsUpdate struct {
			Oddsupdate91_1 int `json:"91_1"`
			Oddsupdate91_2 int `json:"91_2"`
			Oddsupdate91_3 int `json:"91_3"`
		} `json:"odds_update"`
	} `json:"stats"`
	Odds Odds `json:"odds"`
}

type Odds struct {
	Odds91_1 []OddsNumber `json:"91_1"`
	Odds91_2 []OddsNumber `json:"91_2"`
	Odds91_3 []OddsNumber `json:"91_3"`
}

type OddsNumber struct {
	ID      string `json:"id"`
	HomeOd  string `json:"home_od"`
	AwayOd  string `json:"away_od"`
	SS      string `json:"ss"`
	AddTime string `json:"add_time"`
}

type LiveEventResult struct {
	EventID         string
	Favorite        string
	HomeCommandName string
	AwayCommandName string
	LastHomeOdd     float64
	LastAwayOdd     float64
	LeagueName      string
	Score           string
	CreatedAt       time.Time
}
