package main

import (
	"sync"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/daniilsolovey/BetBotGo/handler"
	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/daniilsolovey/BetBotGo/internal/operator"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
	"github.com/daniilsolovey/BetBotGo/internal/statistics"
	"github.com/daniilsolovey/BetBotGo/internal/tools"
	"github.com/daniilsolovey/BetBotGo/internal/transport"
	"github.com/docopt/docopt-go"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
)

var version = "[manual build]"

var usage = `BetBotGo

Download ticks and save them to the database.

Usage:
  BetBotGo [options]

Options:
  -c --config <path>                Read specified config file. [default: config.yaml]
  --debug                           Enable debug messages.
  -v --version                      Print version.
  -h --help                         Show this help.
`

const (
	RECEIVING_EVENTS_DURATION = 5 * time.Minute
)

func main() {
	args, err := docopt.ParseArgs(
		usage,
		nil,
		"BetBotGo "+version,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof(
		karma.Describe("version", version),
		"BetBotGo started",
	)

	if args["--debug"].(bool) {
		log.SetLevel(log.LevelDebug)
	}

	log.Infof(nil, "loading configuration file: %q", args["--config"].(string))

	config, err := config.Load(args["--config"].(string))
	if err != nil {
		log.Fatal(err)
	}

	log.Infof(
		karma.Describe("database", config.Database.Name),
		"connecting to the database",
	)

	database := database.NewDatabase(
		config.Database.Name, config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password,
	)
	err = database.CreateTables()
	if err != nil {
		log.Fatal(err)
	}

	defer database.Close()
	newRequester := requester.NewRequester(config)

	log.Info("creating telegram bot")
	bot, err := tb.NewBot(
		tb.Settings{
			Token:  config.Telegram.Token,
			Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	telegramBot := transport.NewBot(bot)

	log.Info("creating operator")
	newOperator := operator.NewOperator(
		config, database, newRequester, telegramBot,
	)
	newStatistic := statistics.NewStatistics(database, telegramBot)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Info("start cycle with receiving events for today")
		for {
			events, err := newOperator.GetEvents()
			if err != nil {
				log.Error(err)
			}

			err = database.InsertEventsForToday(events)
			if err != nil {
				log.Error(err)
			}

			err = newOperator.CreateRoutinesForHandleLiveEvents(events)
			if err != nil {
				log.Error(err)
			}

			time.Sleep(RECEIVING_EVENTS_DURATION)

		}
	}()

	wg.Add(1)
	go func() {
		log.Info("start cycle with receiving statistic on previous day")
		for {
			timeNow, err := tools.GetCurrentMoscowTime()
			if err != nil {
				log.Error(err)
			}

			beginOfDay := roundToBeginningOfDay(timeNow)
			waitUntill := beginOfDay.Add(24 * time.Hour)
			waitingTime := waitUntill.Sub(timeNow)

			time.Sleep(waitingTime)
			err = newStatistic.GetStatisticOnPreviousDayAndNotify()
			if err != nil {
				log.Error(err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		log.Info("start cycle with receiving statistic on previous week")
		for {
			timeNow, err := tools.GetCurrentMoscowTime()
			if err != nil {
				log.Error(err)
			}

			beginOfDay := roundToBeginningOfDay(timeNow)
			waitUntill := beginOfDay.Add(24 * time.Hour)
			waitingTime := waitUntill.Sub(timeNow)

			weekday := time.Now().Weekday()
			if weekday == time.Monday {
				err = newStatistic.GetStatisticOnPreviousWeekAndNotify()
				if err != nil {
					log.Error(err)
				}
			}

			time.Sleep(waitingTime)
		}
	}()

	newHandler := handler.NewHandler(database, config)
	go func() {
		newHandler.StartServer(config)
	}()

	telegramBot.Handle("/starttest", newOperator.Start)
	log.Infof(nil, "starting to listen and serve telegram bot")
	bot.Start()

	wg.Wait()
}

func roundToBeginningOfDay(t time.Time) time.Time {
	moscowLocation, err := tools.GetTimeMoscowLocation()
	if err != nil {
		log.Error(err)
	}

	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, moscowLocation)
}
