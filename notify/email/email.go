package email

import (
	"fmt"
	"log"
	"net/smtp"
	"qngalert/config"
	"strings"
)

type Email struct {
	Cfg *config.Config
}

func (e *Email) Init(cfg *config.Config) error {
	e.Cfg = cfg
	return nil
}

func (e *Email) Notify(title, content string) error {
	auth := smtp.PlainAuth("", e.Cfg.Email.User, e.Cfg.Email.Pass, e.Cfg.Email.Host)
	sendTo := strings.Split(e.Cfg.Email.To, ";")
	for _, v := range sendTo {
		str := strings.Replace("From: "+e.Cfg.Email.User+"~To: "+v+"~Subject: "+title+"~~", "~", "\r\n", -1) + content
		err := smtp.SendMail(
			fmt.Sprintf("%s:%s", e.Cfg.Email.Host, e.Cfg.Email.Port),
			auth,
			e.Cfg.Email.User,
			[]string{v},
			[]byte(str),
		)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	log.Println("Send Email Message", fmt.Sprintf("%s\n%s", title, content), "|To:", e.Cfg.Email.To)
	return nil
}
