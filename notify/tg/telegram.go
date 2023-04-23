package tg

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"qngmempool/config"
	"time"
)

type TelegramBot struct {
	api      *tgbotapi.BotAPI
	Cfg      *config.Config
	lastSend int64
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
	if time.Now().Unix()-t.lastSend < 600 { // 10分钟内不重发
		return nil
	}
	t.lastSend = time.Now().Unix()
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
