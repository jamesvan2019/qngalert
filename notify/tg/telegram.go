package tg

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"qngalert/config"
)

type TelegramBot struct {
	api *tgbotapi.BotAPI
	Cfg *config.Config
}

func (t *TelegramBot) Init(cfg *config.Config) error {
	t.Cfg = cfg
	bot, err := tgbotapi.NewBotAPI(cfg.Tg.Token)
	if err != nil {
		log.Println(err)
		return err
	}
	//bot.Debug = true
	t.api = bot
	return nil
}

func (t *TelegramBot) Notify(title, content string) error {
	str := fmt.Sprintf("%s\n%s", title, content)
	msg := tgbotapi.NewMessage(t.Cfg.Tg.ChatID, str)
	_, err := t.api.Send(msg)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Send Telegram message", str)
	return nil
}
