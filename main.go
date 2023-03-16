package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"qngalert/config"
	"qngalert/notify"
	"qngalert/notify/email"
	"qngalert/notify/tg"
	"qngalert/qng"
	"sync"
	"syscall"
)

func main() {
	// 1 watch node status
	confPath := flag.String("config", "./config.json", "configPath")
	flag.Parse()
	b, err := ioutil.ReadFile(*confPath)
	if err != nil {
		log.Fatalln(err)
		return
	}
	var cfg config.Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatalln(err)
		return
	}
	notifyClients := notify.Clients{}
	if cfg.Email.Enable {
		emailInstance := &email.Email{}
		err = emailInstance.Init(&cfg)
		if err != nil {
			log.Fatalln(err)
			return
		}
		notifyClients = append(notifyClients, emailInstance)
	}
	if cfg.Tg.Enable {
		tgInstance := &tg.TelegramBot{}
		err = tgInstance.Init(&cfg)
		if err != nil {
			log.Fatalln(err)
			return
		}
		notifyClients = append(notifyClients, tgInstance)
	}
	qng.StateRootObj = qng.StateRootObjStruct{
		StateRoots:    map[int64]qng.StateRoot{},
		StateRootsArr: []int64{},
	}
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go handleSignal(wg, quit, cancel)
	for _, n := range cfg.Nodes {
		nc := &qng.Node{}
		err = nc.Init(&n, notifyClients)
		if err != nil {
			log.Fatalln(err)
			return
		}
		wg.Add(1)
		go func(nc1 *qng.Node) {
			nc1.ListenNodeStatus(ctx, wg)
		}(nc)
	}
	wg.Wait()
}
func handleSignal(wg *sync.WaitGroup, c chan os.Signal, cancel context.CancelFunc) {
	defer wg.Done()
	switch <-c {
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
		log.Println("Shutdown quickly, bye...")
		cancel()
	case syscall.SIGHUP:
		log.Println("Shutdown gracefully, bye...")
		cancel()
	}
}
