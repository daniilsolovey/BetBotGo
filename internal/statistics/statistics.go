package statistics

import (
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/reconquest/karma-go"
)

type Statistics struct {
	database *database.Database
}

func NewStatistics(
	database *database.Database,
) *Statistics {
	statistics := &Statistics{
		database: database,
	}
	return statistics
}

func (statistics *Statistics) GetLiveEventsResultsOnCurrentDateAndWriteToStatistic() error {
	events, err := statistics.database.GetLiveEventsResultsOnCurrentDate()
	if err != nil {
		return karma.Format(
			err,
			"unable to get live events results on current date",
		)
	}

	statistics.database.InsertEventsResultsToStatistic(events)

	return nil
}
