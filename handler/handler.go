package handler

import (
	"encoding/json"

	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
)

type Handler struct {
	database *database.Database
	config   *config.Config
}

type ResultEventsHandler struct {
	EventID string
	League  struct {
		Name string
	}

	EventStartTime  string
	Favorite        string
	HomeCommandName string
	AwayCommandName string
	HomeOdd         float64
	AwayOdd         float64
}

type Response struct {
	UpcomingEvents []ResultEventsHandler `json:"upcomingEvents"`
}

func NewHandler(
	database *database.Database,
	config *config.Config,
) *Handler {
	return &Handler{
		database: database,
		config:   config,
	}
}

func JSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

func (handler *Handler) StartServer(config *config.Config) {
	router := gin.Default()
	router.GET("/", handler.ActionIndex)
	router.Use(JSONMiddleware())
	router.GET("/upcoming_events", handler.UpcomingEvents)

	router.Run(handler.config.Handler.Port)
}

func (handler *Handler) UpcomingEvents(context *gin.Context) {
	events, err := handler.database.GetUpcomingEventsForToday()
	if err != nil {
		log.Error(karma.Format(
			err,
			"unable to get upcoming_events_handler for today from database",
		))

	}
	var resultsEvents []ResultEventsHandler
	for _, event := range events {
		var resultEventsHandler ResultEventsHandler
		resultEventsHandler.EventID = event.EventID
		resultEventsHandler.League.Name = event.League.Name
		resultEventsHandler.HomeCommandName = event.HomeCommandName
		resultEventsHandler.AwayCommandName = event.AwayCommandName
		resultEventsHandler.HomeOdd = event.HomeOdd
		resultEventsHandler.AwayOdd = event.AwayOdd
		resultEventsHandler.Favorite = event.Favorite
		resultEventsHandler.EventStartTime = event.EventStartTime.Format("02 Jan 06 15:04 MSK")
		resultsEvents = append(resultsEvents, resultEventsHandler)
	}
	var response Response

	response.UpcomingEvents = resultsEvents

	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Error("unable to decode to bytes upcoming_events")
	}
	context.Data(
		200,
		"text/plain; charset=UTF-8",
		[]byte(responseBytes),
	)

}

func (handler *Handler) ActionIndex(context *gin.Context) {
	context.Data(
		200,
		"text/plain; charset=UTF-8",
		[]byte("This api version: "+handler.config.Handler.ApiVersion),
	)
}
