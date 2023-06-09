package qng

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

type GlobalRecords struct {
	MemPoolCount map[int64]int64
	Lock         sync.Mutex
	PrintLine    map[int64]int64
}

var globalParam = GlobalRecords{
	MemPoolCount: map[int64]int64{},
	PrintLine:    map[int64]int64{},
}

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

				if n.Cfg.UseStateRoot {
					n.CompareStateRoot(order)
				}
			}
			// 2023-03-15T14:04:09+08:00
			gap := int64(0)
			if strings.Contains(blockDetail.Result.Timestamp, "-04:00") {
				gap = 4 * 3600
			}
			if strings.Contains(blockDetail.Result.Timestamp, "Z") {
				gap = 8 * 3600
			}
			time1 := strings.ReplaceAll(blockDetail.Result.Timestamp, "-04:00", "")
			time1 = strings.ReplaceAll(time1, "+08:00", "")
			time1 = strings.ReplaceAll(time1, "Z", "")

			t1, err := time.Parse("2006-01-02T15:04:05", time1)
			if err != nil {
				n.ErrorMsg(time1+" timestamp parse error", err)
				continue
			}
			if time.Now().Unix()-t1.Unix()-gap >= n.Cfg.Alert.MaxBlockTime {
				n.GetMinerErrorTimes++
				if n.GetMinerErrorTimes >= n.Cfg.Alert.MaxAllowErrorTimes {
					n.NotifyClients.Send("miner alert",
						n.ErrorMsgFormat("long time not got new block",
							fmt.Errorf("latest order:%d , latest block time:%s | long time not got new block",
								blockDetail.Result.Order, blockDetail.Result.Timestamp)))
				}
				continue
			}
			n.GetMinerErrorTimes = 0
			n.Msg(fmt.Sprintf("[node normal] | latest order :%d | latest mining time:%s", n.LastestOrder, blockDetail.Result.Timestamp))
			b, _ := json.Marshal(StateRootObj.StateRoots[order])
			n.Msg(fmt.Sprintf("[stateroot]:%v", string(b)))
		}
	}
}

func (n *Node) ListenCheckPeers(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	n.Msg("start ListenCheckPeers Service")
	for {
		select {
		case <-ctx.Done():
			n.Msg("stop ListenCheckPeers Service,exit...")
			return
		default:
			<-time.After(30 * time.Second)
			t1 := time.Now().Unix()
			reqTime := t1 - t1%30
			count, err := n.GetPeers()
			if err != nil {
				if n.GetPeersErrorTimes >= n.Cfg.Alert.MaxAllowErrorTimes {
					//
					n.NotifyClients.Send("node peers exception",
						n.ErrorMsgFormat("GetPeers Rpc Exception many times,please check", err))
				}
				continue
			}
			globalParam.Lock.Lock()
			_, ok := globalParam.PrintLine[reqTime]
			if !ok {
				globalParam.PrintLine[reqTime] = 1
				fmt.Println("=======================", reqTime, "==========================")
			}
			globalParam.Lock.Unlock()
			if count < 5 {
				if time.Now().Unix()-n.lastReset < 10*60 {
					// 10分钟前刚执行过
					n.ErrorMsg("p2p_resetPeers 10分钟前刚执行过", fmt.Errorf(""))
					continue
				}
				n.NotifyClients.Send("node peers exception",
					n.ErrorMsgFormat("peer nodes count less than 10,please check", errors.New("peers too less")))
				//n.ResetPeer()
				continue
			}
			count1, _ := n.GetMempoolCount(false)
			globalParam.Lock.Lock()
			targetCount, ok := globalParam.MemPoolCount[reqTime]
			if !ok {
				if count1 > 0 {
					globalParam.MemPoolCount[reqTime] = count1
				}
			}
			globalParam.Lock.Unlock()
			if targetCount > 0 && (math.Abs(float64(targetCount-count1)) > 10 || count1 == 0) {
				if math.Abs(float64(targetCount-count1)) > 10 {
					n.zhangben++
				}
				msg := ""
				if n.zhangben > 20 {
					msg = "可能是账本问题"
					n.NotifyClients.Send("node peers exception:"+msg,
						n.ErrorMsgFormat(fmt.Sprintf("reqTimes:%d | targetMempoolCount:%d | currentMempoolCount:%d ", reqTime, targetCount, count1), errors.New("memorypool poor connection")))
					continue
				}
				if count1 == 0 {
					n.MempoolEmptyTimes++
					if n.MempoolEmptyTimes > 6 {
						n.MempoolEmptyTimes = 0
						if time.Now().Unix()-n.lastReset < 10*60 {
							// 10分钟前刚执行过
							n.ErrorMsg("p2p_resetPeers 10分钟前刚执行过", fmt.Errorf(""))
							continue
						}
						msg = "mempool同步问题 需要reset"
						n.NotifyClients.Send("node peers exception:"+msg,
							n.ErrorMsgFormat(fmt.Sprintf("reqTimes:%d | targetMempoolCount:%d | currentMempoolCount:%d ", reqTime, targetCount, count1), errors.New("memorypool poor connection")))
						//n.ResetPeer()
						continue
					}
				}
			}
			if targetCount == 0 && count1 == 0 {
				n.MempoolEmptyTimes = 0
			}
			n.zhangben = 0
			n.Msg(fmt.Sprintf("[node normal] | peersCount :%d | mempool:%d | mempool empty Times:%d", count, count1, n.MempoolEmptyTimes))
		}
	}
}

func (n *Node) ListenNode(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	n.Msg("start ListenNode Service")
	for {
		select {
		case <-ctx.Done():
			n.Msg("stop ListenNode Service,exit...")
			return
		default:
			node, err := n.GetNodeInfo()
			if err != nil {
				n.ErrorMsg("GetNodeInfo Error", fmt.Errorf(""))
			} else {
				n.Node = *node
			}
			<-time.After(120 * time.Second)
		}
	}
}

// 3 times retry
func (n *Node) CompareStateRoot(order int64) {
	for i := int64(0); i < n.Cfg.Alert.MaxAllowErrorTimes; i++ {
		if n.Compare(order, i) {
			return
		}
		time.Sleep(30 * time.Second)
	}
}

func (n *Node) Compare(order, i int64) bool {
	stateRoot, err := n.GetStateRoot(order)
	if err != nil {
		if n.GetStateRootErrorTimes >= n.Cfg.Alert.MaxAllowErrorTimes {
			//
			n.NotifyClients.Send("node exception",
				n.ErrorMsgFormat("GetStateRoot Rpc Exception many times,please check", err))
		}
		return true
	}
	stateRoot.Result.Node = n.Cfg.Rpc
	StateRootObj.lock.Lock()
	defer StateRootObj.lock.Unlock()
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
			return false
		}
		if stateRoot.Result.Hash != target.Hash {
			if i == 2 {
				n.NotifyClients.Send("node hash not equal",
					n.ErrorMsgFormat("node hash not equal,please check",
						fmt.Errorf("target node:%s,order:%d hash:%v | node:%s,order:%d hash:%v",
							target.Node, target.Order, target.Hash, stateRoot.Result.Node, order, stateRoot.Result.Hash)))
			}
			return false
		}
		if stateRoot.Result.StateRoot != target.StateRoot {
			if i == 2 {
				n.NotifyClients.Send("node StateRoot not equal",
					n.ErrorMsgFormat("node StateRoot not equal,please check",
						fmt.Errorf("target node:%s,order:%d stateroot:%v ,number:%d | node:%s,order:%d stateroot:%v,number:%d",
							target.Node, target.Order, target.StateRoot, target.Number, stateRoot.Result.Node, order,
							stateRoot.Result.StateRoot, stateRoot.Result.Number)))
			}
			return false
		}
	}
	return true
}
