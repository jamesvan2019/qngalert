package qng

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func (n *Node) ListenNodeStatus(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	t := time.NewTicker(time.Duration(n.Cfg.Gap) * time.Second)
	defer t.Stop()
	n.Msg("start ListenMinerStatus Service")
	for {
		select {
		case <-ctx.Done():
			n.Msg("stop ListenMinerStatus Service,exit...")
			return
		case <-t.C:
			order, err := n.GetBlockCount()
			if err != nil {
				if n.GetBlockCountErrorTimes >= n.Cfg.Alert.MaxAllowErrorTimes {
					//
					n.NotifyClients.Send("node exception",
						n.ErrorMsgFormat("GetBlockCount Rpc Exception many times,please check", err))
				}
				continue
			}
			blockDetail, err := n.GetBlockByOrder(order)
			if err != nil {
				if n.GetBlockByOrderErrorTimes >= n.Cfg.Alert.MaxAllowErrorTimes {
					//
					n.NotifyClients.Send("node exception",
						n.ErrorMsgFormat("GetBlockByOrder Rpc Exception many times,please check", err))
				}
				continue
			}

			if order > n.LastestOrder {
				n.LastestOrder = order
				if len(blockDetail.Result.Transactions) <= 0 {
					n.NotifyClients.Send("empty block alert",
						n.ErrorMsgFormat("block empty", fmt.Errorf("order:%d is empty block", blockDetail.Result.Order)))
					continue
				}
				if !blockDetail.Result.TxValid {
					n.NotifyClients.Send("block txvalid false alert",
						n.ErrorMsgFormat("block txvalid false exception", fmt.Errorf("order:%d block txvalid is false", blockDetail.Result.Order)))
					continue
				}
			}
			// 2023-03-15T14:04:09+08:00
			t1, err := time.Parse("2006-01-02T15:04:05+08:00", blockDetail.Result.Timestamp)
			if err != nil {
				n.ErrorMsg("timestamp parse error", err)
				continue
			}
			if time.Now().Unix()+8*3600-t1.Unix() >= n.Cfg.Alert.MaxBlockTime {
				n.NotifyClients.Send("miner alert",
					n.ErrorMsgFormat("long time not got new block",
						fmt.Errorf("latest order:%d , latest block time:%s | long time not got new block",
							blockDetail.Result.Order, blockDetail.Result.Timestamp)))
				continue
			}
			n.Msg(fmt.Sprintf("node normal | latest order :%d | latest mining time:%s", n.LastestOrder, blockDetail.Result.Timestamp))
		}
	}
}
