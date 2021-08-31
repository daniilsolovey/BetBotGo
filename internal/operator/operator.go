package operator

import (
	"fmt"

	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/reconquest/karma-go"
)

type Operator struct {
	config    *config.Config
	database  *database.Database
	requester *requester.Requester
}

func NewOperator(
	config *config.Config,
	database *database.Database,
	requester *requester.Requester,
) *Operator {
	return &Operator{
		config:    config,
		database:  database,
		requester: requester,
	}
}

func (operator *Operator) HandleRequests() error {

	upcomingEvents, err := operator.requester.GetUpcomingEventsOnCurrentDate()
	if err != nil {
		return karma.Format(
			err,
			"unable to get upcoming events",
		)
	}

	fmt.Println("upcomingEvents", upcomingEvents)
	return nil
}
