package statistics

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/daniilsolovey/BetBotGo/internal/requester"
)

func TestStatistics_getAverageOdd_ReturnRightAverageOddForHomeFavorite(
	t *testing.T,
) {

	event1 := requester.LiveEventResult{
		EventID:           "1",
		Favorite:          "home",
		LastHomeOdd:       2.00,
		LastAwayOdd:       0,
		WinnerInSecondSet: "home",
	}

	event2 := requester.LiveEventResult{
		EventID:           "2",
		Favorite:          "home",
		LastHomeOdd:       4.00,
		LastAwayOdd:       0,
		WinnerInSecondSet: "home",
	}

	event3 := requester.LiveEventResult{
		EventID:           "3",
		Favorite:          "home",
		LastHomeOdd:       6.00,
		LastAwayOdd:       0,
		WinnerInSecondSet: "home",
	}

	var events []requester.LiveEventResult
	events = append(events, event1, event2, event3)

	result := getAverageOdd(events)

	assert.Equal(t, float64(4), result)

}

func TestStatistics_getAverageOdd_ReturnRightAverageOddForAwayFavorite(
	t *testing.T,
) {

	event1 := requester.LiveEventResult{
		EventID:           "1",
		Favorite:          "away",
		LastHomeOdd:       2.00,
		LastAwayOdd:       8,
		WinnerInSecondSet: "away",
	}

	event2 := requester.LiveEventResult{
		EventID:           "2",
		Favorite:          "away",
		LastHomeOdd:       4.00,
		LastAwayOdd:       7,
		WinnerInSecondSet: "away",
	}

	event3 := requester.LiveEventResult{
		EventID:           "3",
		Favorite:          "away",
		LastHomeOdd:       6.00,
		LastAwayOdd:       7,
		WinnerInSecondSet: "away",
	}

	var events []requester.LiveEventResult
	events = append(events, event1, event2, event3)

	result := getAverageOdd(events)

	assert.Equal(t, float64(7.33), result)
}
