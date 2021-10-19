package operator

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/daniilsolovey/BetBotGo/internal/constants"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/tools"
	"github.com/reconquest/karma-go"
)

func getUpcomingEventsForToday(upcomingEvents *requester.UpcomingEvents) (*requester.UpcomingEvents, error) {
	var result requester.UpcomingEvents
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
			convertedTime.Before(timeNow.Truncate(24*time.Hour).Add(21*time.Hour).Add(1*time.Second)) {
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
			if primaryOdds[0].HomeOd == "-" || primaryOdds[0].AwayOd == "-" {
				continue
			}
		}

		if len(primaryOdds) != 0 {
			homeOdd, err = convertStringToFloat(primaryOdds[0].HomeOd)
			if err != nil {
				return nil, err
			}

			awayOdd, err = convertStringToFloat(primaryOdds[0].AwayOd)
			if err != nil {
				return nil, err
			}
		} else {
			continue
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

func handleEventsByLeagues(events []requester.EventWithOdds) []requester.EventWithOdds {
	countries := strings.Split(constants.COUNTRIES, ",")
	var result []requester.EventWithOdds
	for _, event := range events {
		for _, country := range countries {
			if strings.Contains(event.League.Name, country) {
				if isSpecificLeagueContainsWomen(event.League.Name, country) {
					continue
				}
				result = append(result, event)
			}
		}
	}

	return result
}

func isSpecificLeagueContainsWomen(leagueName, country string) bool {
	countries := strings.Split(constants.SPECIFIC_COUNTRIES, ",")
	if tools.Find(countries, country) {
		if strings.Contains(leagueName, constants.WOMEN) {
			return true
		} else {
			return false
		}
	}

	return false
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

func handleLiveEventOdds(event requester.EventWithOdds) (bool, int, error) {
	if reflect.DeepEqual(event.ResultEventWithOdds.Odds, requester.Odds{}) {
		return false, 0, errors.New("event.ResultEventWithOdds.Odds is empty")
	}

	if len(event.ResultEventWithOdds.Odds.Odds91_1) == 0 {
		return false, 0, errors.New("len of sets is null")
	}

	setData := event.ResultEventWithOdds.Odds.Odds91_1[0].SS
	if event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd == "-" {
		event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "0000"
	}

	if event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd == "-" {
		event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "0000"
	}

	homeOdd, err := convertStringToFloat(event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd)
	if err != nil {
		return false, 0, err
	}

	awayOdd, err := convertStringToFloat(event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd)
	if err != nil {
		return false, 0, err
	}

	var mainOdd float64

	if event.Favorite == constants.FAVORITE_IS_HOME {
		mainOdd = homeOdd
	} else {
		mainOdd = awayOdd
	}

	winner := getWinnerInFirstSet(setData)
	if winner == "" {
		return false, 0, nil
	}

	if getNumberOfSet(setData) == 3 {
		return false, 3, nil
	}

	if event.Favorite != winner && getNumberOfSet(setData) == 2 && mainOdd > 1.5 {
		return true, 2, nil
	} else {
		return false, 2, nil
	}

}

func handleFinalLiveSet(event requester.EventWithOdds) (int, error) {
	if len(event.ResultEventWithOdds.Odds.Odds91_1) == 0 {
		return 0, errors.New("len of sets is null")
	}

	setData := event.ResultEventWithOdds.Odds.Odds91_1[0].SS

	if getNumberOfSet(setData) == 3 {
		return 3, nil
	}

	return 0, nil
}

func getWinnerInFirstSet(setData string) string {
	liveSetScore := getLiveSetScore(setData)
	numberOfSet := getNumberOfSet(setData)
	if numberOfSet == 2 {
		return getWinner(liveSetScore)
	}

	return ""
}

func getWinnerInSecondSet(setData string) string {
	liveSetScore := getLiveSetScore(setData)
	numberOfSet := getNumberOfSet(setData)
	if numberOfSet == 3 {
		splittedScore := strings.Split(liveSetScore[1], "-")
		if splittedScore[0] > splittedScore[1] {
			return constants.WINNER_HOME
		} else {
			return constants.WINNER_AWAY
		}
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
