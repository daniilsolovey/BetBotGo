package operator

import (
	"github.com/reconquest/karma-go"
	"gopkg.in/tucnak/telebot.v2"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (operator *Operator) Start(message *tb.Message) error {
	text := "Hi! I am a telegram bot and I can notify you about all events for today" +
		" on volleyball"
	var recipient telebot.Recipient
	var recipientID int
	if message.Chat != nil {
		recipient = message.Chat
		recipientID = int(message.Chat.ID)
	} else {
		recipient = message.Sender
		recipientID = message.Sender.ID
	}

	err := operator.transport.SendMessage(recipient, text)
	if err != nil {
		return karma.Format(err, "unable to send message to user: %d ",
			recipientID)
	}

	return nil
}
