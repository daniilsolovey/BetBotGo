package statistics

import "github.com/daniilsolovey/BetBotGo/internal/database"

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

func (statistics *Statistics) GetLiveEventsResultsOnCurrentDate() {

}
