package operator

import (
	"fmt"
	"time"

	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/tools"
	"github.com/daniilsolovey/BetBotGo/internal/transport"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
)

const (
	MONITORING_LIVE_EVENT_TIME_DELAY          = 30 * time.Minute
	MAX_ERROR_COUNT_IN_MONITORING_LIVE_EVENTS = 100
	SELF_DISTRUCT_ROUTINE_LIVE_EVENT_TIMER    = 2 * time.Hour
)

type Operator struct {
	config    *config.Config
	database  *database.Database
	requester requester.RequesterInterface
	transport transport.Transport
}

func NewOperator(
	config *config.Config,
	database *database.Database,
	requester requester.RequesterInterface,
	transport transport.Transport,
) *Operator {
	return &Operator{
		config:    config,
		database:  database,
		requester: requester,
		transport: transport,
	}
}

func (operator *Operator) GetEvents() ([]requester.EventWithOdds, error) {
	upcomingEvents, err := operator.requester.GetUpcomingEvents()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get upcoming events",
		)
	}
	log.Info("handle upcoming events and receiving events for today")
	if upcomingEvents == nil {
		return nil, nil
	}

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

func (operator *Operator) RoutineHandleLiveOdds(event requester.EventWithOdds) error {
	log.Infof(nil, "creating routine for event_id: %s", event.EventID)
	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return karma.Format(
			err,
			"unable to get moscow time",
		)
	}

	if timeNow.Before(event.EventStartTime) {
		diff := event.EventStartTime.Sub(timeNow)
		time.Sleep(diff)
	}

	liveEventResult := operator.createLoopForGetWinner(event)
	if liveEventResult {
		err := operator.SendMessageAboutWinnerToTelegram(event)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof(nil, "live event sent to telegram, event: %v", event)
		}

		err = operator.database.InsertLiveEventResult(event)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof(nil, "live event inserted to database, event: %v", event)
		}
	}

	log.Infof(nil, "routine successfully finished for event_id: %s", event.EventID)
	return nil
}

func (operator *Operator) SendMessageAboutWinnerToTelegram(event requester.EventWithOdds) error {
	text := fmt.Sprintf(
		TEXT_ABOUT_WINNER,
		event.EventID,
		event.League.Name,
		event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd,
		event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd,
		event.Favorite,
	)

	err := operator.transport.SendMessage(TEMP_RECIPIENT, text)
	if err != nil {
		return err
	}

	return nil
}

func (operator *Operator) createLoopForGetWinner(event requester.EventWithOdds) bool {
	timer := time.NewTimer(SELF_DISTRUCT_ROUTINE_LIVE_EVENT_TIMER)
	<-timer.C
	log.Info("timer routine started")
	for {
		if timer.Stop() {
			log.Info("timer routine stopped")
			return false
		}

		liveEventResult := operator.getWinner(event, 0)
		switch liveEventResult {
		case "finished with error":
			return false
		case "numberOfSet=3":
			return false
		case "true":
			return true
		case "":
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func (operator *Operator) getWinner(event requester.EventWithOdds, errCount int) string {
	log.Infof(nil, "handle live odds for event: %v", event)
	liveEvent, err := operator.requester.GetLiveEventByID(event.EventID)
	if err != nil {
		if errCount < MAX_ERROR_COUNT_IN_MONITORING_LIVE_EVENTS {
			log.Errorf(err, "unable to get live event data by event_id: %s", event.EventID)
			errCount = +1
			time.Sleep(2 * time.Second)
		} else {
			// break
			return "finished with error"
		}
	}

	liveEventResult, numberOfSet, err := handleLiveEventOdds(liveEvent)
	if err != nil {
		if errCount < MAX_ERROR_COUNT_IN_MONITORING_LIVE_EVENTS {
			log.Errorf(err, "unable to handle live event event_id: %s", liveEvent.EventID)
			errCount = +1
			time.Sleep(2 * time.Second)
		} else {
			return "finished with error"
		}
	}

	if liveEventResult {
		return "true"

	}

	if numberOfSet == 3 {
		return "numberOfSet=3"
	}

	return ""
}

func (operator *Operator) CreateRoutinesForEachEvent(events []requester.EventWithOdds) error {
	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return karma.Format(
			err,
			"unable to get moscow time for check that event ready to start in go-routine",
		)
	}

	for _, event := range events {
		if event.EventStartTime.After(timeNow) || event.EventStartTime.Before(event.EventStartTime.Add(MONITORING_LIVE_EVENT_TIME_DELAY)) {
			go operator.RoutineHandleLiveOdds(event)
		}
	}

	return nil
}

// func (operator *Operator) CreateRoutinesForEachEvent(events []requester.EventWithOdds) error {
// 	timeNow, err := tools.GetCurrentMoscowTime()
// 	if err != nil {
// 		return karma.Format(
// 			err,
// 			"unable to get moscow time for check that event ready to start in go-routine",
// 		)
// 	}

// 	for _, event := range events {
// 		if event.HumanTime.After(timeNow) || event.HumanTime.Before(timeNow) && event.HumanTime.Before(timeNow.Add(30*time.Minute)) {
// 			go operator.RoutineHandleLiveOdds(event)
// 		}

// 		// if event.HumanTime.Before(timeNow) && event.HumanTime.Before(timeNow.Add(60*time.Minute)) {
// 		// 	go operator.RoutineHandleLiveOdds(event)
// 		// }

// 	}

// 	return nil
// }
