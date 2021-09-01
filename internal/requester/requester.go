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

func (requester *Requester) GetUpcomingEventsOnCurrentDate() (*UpcomingEvents, error) {
	log.Info("receiving upcoming events")

	url := requester.config.BetApi.BaseUrlUpcomingEvents + requester.config.BetApi.Token
	log.Warning("url ", url)
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

	log.Warning("baseUrl ", baseURL)
	var result []EventWithOdds
	for _, event := range events.Results {
		url := baseURL + event.ID
		log.Infof(
			karma.Describe("event_id", event.ID).Describe("url", url),
			"receiving odds for event",
		)

		log.Warning("request before ")
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to get request by url: %s", url,
			)
		}
		log.Warning("request after ")

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to send http request by url: %s", url,
			)
		}
		log.Warning("request after 2 ")

		// defer response.Body.Close()

		var eventWithOdds EventWithOdds
		err = json.NewDecoder(response.Body).Decode(&eventWithOdds)
		log.Warning("request after 222222 ")
		log.Warning("err ", err)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to decode response, response status code: %d ",
				response.StatusCode,
			)
		}

		eventWithOdds.League = event.League
		eventWithOdds.HumanTime = event.HumanTime
		log.Warning("eventWithOdds.HumanTime ", eventWithOdds.HumanTime)
		log.Warning("result after 3 ")

		result = append(result, eventWithOdds)
		log.Warning("result after 4 ")

		response.Body.Close()

	}

	return result, nil
}

// bodyBytes, err := ioutil.ReadAll(response.Body)
// if err != nil {
// 	log.Fatal(err)
// }
// log.Warning("string(bodyBytes) ", string(bodyBytes))
