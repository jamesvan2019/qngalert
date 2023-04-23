package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"qngmempool/config"
	"qngmempool/notify"
	"qngmempool/notify/whatsapp"
	"qngmempool/qng"
	"strings"
	"sync"
	"syscall"
	"time"
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
	// whatsapp
	whatsappClient := &whatsapp.WhatsappBot{}
	whatsappClient.Init(&cfg)
	wg := &sync.WaitGroup{}
	notifyClients = append(notifyClients, whatsappClient)
	qng.StateRootObj = qng.StateRootObjStruct{
		StateRoots:    map[int64]qng.StateRoot{},
		StateRootsArr: []int64{},
	}
	// 初始化
	//
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	wg.Add(1)
	go handleSignal(wg, quit, cancel)
	baseNode := &qng.Node{}
	err = baseNode.Init(cfg.Nodes[0], notifyClients)
	if err != nil {
		log.Fatalln(err)
		return
	}
	nodes, err := baseNode.GetPeerList()
	if err != nil {
		log.Fatalln(err)
		return
	}
	dockerNodes := make([]*qng.Node, 0)
	for _, node := range nodes {
		if !strings.Contains(node.Version, "docker") {
			continue
		}
		nc := &qng.Node{}
		n := config.Node{
			Rpc:  getRpc(node.Address),
			User: "test",
			Pass: "test",
			Gap:  15,
			Alert: config.Alert{
				MaxAllowErrorTimes: 3,
				MaxBlockTime:       10 * 60,
			},
		}
		err = nc.Init(n, notifyClients)
		if err != nil {
			log.Fatalln(err)
			return
		}
		nc.Node.ID = node.Address
		nc.Node.Buildversion = node.Version
		dockerNodes = append(dockerNodes, nc)
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:

		}
		nodeWg := &sync.WaitGroup{}
		for _, v := range dockerNodes {
			nodeWg.Add(1)
			go func(node *qng.Node) {
				defer nodeWg.Done()
				// 查询最新区块order
				node.NewestNodeInfo.BlockCount, _ = node.GetBlockCount()
				node.NewestNodeInfo.MempoolCount, _ = node.GetMempoolCount(false)
				l, _ := node.GetPeerList()
				pc := int64(0)
				for _, vv := range l {
					if vv.State && vv.Active {
						pc++
					}
				}
				node.NewestNodeInfo.PeerCount = pc
			}(v)
		}
		nodeWg.Wait()
		msg := ""
		for i, v := range dockerNodes {
			msg += fmt.Sprintf("[docker-%d]%s/%s[peerCount]*%d*[blockOrder]*%d*[mempoolCount]*%d*\n", i+1, v.Node.ID, v.Node.Buildversion, v.NewestNodeInfo.PeerCount,
				v.NewestNodeInfo.BlockCount, v.NewestNodeInfo.MempoolCount)
		}
		notifyClients.Send("docker nodes status", msg)
		time.Sleep(30 * time.Minute)
	}
	wg.Wait()
	whatsappClient.Stop()
}
func getRpc(addr string) string {
	arr := strings.Split(addr, "/")
	if len(arr) < 5 {
		return ""
	}
	return fmt.Sprintf("https://%s:18131", arr[2])
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
