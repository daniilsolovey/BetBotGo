package operator

import (
	"fmt"

	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
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
	fmt.Println("upcomingEvents", upcomingEventsForToday)

	eventsWithOdds, err := operator.requester.GetEventOddsByEventIDs(upcomingEventsForToday)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to get upcoming events",
		)
	}
	fmt.Println("eventsWithOdds", eventsWithOdds)

	sortedEventsWithOdds, err := sortEventsByOdds(eventsWithOdds)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to sort events by odds",
		)
	}

	fmt.Println("sortedEvents", sortedEventsWithOdds)
	fmt.Println("len(eventsWithOdds)", len(eventsWithOdds))
	fmt.Println("len(sortedEvents)", len(sortedEventsWithOdds))

	return sortedEventsWithOdds, nil
}

func (operator *Operator) HandleLiveEvents([]requester.EventWithOdds) error {
	return nil
}

// def scan_live_event(self):
// print("in worker func")
// current_time = datetime.datetime.now()
// self_primary_odds = self.event['results']['odds'].get('91_1')
// print("self_primary_odds",self_primary_odds)
// event_add_time = datetime.datetime.fromtimestamp(
// 	int(self_primary_odds[0]['add_time']))
// response_json = self.requester.get_event_data_by_id(
// 	str(self.url), str(self.event_id))
// primary_odds = response_json['results']['odds'].get('91_1')
// play_set = primary_odds[0]['ss']
// winner = get_winner_of_first_set_of_game(play_set)
// if winner != self.favorite and \
// 	get_number_of_set(play_set) == 2 and \
// 		float(primary_odds[0][self.favorite]) > 1.5:
// 	return response_json
