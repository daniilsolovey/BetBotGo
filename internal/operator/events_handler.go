package operator

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/daniilsolovey/BetBotGo/internal/constants"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/tools"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
)

func getUpcomingEventsForToday(upcomingEvents *requester.UpcomingEvents) (*requester.UpcomingEvents, error) {
	var result requester.UpcomingEvents
	// moscowTime, err := time.LoadLocation("Europe/Moscow")
	// if err != nil {
	// 	return nil, karma.Format(
	// 		err,
	// 		"unable to load location: Europe/Moscow",
	// 	)
	// }

	// timeNow := time.Now().In(moscowLocation)
	moscowLocation, err := tools.GetTimeMoscowLocation()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get moscow location",
		)
	}

	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get moscow time",
		)
	}

	for _, event := range upcomingEvents.Results {
		parsedTime, err := strconv.ParseInt(event.Time, 10, 64)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to parse time for event_id: %s",
				event.ID,
			)
		}

		convertedTime := time.Unix(parsedTime, 0).In(moscowLocation)
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
			event.HomeOdd = homeOdd
			event.AwayOdd = awayOdd
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

func handleLiveEventOdds(event *requester.EventWithOdds) (bool, error) {
	if len(event.ResultEventWithOdds.Odds.Odds91_1) == 0 {
		return false, errors.New("len of sets is null")
	}

	setData := event.ResultEventWithOdds.Odds.Odds91_1[0].SS
	homeOdd, err := convertStringToFloat(event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd)
	if err != nil {
		return false, err
	}

	awayOdd, err := convertStringToFloat(event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd)
	if err != nil {
		return false, err
	}

	var mainOdd float64

	if event.Favorite == constants.FAVORITE_IS_HOME {
		mainOdd = homeOdd
	} else {
		mainOdd = awayOdd
	}

	winner := getWinnerInFirstSet(setData)
	log.Warning("winner ", winner)
	if winner == "" {
		return false, nil
	}

	if event.Favorite != winner && getNumberOfSet(setData) == 2 && mainOdd > 1.5 {
		return true, nil
	}

	return false, nil
}

func getWinnerInFirstSet(setData string) string {
	liveSetScore := getLiveSetScore(setData)
	numberOfSet := getNumberOfSet(setData)
	log.Warning("numberOfSet ", numberOfSet)

	if numberOfSet == 2 {
		return getWinner(liveSetScore)
	}

	return ""
}

func getNumberOfSet(set string) int {
	liveSetScore := getLiveSetScore(set)
	if liveSetScore == nil {
		return 0
	}

	return len(liveSetScore)
}

func getLiveSetScore(set string) []string {
	if set == "" {
		return nil
	}

	splittedSet := strings.Split(set, ",")
	return splittedSet

}

func getWinner(score []string) string {
	splittedScore := strings.Split(score[0], "-")
	if splittedScore[0] > splittedScore[1] {
		return constants.WINNER_HOME
	} else {
		return constants.WINNER_AWAY
	}
}
