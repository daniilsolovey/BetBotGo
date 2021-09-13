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
	REQUEST_FREQUENCY_DELAY                   = 5 * time.Second
)

type Operator struct {
	config                     *config.Config
	database                   *database.Database
	requester                  requester.RequesterInterface
	transport                  transport.Transport
	RoutineCache               []string
	allEventsOnCurrentDayCache []string
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

func (operator *Operator) AddEventsIDsOnCurrentDayToCache(events []requester.EventWithOdds) {
	for _, event := range events {
		if len(operator.allEventsOnCurrentDayCache) != 0 &&
			!operator.IsAllEventsCacheContainsEvent(event.EventID) {
			operator.allEventsOnCurrentDayCache = append(
				operator.allEventsOnCurrentDayCache,
				event.EventID,
			)
		}
	}
}

func (operator *Operator) AddEventsIDsAboutCreatedRoutines(events []requester.EventWithOdds) {
	for _, event := range events {
		if len(operator.RoutineCache) != 0 &&
			!operator.IsRoutineCacheContainsEvent(event.EventID) {
			operator.RoutineCache = append(
				operator.RoutineCache,
				event.EventID,
			)
		}
	}
}

func (operator *Operator) IsRoutineCacheContainsEvent(eventID string) bool {
	for _, value := range operator.RoutineCache {
		if eventID == value {
			return true
		}
	}

	return false
}

func (operator *Operator) IsAllEventsCacheContainsEvent(eventID string) bool {
	for _, value := range operator.allEventsOnCurrentDayCache {
		if eventID == value {
			return true
		}
	}

	return false
}

func (operator *Operator) GetEvents() ([]requester.EventWithOdds, error) {
	log.Info("receiving events for today")
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

func (operator *Operator) SendMessageAboutWinnerToTelegram(event requester.EventWithOdds) error {
	text := fmt.Sprintf(
		TEXT_ABOUT_WINNER,
		event.EventID,
		event.League.Name,
		event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd,
		event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd,
		event.HomeCommandName,
		event.AwayCommandName,
		event.Favorite,
	)

	err := operator.transport.SendMessage(TEMP_RECIPIENT, text)
	if err != nil {
		return err
	}

	return nil
}

func (operator *Operator) CreateRoutinesForHandleLiveEvents(events []requester.EventWithOdds) error {
	if len(events) == 0 {
		return nil
	}

	log.Info("creating routines for each event")

	timeNow, err := tools.GetCurrentMoscowTime()
	if err != nil {
		return karma.Format(
			err,
			"unable to get moscow time for check that event ready to start in go-routine",
		)
	}

	for _, event := range events {
		if event.EventStartTime.After(timeNow) ||
			event.EventStartTime.Before(event.EventStartTime.Add(MONITORING_LIVE_EVENT_TIME_DELAY)) &&
				!operator.IsRoutineCacheContainsEvent(event.EventID) {
			go operator.routineStartHandleLiveOdds(event)
			eventForCache := []requester.EventWithOdds{event}
			operator.AddEventsIDsAboutCreatedRoutines(eventForCache)
		}
	}

	return nil
}

func (operator *Operator) routineFinalHandleLiveOdds(event requester.EventWithOdds) {
	liveEvent, secondSetIsFinished := operator.createHandlerFinalOdds(event)
	if secondSetIsFinished {
		setData := liveEvent.ResultEventWithOdds.Odds.Odds91_1[0].SS
		winner := getWinnerInSecondSet(setData)
		liveEvent.WinnerInSecondSet = winner
		//write to database result of second set
		err := operator.database.UpdateLiveEventsResultsScoreAndWinnerFields(liveEvent)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof(nil, "live event successfully updated in database, live_event: %v", liveEvent)
		}
	}
}

func (operator *Operator) routineStartHandleLiveOdds(event requester.EventWithOdds) error {
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
		log.Warningf(nil, "waiting time for routine: %s ", diff.String())
		time.Sleep(diff)
	}

	liveEvent, liveEventResult := operator.createHandlerLiveOdds(event)
	if liveEventResult {
		err := operator.SendMessageAboutWinnerToTelegram(liveEvent)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof(nil, "live event sent to telegram, event: %v", liveEvent)
		}

		err = operator.database.InsertLiveEventResult(liveEvent)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof(nil, "live event inserted to database, event: %v", liveEvent)
		}

		go operator.routineFinalHandleLiveOdds(liveEvent)
	}

	log.Infof(nil, "routine successfully finished for event_id: %s", event.EventID)
	return nil
}

func (operator *Operator) createHandlerLiveOdds(event requester.EventWithOdds) (requester.EventWithOdds, bool) {
	startTime, err := tools.GetCurrentMoscowTime()
	if err != nil {
		log.Error(err)
	}

	log.Infof(nil, "routine for receivnig winner started, start_time: %s, event: %v", startTime.String(), event)

	for {
		timeNow, err := tools.GetCurrentMoscowTime()
		if err != nil {
			log.Error(err)
		}

		if timeNow.After(startTime.Add(SELF_DISTRUCT_ROUTINE_LIVE_EVENT_TIMER)) {
			log.Infof(nil, "routine for receivnig winner stopped by timeout for event_id: %s", event.EventID)
			return requester.EventWithOdds{}, false
		}

		log.Infof(nil, "handle live odds for event: %v", event)
		liveEvent, err := operator.requester.GetLiveEventByID(event.EventID)
		if err != nil {
			log.Errorf(err, "unable to get live event data by event_id: %s", event.EventID)
		}

		liveEventResult := operator.getWinner(*liveEvent, 0)
		switch liveEventResult {
		case "finished with error":
			return *liveEvent, false
		case "numberOfSet=3":
			return *liveEvent, false
		case "true":
			return *liveEvent, true
		case "":
			time.Sleep(REQUEST_FREQUENCY_DELAY)
			continue
		}
	}
}

func (operator *Operator) createHandlerFinalOdds(event requester.EventWithOdds) (requester.EventWithOdds, bool) {
	startTime, err := tools.GetCurrentMoscowTime()
	if err != nil {
		log.Error(err)
	}

	log.Infof(nil, "routine for second final set started, start_time: %s, event: %v", startTime.String(), event)
	for {
		timeNow, err := tools.GetCurrentMoscowTime()
		if err != nil {
			log.Error(err)
		}

		if timeNow.After(startTime.Add(SELF_DISTRUCT_ROUTINE_LIVE_EVENT_TIMER)) {
			log.Infof(nil, "routine for second set stopped by timeout for event: %s", event.EventID)
			return requester.EventWithOdds{}, false
		}

		log.Infof(nil, "handle live odds for event_id: %s", event.EventID)
		liveEvent, err := operator.requester.GetLiveEventByID(event.EventID)
		if err != nil {
			log.Errorf(err, "unable to get live event data by event_id: %s", event.EventID)
		}

		liveEventResult := operator.getResultsOfSecondSet(*liveEvent, 0)
		switch liveEventResult {
		case "finished with error":
			return *liveEvent, false
		case "numberOfSet=3":
			return *liveEvent, true
		case "":
			time.Sleep(REQUEST_FREQUENCY_DELAY)
			continue
		}
	}
}

func (operator *Operator) getResultsOfSecondSet(liveEvent requester.EventWithOdds, errCount int) string {
	// log.Infof(nil, "handle live odds for event: %v", event)
	// liveEvent, err := operator.requester.GetLiveEventByID(event.EventID)
	// if err != nil {
	// 	if errCount < MAX_ERROR_COUNT_IN_MONITORING_LIVE_EVENTS {
	// 		log.Errorf(err, "unable to get live event data by event_id: %s", event.EventID)
	// 		errCount = +1
	// 		time.Sleep(2 * time.Second)
	// 	} else {
	// 		// break
	// 		return "finished with error"
	// 	}
	// }

	_, numberOfSet, err := handleLiveEventOdds(&liveEvent)
	if err != nil {
		if errCount < MAX_ERROR_COUNT_IN_MONITORING_LIVE_EVENTS {
			log.Errorf(err, "unable to handle live event event_id: %s", liveEvent.EventID)
			errCount = +1
			time.Sleep(2 * time.Second)
		} else {
			return "finished with error"
		}
	}

	if numberOfSet == 3 {
		return "numberOfSet=3"
	}

	return ""
}

func (operator *Operator) getWinner(liveEvent requester.EventWithOdds, errCount int) string {
	// log.Infof(nil, "handle live odds for event: %v", event)
	// liveEvent, err := operator.requester.GetLiveEventByID(event.EventID)
	// if err != nil {
	// 	if errCount < MAX_ERROR_COUNT_IN_MONITORING_LIVE_EVENTS {
	// 		log.Errorf(err, "unable to get live event data by event_id: %s", event.EventID)
	// 		errCount = +1
	// 		time.Sleep(2 * time.Second)
	// 	} else {
	// 		// break
	// 		return "finished with error"
	// 	}
	// }

	liveEventResult, numberOfSet, err := handleLiveEventOdds(&liveEvent)
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
