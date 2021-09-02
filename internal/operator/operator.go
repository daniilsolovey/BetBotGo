package operator

import (
	"fmt"
	"time"

	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/tools"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
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

func (operator *Operator) GetEventsForToday() ([]requester.EventWithOdds, error) {
	upcomingEvents, err := operator.requester.GetUpcomingEventsOnCurrentDate()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get upcoming events",
		)
	}
	log.Info("handle upcoming events and receiving events for today")

	upcomingEventsForToday, err := getUpcomingEventsForToday(upcomingEvents)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get upcoming events for today",
		)
	}

	log.Info("upcoming events for today received successfully")

	eventsWithOdds, err := operator.requester.GetEventOddsByEventIDs(upcomingEventsForToday)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get upcoming events",
		)
	}

	sortedEventsWithOdds, err := sortEventsByOdds(eventsWithOdds)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to sort events by odds",
		)
	}

	return sortedEventsWithOdds, nil
}

func (operator *Operator) EventRotuneGetLiveOdds(event *requester.EventWithOdds) error {
	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return karma.Format(
			err,
			"unable to get moscow time",
		)
	}

	diff := event.HumanTime.Sub(timeNow)
	time.Sleep(diff)
	errCount := 0
	liveEventResult := false
	for {
		liveEvent, err := operator.requester.GetLiveEventByID(event.EventID)
		if err != nil {
			if errCount < 3 {
				log.Errorf(err, "unable to get live event data by event_id: %s", event.EventID)
				errCount = 1
				time.Sleep(10 * time.Second)
			} else {
				break
			}
		}

		liveEventResult, err := handleLiveEventOdds(liveEvent)
		if err != nil {
			if errCount < 3 {
				log.Errorf(err, "unable to handle live event event_id: %s", liveEvent.EventID)
				errCount = 1
				time.Sleep(10 * time.Second)
			} else {
				break
			}
			if liveEventResult {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	if liveEventResult {
		fmt.Println("sent to bot")
		// send message to telegram bot
	}
	return nil

}

func (operator *Operator) CreateRoutinesForEachEvent([]requester.EventWithOdds) error {
	// timeNow, err := tools.GetCurrentMoscowTime()
	// if err != nil {
	// 	return nil, karma.Format(
	// 		err,
	// 		"unable to get moscow time",
	// 	)
	// }

	return nil
}
