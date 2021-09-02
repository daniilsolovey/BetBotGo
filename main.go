package main

import (
	"sync"
	"time"

	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/daniilsolovey/BetBotGo/internal/database"
	"github.com/daniilsolovey/BetBotGo/internal/operator"
	"github.com/daniilsolovey/BetBotGo/internal/requester"
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
	operator := operator.NewOperator(
		config, nil, requester,
	)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		log.Info("start cycle with receiving events for today")
		for {
			events, err := operator.GetEventsForToday()
			if err != nil {
				log.Error(err)
			}

			log.Warning("events result2222!!!!", events)
			log.Warning("len(events)", len(events))
			err = database.InsertEventsForToday(events)
			if err != nil {
				log.Error(err)
			}

			err = operator.CreateRoutinesForEachEvent(events)
			if err != nil {
				log.Error(err)
			}

			time.Sleep(1 * time.Hour)
		}
	}()
	wg.Wait()
}
