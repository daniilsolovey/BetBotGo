package requester

import (
	"encoding/json"
	"net/http"

	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
)

const (
	BASE_URL_EVENT_ID = "&event_id="
)

type RequesterInterface interface {
	GetUpcomingEvents() (*UpcomingEvents, error)
	GetEventOddsByEventIDs(*UpcomingEvents) ([]EventWithOdds, error)
	GetLiveEventByID(string) (*EventWithOdds, error)
}

type Requester struct {
	config *config.Config
}

func NewRequester(
	config *config.Config,
) *Requester {
	return &Requester{
		config: config,
	}
}

func (requester *Requester) GetUpcomingEvents() (*UpcomingEvents, error) {
	log.Info("receiving upcoming events")
	url := requester.config.BetApi.BaseUrlUpcomingEvents + requester.config.BetApi.Token
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get request by url: %s", url,
		)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to send http request by url: %s", url,
		)
	}

	defer response.Body.Close()

	var upcomingEvents UpcomingEvents
	err = json.NewDecoder(response.Body).Decode(&upcomingEvents)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to decode response, response status code: %d ",
			response.StatusCode,
		)
	}
	return &upcomingEvents, nil
}

func (requester *Requester) GetEventOddsByEventIDs(events *UpcomingEvents) ([]EventWithOdds, error) {
	log.Info("receiving event odds by event ids")
	baseURL := requester.config.BetApi.BaseUrlGetEventOddsById + requester.config.BetApi.Token +
		BASE_URL_EVENT_ID

	var result []EventWithOdds
	for _, event := range events.Results {
		url := baseURL + event.ID
		log.Infof(
			karma.Describe("event_id", event.ID).Describe("url", url),
			"receiving odds for event",
		)

		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to get request by url: %s", url,
			)
		}

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to send http request by url: %s", url,
			)
		}

		// defer response.Body.Close()
		var eventWithOdds EventWithOdds
		err = json.NewDecoder(response.Body).Decode(&eventWithOdds)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to decode response, response status code: %d ",
				response.StatusCode,
			)
		}

		eventWithOdds.League = event.League
		eventWithOdds.HumanTime = event.HumanTime
		eventWithOdds.EventID = event.ID
		eventWithOdds.HomeCommandName = event.Home.Name
		eventWithOdds.AwayCommandName = event.Away.Name

		result = append(result, eventWithOdds)
		response.Body.Close()
	}

	return result, nil
}

func (requester *Requester) GetLiveEventByID(eventID string) (*EventWithOdds, error) {
	baseURL := requester.config.BetApi.BaseUrlGetEventOddsById + requester.config.BetApi.Token +
		BASE_URL_EVENT_ID

	url := baseURL + eventID
	log.Infof(
		karma.Describe("event_id", eventID).Describe("url", url),
		"receiving odds for event in live match",
	)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get request by url: %s", url,
		)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to send http request by url: %s", url,
		)
	}

	defer response.Body.Close()
	var eventWithOdds EventWithOdds
	err = json.NewDecoder(response.Body).Decode(&eventWithOdds)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to decode response, response status code: %d ",
			response.StatusCode,
		)
	}

	return &eventWithOdds, nil
}

// bodyBytes, err := ioutil.ReadAll(response.Body)
// if err != nil {
// 	log.Fatal(err)
// }
// log.Warning("string(bodyBytes) ", string(bodyBytes))
