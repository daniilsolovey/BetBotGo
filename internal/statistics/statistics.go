package statistics

import (
	"fmt"
	"math"

	"github.com/daniilsolovey/BetBotGo/internal/constants"
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/daniilsolovey/BetBotGo/internal/operator"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/transport"
	"github.com/reconquest/karma-go"
)

const (
	PLAYER_IS_WIN                   = "true"
	TEXT_STATISTICS_ON_PREVIOUS_DAY = "Результаты за вчера:\n" +
		"  win: %d\n" +
		"  lose: %d\n" +
		"  average odd: %f\n"
	TEXT_STATISTICS_ON_PREVIOUS_WEEK = "Результаты за прошлую неделю:\n" +
		"  win: %d\n" +
		"  lose: %d\n" +
		"  average odd: %f\n"
)

type ResultOfPreviousDay struct {
	Win        int
	Lose       int
	AverageOdd float64
}

type Statistics struct {
	database  *database.Database
	transport transport.Transport
}

func NewStatistics(
	database *database.Database,
	transport transport.Transport,
) *Statistics {
	statistics := &Statistics{
		database:  database,
		transport: transport,
	}
	return statistics
}

func (statistics *Statistics) GetStatisticOnPreviousDayAndNotify() error {
	events, err := statistics.getLiveEventsResultsOnPreviousDateAndWriteToStatistic()
	if err != nil {
		return karma.Format(
			err,
			"unable to get live events results on current date",
		)
	}

	handledEvents := handleResultsOfPreviousDay(events)
	text := fmt.Sprintf(
		TEXT_STATISTICS_ON_PREVIOUS_DAY,
		handledEvents.Win,
		handledEvents.Lose,
		handledEvents.AverageOdd,
	)

	err = statistics.transport.SendMessage(operator.TEMP_RECIPIENT, text)
	if err != nil {
		return karma.Format(
			err,
			"unable to send statistic on previous day to telegram",
		)
	}

	return nil
}

func (statistics *Statistics) GetStatisticOnPreviousWeekAndNotify() error {
	results, err := statistics.database.GetStatisticOnPreviousWeek()
	if err != nil {
		return karma.Format(
			err,
			"unable to get statistic results on previous week",
		)
	}

	handledResults := handleResultsOfPreviousWeek(results)
	text := fmt.Sprintf(
		TEXT_STATISTICS_ON_PREVIOUS_WEEK,
		handledResults.Win,
		handledResults.Lose,
		handledResults.AverageOdd,
	)

	err = statistics.transport.SendMessage(operator.TEMP_RECIPIENT, text)
	if err != nil {
		return karma.Format(
			err,
			"unable to send statistic on previous day to telegram",
		)
	}

	return nil
}

func (statistics *Statistics) getLiveEventsResultsOnPreviousDateAndWriteToStatistic() (
	[]requester.LiveEventResult,
	error,
) {
	events, err := statistics.database.GetLiveEventsResultsOnPreviousDate()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get live events results on previous day",
		)
	}

	err = statistics.database.InsertEventsResultsToStatistic(events)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to insert events results on previous day",
		)
	}

	return events, nil
}

func handleResultsOfPreviousWeek(
	data []database.StatisticResultOfPreviousDay,
) ResultOfPreviousDay {
	var result ResultOfPreviousDay
	for _, item := range data {
		if item.PlayerIsWin == PLAYER_IS_WIN {
			result.Win = result.Win + 1
		} else {
			result.Lose = result.Lose + 1
		}
	}

	return result
}

func handleResultsOfPreviousDay(
	events []requester.LiveEventResult,
) ResultOfPreviousDay {
	var result ResultOfPreviousDay
	for _, event := range events {
		if handleResultOfPreviousDay(event) {
			result.Win = result.Win + 1
		} else {
			result.Lose = result.Lose + 1
		}
	}

	result.AverageOdd = getAverageOdd(events)
	return result
}

func handleResultOfPreviousDay(event requester.LiveEventResult) bool {
	if event.Favorite == event.WinnerInSecondSet {
		return true
	}

	return false
}

func getAverageOdd(events []requester.LiveEventResult) float64 {
	var allOdds []float64
	for _, event := range events {
		if event.Favorite == event.WinnerInSecondSet {
			if event.Favorite == constants.FAVORITE_IS_HOME {
				allOdds = append(allOdds, event.LastHomeOdd)
			} else {
				allOdds = append(allOdds, event.LastAwayOdd)
			}
		}
	}

	var sum float64
	for i := 0; i < len(allOdds); i++ {
		sum += allOdds[i]
	}

	averageOdd := (sum) / float64((len(allOdds)))
	return roundNumber(averageOdd, 0.01)
}

func roundNumber(number, unit float64) float64 {
	return math.Round(number/unit) * unit
}
