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
	MAX_ERROR_COUNT_IN_MONITORING_LIVE_EVENTS = 1000
	SELF_DISTRUCT_ROUTINE_LIVE_EVENT_TIMER    = 3 * time.Hour
	REQUEST_FREQUENCY_DELAY                   = 7 * time.Second

	CODE_FINISHED_WITH_ERROR = "finished with error"
	CODE_NUMBER_OF_SET_3     = "numberOfSet=3"
	CODE_IS_WINNER_TRUE      = "true"
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
		if !operator.IsAllEventsCacheContainsEvent(event.EventID) {
			operator.allEventsOnCurrentDayCache = append(
				operator.allEventsOnCurrentDayCache,
				event.EventID,
			)
		}
	}
}

func (operator *Operator) AddEventsIDsAboutCreatedRoutines(events []requester.EventWithOdds) {
	for _, event := range events {
		if !operator.IsRoutineCacheContainsEvent(event.EventID) {
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
func (operator *Operator) HandleEventsByLeagues(events []requester.EventWithOdds) []requester.EventWithOdds {
	return handleEventsByLeagues(events)
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

	return handleEventsByCountries(sortedEventsWithOdds), nil
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
			event.EventStartTime.Before(event.EventStartTime.Add(MONITORING_LIVE_EVENT_TIME_DELAY)) {
			if !operator.IsRoutineCacheContainsEvent(event.EventID) {
				go operator.routineStartHandleLiveOdds(event)
				eventForCache := []requester.EventWithOdds{event}
				operator.AddEventsIDsAboutCreatedRoutines(eventForCache)
			} else {
				log.Infof(nil, "routine for event created before, event_id: %s", event.EventID)
			}
		}
	}

	return nil
}

func (operator *Operator) routineFinalHandleLiveOdds(event requester.EventWithOdds) {
	liveEvent, secondSetIsFinished := operator.createHandlerFinalOdds(event)
	if secondSetIsFinished {
		setData := liveEvent.ResultEventWithOdds.Odds.Odds91_1[0].SS
		winner := getWinnerInSecondSet(setData)
		log.Infof(nil, "final set data: %s", setData)
		log.Infof(nil, "winner: %s", winner)
		//write to database result of second set
		err := operator.database.UpdateLiveEventsResultsScoreAndWinnerFields(event.EventID, setData, winner)
		if err != nil {
			log.Errorf(err, "unable to update live events results score and winner fields")
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
		// liveEvent.EventID = event.EventID
		// liveEvent.League.Name = event.League.Name
		// liveEvent.HomeCommandName = event.HomeCommandName
		// liveEvent.AwayCommandName = event.AwayCommandName
		// liveEvent.Favorite = event.Favorite
		err := operator.SendMessageAboutWinnerToTelegram(*liveEvent)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof(nil, "live event sent to telegram, event_id: %s", liveEvent.EventID)
		}

		err = operator.database.InsertLiveEventResult(*liveEvent)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof(nil, "live event inserted to database, event_id: %s", liveEvent.EventID)
		}

		go operator.routineFinalHandleLiveOdds(*liveEvent)
	}

	log.Infof(nil, "routine successfully finished for event_id: %s", event.EventID)
	return nil
}

func (operator *Operator) createHandlerLiveOdds(event requester.EventWithOdds) (*requester.EventWithOdds, bool) {
	startTime, err := tools.GetCurrentMoscowTime()
	if err != nil {
		log.Error(err)
		return nil, false
	}

	log.Infof(nil, "routine for receiving winner started, start_time: %s, event: %v", startTime.String(), event)

	for {
		timeNow, err := tools.GetCurrentMoscowTime()
		if err != nil {
			log.Error(err)
			time.Sleep(REQUEST_FREQUENCY_DELAY)
			continue
		}

		if timeNow.After(startTime.Add(SELF_DISTRUCT_ROUTINE_LIVE_EVENT_TIMER)) {
			log.Infof(nil, "routine for receivnig winner stopped by timeout for event_id: %s", event.EventID)
			return nil, false
		}

		liveEvent, err := operator.requester.GetLiveEventByID(event.EventID)
		if err != nil {
			log.Errorf(err, "unable to get live event data by event_id: %s", event.EventID)
			time.Sleep(REQUEST_FREQUENCY_DELAY)
			continue
		}

		liveEvent.EventID = event.EventID
		liveEvent.League.Name = event.League.Name
		liveEvent.HomeCommandName = event.HomeCommandName
		liveEvent.AwayCommandName = event.AwayCommandName
		liveEvent.Favorite = event.Favorite

		log.Infof(nil, "handle live odds for event_id: %s", liveEvent.EventID)

		liveEventResult := operator.getWinnerOfSecondSet(*liveEvent)
		switch liveEventResult {
		case CODE_FINISHED_WITH_ERROR:
			return liveEvent, false
		case CODE_NUMBER_OF_SET_3:
			return liveEvent, false
		case CODE_IS_WINNER_TRUE:
			return liveEvent, true
		case "":
			time.Sleep(REQUEST_FREQUENCY_DELAY)
			continue
		}
	}
}

func (operator *Operator) createHandlerFinalOdds(event requester.EventWithOdds) (*requester.EventWithOdds, bool) {
	startTime, err := tools.GetCurrentMoscowTime()
	if err != nil {
		log.Error(err)
	}

	log.Infof(nil, "routine for second final set started, start_time: %s, event: %v", startTime.String(), event)
	for {
		timeNow, err := tools.GetCurrentMoscowTime()
		if err != nil {
			log.Error(err)
			time.Sleep(REQUEST_FREQUENCY_DELAY)
			continue
		}

		if timeNow.After(startTime.Add(SELF_DISTRUCT_ROUTINE_LIVE_EVENT_TIMER)) {
			log.Infof(nil, "routine for handle final odds stopped by timeout for event: %s", event.EventID)
			return nil, false
		}

		log.Info("receiving odds for final set")
		liveEvent, err := operator.requester.GetLiveEventByID(event.EventID)
		if err != nil {
			log.Errorf(err, "unable to get live event data by event_id: %s", event.EventID)
			time.Sleep(REQUEST_FREQUENCY_DELAY)
			continue
		}

		log.Infof(nil, "handle final live odds for event_id: %s", liveEvent.EventID)
		liveEvent.EventID = event.EventID
		liveEvent.League.Name = event.League.Name
		liveEvent.HomeCommandName = event.HomeCommandName
		liveEvent.AwayCommandName = event.AwayCommandName
		liveEvent.Favorite = event.Favorite

		liveEventResult := operator.getFinalResultsOfSecondSet(*liveEvent)
		switch liveEventResult {
		case CODE_NUMBER_OF_SET_3:
			return liveEvent, true
		case "":
			time.Sleep(REQUEST_FREQUENCY_DELAY)
			continue
		}
	}
}

func (operator *Operator) getFinalResultsOfSecondSet(liveEvent requester.EventWithOdds) string {
	numberOfSet, err := handleFinalLiveSet(liveEvent)
	if err != nil {
		log.Errorf(err, "unable to handle final live results of second set event_id: %s", liveEvent.EventID)
		return ""
	}

	if numberOfSet == 3 {
		return CODE_NUMBER_OF_SET_3
	}

	return ""
}

func (operator *Operator) getWinnerOfSecondSet(liveEvent requester.EventWithOdds) string {
	liveEventResult, numberOfSet, err := handleLiveEventOdds(liveEvent)
	if err != nil {
		log.Errorf(err, "unable to handle live event and receive winner event_id: %s", liveEvent.EventID)
		time.Sleep(2 * REQUEST_FREQUENCY_DELAY)
		return ""
	}

	if liveEventResult {
		return CODE_IS_WINNER_TRUE

	}

	if numberOfSet == 3 {
		return CODE_NUMBER_OF_SET_3
	}

	return ""
}
