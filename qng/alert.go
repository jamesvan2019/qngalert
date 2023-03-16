package qng

import (
	"context"
	"fmt"
	"strings"
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

				stateRoot, err := n.GetStateRoot(order)
				if err != nil {
					if n.GetStateRootErrorTimes >= n.Cfg.Alert.MaxAllowErrorTimes {
						//
						n.NotifyClients.Send("node exception",
							n.ErrorMsgFormat("GetStateRoot Rpc Exception many times,please check", err))
					}
					continue
				}
				stateRoot.Result.Node = n.Cfg.Rpc
				StateRootObj.lock.Lock()
				if _, ok := StateRootObj.StateRoots[order]; !ok {
					StateRootObj.StateRoots[order] = stateRoot.Result
					if len(StateRootObj.StateRootsArr) > 20 {
						delete(StateRootObj.StateRoots, StateRootObj.StateRootsArr[0])
						StateRootObj.StateRootsArr = StateRootObj.StateRootsArr[1:]
					}
					StateRootObj.StateRootsArr = append(StateRootObj.StateRootsArr, order)
				} else {
					// compare start
					target := StateRootObj.StateRoots[order]
					if stateRoot.Result.Valid != target.Valid {
						n.NotifyClients.Send("node valid not equal",
							n.ErrorMsgFormat("node valid not equal,please check",
								fmt.Errorf("target node:%s,order:%d valid:%v | node:%s,order:%d valid:%v",
									target.Node, target.Order, target.Valid, stateRoot.Result.Node, order, stateRoot.Result.Valid)))
						continue
					}
					if stateRoot.Result.Hash != target.Hash {
						n.NotifyClients.Send("node hash not equal",
							n.ErrorMsgFormat("node hash not equal,please check",
								fmt.Errorf("target node:%s,order:%d hash:%v | node:%s,order:%d hash:%v",
									target.Node, target.Order, target.Hash, stateRoot.Result.Node, order, stateRoot.Result.Hash)))
						continue
					}
					if stateRoot.Result.StateRoot != target.StateRoot {
						n.NotifyClients.Send("node StateRoot not equal",
							n.ErrorMsgFormat("node StateRoot not equal,please check",
								fmt.Errorf("target node:%s,order:%d stateroot:%v ,number:%d | node:%s,order:%d stateroot:%v,number:%d",
									target.Node, target.Order, target.StateRoot, target.Number, stateRoot.Result.Node, order,
									stateRoot.Result.StateRoot, stateRoot.Result.Number)))
						continue
					}
				}
				// compare end
				StateRootObj.lock.Unlock()
			}
			// 2023-03-15T14:04:09+08:00
			time1 := ""
			if strings.Contains(blockDetail.Result.Timestamp, "-") {
				time1 = strings.Split(blockDetail.Result.Timestamp, "-")[0]
			}
			if strings.Contains(blockDetail.Result.Timestamp, "+") {
				time1 = strings.Split(blockDetail.Result.Timestamp, "+")[0]
			}

			t1, err := time.Parse("2006-01-02T15:04:05", time1)
			if err != nil {
				n.ErrorMsg("timestamp parse error", err)
				continue
			}
			if time.Now().Unix()-t1.Unix() >= n.Cfg.Alert.MaxBlockTime {
				n.NotifyClients.Send("miner alert",
					n.ErrorMsgFormat("long time not got new block",
						fmt.Errorf("latest order:%d , latest block time:%s | long time not got new block",
							blockDetail.Result.Order, blockDetail.Result.Timestamp)))
				continue
			}

			n.Msg(fmt.Sprintf("node normal | latest order :%d | latest mining time:%s", n.LastestOrder, blockDetail.Result.Timestamp))
			n.Msg(fmt.Sprintf("stateroot:%v", StateRootObj.StateRoots[order]))
		}
	}
}

// 3 times retry
func (n *Node) CompareStateRoot(order int64) {
	for i := int64(0); i < n.Cfg.Alert.MaxAllowErrorTimes; i++ {
		stateRoot, err := n.GetStateRoot(order)
		if err != nil {
			if n.GetStateRootErrorTimes >= n.Cfg.Alert.MaxAllowErrorTimes {
				//
				n.NotifyClients.Send("node exception",
					n.ErrorMsgFormat("GetStateRoot Rpc Exception many times,please check", err))
			}
			return
		}
		stateRoot.Result.Node = n.Cfg.Rpc
		StateRootObj.lock.Lock()
		if _, ok := StateRootObj.StateRoots[order]; !ok {
			StateRootObj.StateRoots[order] = stateRoot.Result
			if len(StateRootObj.StateRootsArr) > 20 {
				delete(StateRootObj.StateRoots, StateRootObj.StateRootsArr[0])
				StateRootObj.StateRootsArr = StateRootObj.StateRootsArr[1:]
			}
			StateRootObj.StateRootsArr = append(StateRootObj.StateRootsArr, order)
		} else {
			// compare start
			target := StateRootObj.StateRoots[order]
			if stateRoot.Result.Valid != target.Valid {
				if i == 2 {
					n.NotifyClients.Send("node valid not equal",
						n.ErrorMsgFormat("node valid not equal,please check",
							fmt.Errorf("target node:%s,order:%d valid:%v | node:%s,order:%d valid:%v",
								target.Node, target.Order, target.Valid, stateRoot.Result.Node, order, stateRoot.Result.Valid)))
				}
				continue
			}
			if stateRoot.Result.Hash != target.Hash {
				if i == 2 {
					n.NotifyClients.Send("node hash not equal",
						n.ErrorMsgFormat("node hash not equal,please check",
							fmt.Errorf("target node:%s,order:%d hash:%v | node:%s,order:%d hash:%v",
								target.Node, target.Order, target.Hash, stateRoot.Result.Node, order, stateRoot.Result.Hash)))
				}
				continue
			}
			if stateRoot.Result.StateRoot != target.StateRoot {
				if i == 2 {
					n.NotifyClients.Send("node StateRoot not equal",
						n.ErrorMsgFormat("node StateRoot not equal,please check",
							fmt.Errorf("target node:%s,order:%d stateroot:%v ,number:%d | node:%s,order:%d stateroot:%v,number:%d",
								target.Node, target.Order, target.StateRoot, target.Number, stateRoot.Result.Node, order,
								stateRoot.Result.StateRoot, stateRoot.Result.Number)))
				}
				continue
			}
		}
		return
	}
	time.Sleep(30 * time.Second)
}
