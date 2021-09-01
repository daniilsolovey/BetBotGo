package operator

import (
	"strconv"
	"time"

	"github.com/daniilsolovey/BetBotGo/internal/constants"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
)

func getUpcomingEventsForToday(upcomingEvents *requester.UpcomingEvents) (*requester.UpcomingEvents, error) {
	var result requester.UpcomingEvents
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
		log.Warning("convertedTime ", convertedTime)
		if convertedTime.After(timeNow) &&
			convertedTime.Before(timeNow.Truncate(24*time.Hour).Add(21*time.Hour)) {
			event.HumanTime = convertedTime
			result.Results = append(result.Results, event)
		}
	}

	return &result, nil
}

func Truncate(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}

func sortEventsByOdds(eventsWithOdds []requester.EventWithOdds) ([]requester.EventWithOdds, error) {
	var result []requester.EventWithOdds
	var primaryOdds []requester.OddsNumber
	var homeOdd float64
	var awayOdd float64
	var err error
	for _, event := range eventsWithOdds {
		primaryOdds = event.ResultEventWithOdds.Odds.Odds91_1
		if len(primaryOdds) != 0 {
			homeOdd, err = convertStringToFloat(primaryOdds[0].HomeOd)
			if err != nil {
				return nil, err
			}

			awayOdd, err = convertStringToFloat(primaryOdds[0].AwayOd)
			if err != nil {
				return nil, err
			}
		}

		if homeOdd < constants.ODD_FAVORITE_MAX || awayOdd < constants.ODD_FAVORITE_MAX {
			if homeOdd < awayOdd {
				event.Favorite = constants.FAVORITE_IS_HOME
			} else {
				event.Favorite = constants.FAVORITE_IS_AWAY
			}
			result = append(result, event)
		}
	}

	return result, nil
}

func convertStringToFloat(data string) (float64, error) {
	result, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return 0, karma.Format(
			err,
			"unable to convert data to int format, data: %s",
			data,
		)
	}

	return result, nil
}
