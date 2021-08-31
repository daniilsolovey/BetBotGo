package requester

import (
	"strconv"
	"time"

	"github.com/reconquest/karma-go"
)

func getUpcomingEventsForToday(upcomingEvents *UpcomingEvents) (*UpcomingEvents, error) {
	var result UpcomingEvents
	moscowTime, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to load location: Europe/Moscow",
		)
	}

	timeNow := time.Now().In(moscowTime)
	for _, event := range upcomingEvents.Results {
		parsedTime, err := strconv.ParseInt(event.Time, 10, 64)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to parse time for event_id: %s",
				event.ID,
			)
		}

		convertedTime := time.Unix(parsedTime, 0).In(moscowTime)
		if convertedTime.After(timeNow) &&
			convertedTime.Before(timeNow.Truncate(24*time.Hour).Add(21*time.Hour)) {
			result.Results = append(result.Results, event)
		}
	}

	return &result, nil
}

func Truncate(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}
