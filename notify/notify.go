package notify

import (
	"log"
	"qngmempool/config"
)

type Notify interface {
	Init(cfg *config.Config) error
	Notify(title, content string) error
}

type Clients []Notify

func (nc *Clients) Send(title, content string) {
	for _, c := range *nc {
		err := c.Notify(title, content)
		if err != nil {
			log.Println("发送通知失败", err)
		}
	}
}
