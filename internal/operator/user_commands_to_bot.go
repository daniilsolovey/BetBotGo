package operator

import (
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
	"gopkg.in/tucnak/telebot.v2"
	tb "gopkg.in/tucnak/telebot.v2"
)

var TEMP_RECIPIENT telebot.Recipient

const (
	TEXT_ABOUT_WINNER = "WARNING! Делай ставку!\n" +
		"  event_id: %s\n" +
		"  league_name: %s\n" +
		"  last_odd_home: %s\n" +
		"  last_odd_away: %s\n" +
		"  home_command_name: %s\n" +
		"  away_command_name: %s\n" +
		"  favorite: %s\n"
)

func (operator *Operator) Start(message *tb.Message) error {
	text := "Hi! I am a telegram bot and I can notify you about all events for today" +
		" on volleyball"
	var recipient telebot.Recipient
	var recipientID int
	if message.Chat != nil {
		recipient = message.Chat
		TEMP_RECIPIENT = message.Chat
		recipientID = int(message.Chat.ID)
	} else {
		recipient = message.Sender
		TEMP_RECIPIENT = message.Sender
		recipientID = message.Sender.ID
	}

	err := operator.transport.SendMessage(recipient, text)
	if err != nil {
		return karma.Format(err, "unable to send message to user: %d ",
			recipientID)
	}

	log.Warning("recipient ", recipient)
	return nil
}
