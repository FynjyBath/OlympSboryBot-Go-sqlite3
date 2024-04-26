package tgbot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TGBot struct {
	Bot *tgbotapi.BotAPI
}

func (tgbot *TGBot) Init(tgbotkey string) {
	bot, err := tgbotapi.NewBotAPI(tgbotkey)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	tgbot.Bot = bot
}

func (bot *TGBot) SendMessage(id int, text string) (int, error) {
	msg, err := bot.Bot.Send(tgbotapi.NewMessage(int64(id), text))
	if err != nil {
		return 0, err
	}
	return msg.MessageID, nil
}

func (bot *TGBot) SendForward(id1, id2, id3 int) (int, error) {
	msg, err := bot.Bot.Send(tgbotapi.NewForward(int64(id1), int64(id2), id3))
	if err != nil {
		return 0, err
	}
	return msg.MessageID, nil
}
