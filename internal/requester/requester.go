package requester

import (
	"encoding/json"
	"net/http"

	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
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

type UpcomingEvents struct {
	Results []Result `json:"results"`
}

type Result struct {
	ID      string `json:"id"`
	SportID string `json:"sport_id"`
	Time    string `json:"time"`
	League  League `json:"league"`
	Home    Home   `json:"home"`
	Away    Away   `json:"away"`
}

type League struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CC   string `json:"cc"`
}

type Home struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CC   string `json:"cc"`
}

type Away struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CC   string `json:"cc"`
}

func (requester *Requester) GetUpcomingEventsOnCurrentDate() (*UpcomingEvents, error) {
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

	resEvents, err := getUpcomingEventsForToday(&upcomingEvents)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get upcoming events on current date",
		)
	}

	return resEvents, nil
}

// bodyBytes, err := ioutil.ReadAll(response.Body)
// if err != nil {
// 	log.Fatal(err)
// }
// log.Warning("string(bodyBytes) ", string(bodyBytes))
