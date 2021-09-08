package main

import (
	"sync"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"

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
	requester := requester.NewRequester(config)

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
		config, database, requester, telegramBot,
	)
	newStatistic := statistics.NewStatistics(database)

	err = newStatistic.GetLiveEventsResultsOnPreviousDateAndWriteToStatistic()
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Info("start cycle with receiving events for today")
		for {
			timeNow, err := tools.GetCurrentMoscowTime()
			if err != nil {
				log.Error(err)
			}

			err = newStatistic.GetLiveEventsResultsOnPreviousDateAndWriteToStatistic()
			if err != nil {
				log.Fatal(err)
			}

			events, err := newOperator.GetEvents()
			if err != nil {
				log.Error(err)
			}

			// if len(events) == 0 {
			// 	log.Warning("len(events) ", len(events))
			// 	diff := timeNow.Add(1 * time.Hour).Sub(timeNow)
			// 	time.Sleep(diff)
			// 	continue
			// }

			log.Info("insert events for today to database")
			err = database.InsertEventsForToday(events)
			if err != nil {
				log.Error(err)
			}

			log.Info("creating routines for each event")
			err = newOperator.CreateRoutinesForEachEvent(events)
			if err != nil {
				log.Error(err)
			}

			log.Warning("timeNow.Truncate(24 * time.Hour).Add(21 * time.Hour).Add(1 * time.Second) ", timeNow.Truncate(24*time.Hour).Add(21*time.Hour).Add(1*time.Second))
			diff := timeNow.Truncate(24 * time.Hour).Add(21 * time.Hour).Sub(timeNow)
			log.Warning("timeNow ", timeNow)
			log.Warning("diff ", diff)
			time.Sleep(diff)
		}
	}()

	wg.Add(2)
	go func() {
		telegramBot.Handle("/starttest", newOperator.Start)
		log.Infof(nil, "starting to listen and serve telegram bot")
		bot.Start()
	}()

	wg.Wait()
}
