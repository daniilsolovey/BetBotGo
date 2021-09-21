package operator

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/tools"
	"github.com/reconquest/karma-go"
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

	event1.EventStartTime = events.Results[0].HumanTime
	event2.EventStartTime = events.Results[1].HumanTime
	event3.EventStartTime = events.Results[2].HumanTime
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

	operator := NewOperator(nil, nil, requester, nil)

	events, err := operator.GetEvents()
	if err != nil {
		log.Fatal(err)
	}

	assert.NotNil(t, events, "")
}

func TestOperator_handleLiveEventOdds_ReturnWinnerResult(
	t *testing.T,
) {
	tools.TimeNow = func() time.Time {
		return time.Date(2021, 9, 03, 16, 0, 0, 0, time.UTC)
	}

	// testRequester := createRequester()
	// operator := NewOperator(nil, nil, testRequester, nil)

	event := requester.EventWithOdds{
		EventID:             "1111",
		Favorite:            "away",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.533"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "27-11,1-3"

	result, numberOfset, err := handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, true)
	assert.Equal(t, numberOfset, 2)

	event = requester.EventWithOdds{
		EventID:             "2222",
		Favorite:            "home",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.533"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "27-11,1-3"

	result, numberOfset, err = handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, false)
	assert.Equal(t, numberOfset, 2)

	event = requester.EventWithOdds{
		EventID:             "3333",
		Favorite:            "home",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.433"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "27-11,1-3"

	result, numberOfset, err = handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, false)
	assert.Equal(t, numberOfset, 2)

	event = requester.EventWithOdds{
		EventID:             "3333",
		Favorite:            "home",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.52"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "27-11,1-3"

	result, numberOfset, err = handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, false)
	assert.Equal(t, numberOfset, 2)

	event = requester.EventWithOdds{
		EventID:             "4444",
		Favorite:            "home",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.52"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "27-31,1-3"

	result, numberOfset, err = handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, true)
	assert.Equal(t, numberOfset, 2)

	event = requester.EventWithOdds{
		EventID:             "5555",
		Favorite:            "away",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.52"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "27-31,1-3"

	result, numberOfset, err = handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, false)
	assert.Equal(t, numberOfset, 2)

	event = requester.EventWithOdds{
		EventID:             "6666",
		Favorite:            "away",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.52"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "30-21,1-3"

	result, numberOfset, err = handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, true)
	assert.Equal(t, numberOfset, 2)

	event = requester.EventWithOdds{
		EventID:             "7777",
		Favorite:            "home",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.52"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "25-18,0-0"

	result, numberOfset, err = handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, false)
	assert.Equal(t, numberOfset, 2)

	event = requester.EventWithOdds{
		EventID:             "8888",
		Favorite:            "away",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.025"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "10.500"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "17-25,0-0"

	result, numberOfset, err = handleLiveEventOdds(event)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, result, false)
	assert.Equal(t, numberOfset, 2)

}
func TestOperator_getWinnerInSecondSet_ReturnWinner(
	t *testing.T,
) {
	tools.TimeNow = func() time.Time {
		return time.Date(2021, 9, 03, 16, 0, 0, 0, time.UTC)
	}

	event := requester.EventWithOdds{
		EventID:             "1111",
		Favorite:            "away",
		ResultEventWithOdds: requester.ResultEventWithOdds{Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}}},
	}

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.533"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "27-11,1-3,0-0"
	setData := event.ResultEventWithOdds.Odds.Odds91_1[0].SS

	result := getWinnerInSecondSet(setData)

	assert.Equal(t, result, "away")

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.533"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "2"
	event.ResultEventWithOdds.Odds.Odds91_1[0].SS = "27-11,9-3,0-0"
	setData = event.ResultEventWithOdds.Odds.Odds91_1[0].SS

	result = getWinnerInSecondSet(setData)

	assert.Equal(t, result, "home")
}

func TestOperator_sortEventsByOdds_ReturnFavoriteIsHome(
	t *testing.T,
) {
	tools.TimeNow = func() time.Time {
		return time.Date(2021, 9, 03, 16, 0, 0, 0, time.UTC)
	}

	event := requester.EventWithOdds{
		EventID: "1111",
		ResultEventWithOdds: requester.ResultEventWithOdds{
			Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}},
		},
	}
	var events []requester.EventWithOdds

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.233"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "3"

	events = append(events, event)

	result, err := sortEventsByOdds(events)
	assert.NotNil(t, result)
	assert.NoError(t, err)
	for _, item := range result {
		assert.Equal(t, item.Favorite, "home")
	}
}

func TestOperator_sortEventsByOdds_ReturnFavoriteIsAway(
	t *testing.T,
) {
	tools.TimeNow = func() time.Time {
		return time.Date(2021, 9, 03, 16, 0, 0, 0, time.UTC)
	}

	event := requester.EventWithOdds{
		EventID: "1111",
		ResultEventWithOdds: requester.ResultEventWithOdds{
			Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}},
		},
	}
	var events []requester.EventWithOdds

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "1.3"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "1.233"

	events = append(events, event)

	result, err := sortEventsByOdds(events)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	for _, item := range result {
		assert.Equal(t, item.Favorite, "away")
	}
}

func TestOperator_sortEventsByOdds_MissEvent(
	t *testing.T,
) {
	tools.TimeNow = func() time.Time {
		return time.Date(2021, 9, 03, 16, 0, 0, 0, time.UTC)
	}

	event := requester.EventWithOdds{
		EventID: "1111",
		ResultEventWithOdds: requester.ResultEventWithOdds{
			Odds: requester.Odds{Odds91_1: []requester.OddsNumber{requester.OddsNumber{}}},
		},
	}
	var events []requester.EventWithOdds

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "7"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "1.33"

	events = append(events, event)

	result, err := sortEventsByOdds(events)
	assert.NoError(t, err)
	assert.Nil(t, result)

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "7"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "1.31"

	events = append(events, event)
	result, err = sortEventsByOdds(events)
	assert.NoError(t, err)
	assert.Nil(t, result)

	event.ResultEventWithOdds.Odds.Odds91_1[0].HomeOd = "7"
	event.ResultEventWithOdds.Odds.Odds91_1[0].AwayOd = "1.30"

	events = append(events, event)
	result, err = sortEventsByOdds(events)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
