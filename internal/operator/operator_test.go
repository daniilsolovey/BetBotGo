package operator

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/daniilsolovey/BetBotGo/internal/constants"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/tools"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/assert"
	"github.com/reconquest/pkg/log"
)

const (
	TEST_EVENTS_PATH  = "../../testdata/upcoming_events.json"
	TEST_EVENT_1_PATH = "../../testdata/1_event.json"
	TEST_EVENT_2_PATH = "../../testdata/2_event.json"
	TEST_EVENT_3_PATH = "../../testdata/3_event.json"
)

type TestRequester struct {
}

func createRequester() *TestRequester {
	return &TestRequester{}
}

func getTestData(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte{}, karma.Format(
			err,
			"unable to get data from the path: %s", path)
	}
	return data, nil
}

func getEventsFromTestPath() (events *requester.UpcomingEvents, err error) {
	path := TEST_EVENTS_PATH
	data, err := getTestData(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func getOneEventFromTestPath(path string) (*requester.EventWithOdds, error) {
	data, err := getTestData(path)
	if err != nil {
		return nil, err
	}

	var eventWithOdds requester.EventWithOdds

	err = json.Unmarshal(data, &eventWithOdds)
	if err != nil {
		return nil, err
	}

	return &eventWithOdds, nil
}

func (testRequester *TestRequester) GetUpcomingEvents() (*requester.UpcomingEvents, error) {
	events, err := getEventsFromTestPath()
	if err != nil {
		log.Fatal(err)
	}

	return events, nil
}

func (testRequester *TestRequester) GetEventOddsByEventIDs(events *requester.UpcomingEvents) ([]requester.EventWithOdds, error) {
	var result []requester.EventWithOdds
	path := TEST_EVENT_1_PATH
	event1, err := getOneEventFromTestPath(path)
	if err != nil {
		log.Fatal(err)
	}

	path = TEST_EVENT_2_PATH
	event2, err := getOneEventFromTestPath(path)
	if err != nil {
		log.Fatal(err)
	}

	path = TEST_EVENT_3_PATH
	event3, err := getOneEventFromTestPath(path)
	if err != nil {
		log.Fatal(err)
	}

	event1.HumanTime = events.Results[0].HumanTime
	event2.HumanTime = events.Results[1].HumanTime
	event3.HumanTime = events.Results[2].HumanTime
	event1.EventID = events.Results[0].ID
	event2.EventID = events.Results[1].ID
	event3.EventID = events.Results[2].ID

	result = append(result, *event1, *event2, *event3)
	return result, nil
}

func (testRequester *TestRequester) GetLiveEventByID(eventID string) (*requester.EventWithOdds, error) {
	upcomingEvents, err := testRequester.GetUpcomingEvents()
	if err != nil {
		log.Fatal(err)
	}

	upcomingEventsForToday, err := getUpcomingEventsForToday(upcomingEvents)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get upcoming events for today",
		)
	}

	events, err := testRequester.GetEventOddsByEventIDs(upcomingEventsForToday)
	if err != nil {
		log.Fatal(err)
	}
	for _, event := range events {
		log.Warning("event.Id ", event.EventID)
		if event.EventID == eventID {
			return &event, nil
		}
	}

	return nil, nil
}

func TestOperator_GetEvents_ReturnWinnerResult(
	t *testing.T,
) {
	tools.TimeNow = func() time.Time {
		return time.Date(2021, 9, 03, 16, 0, 0, 0, time.UTC)
	}

	requester := createRequester()

	operator := NewOperator(nil, nil, requester)

	events, err := operator.GetEvents()
	if err != nil {
		log.Fatal(err)
	}
	assert.NotNil(events, "")
}

func TestOperator_getWinner_ReturnWinnerResult(
	t *testing.T,
) {
	tools.TimeNow = func() time.Time {
		return time.Date(2021, 9, 03, 16, 0, 0, 0, time.UTC)
	}

	testRequester := createRequester()
	operator := NewOperator(nil, nil, testRequester)

	upcomingEvents, err := testRequester.GetUpcomingEvents()
	if err != nil {
		log.Fatal(err)
	}

	eventsWithOdds, err := testRequester.GetEventOddsByEventIDs(upcomingEvents)
	if err != nil {
		log.Fatal(err)
	}
	eventsWithOdds[1].EventID = "4009234"
	// eventsWithOdds[1].ResultEventWithOdds.Odds.Odds91_1[0].SS = "17-25,25-21"
	// eventsWithOdds[1].ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.67"
	eventsWithOdds[1].Favorite = constants.FAVORITE_IS_HOME
	result := operator.getWinner(&eventsWithOdds[1])
	log.Warning("result ", result)
}
