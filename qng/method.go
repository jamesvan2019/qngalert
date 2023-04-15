package qng

import (
	"encoding/json"
	"fmt"
	"time"
)

func (n *Node) GetBlockCount() (int64, error) {
	b, err := n.rpcResult("getBlockCount", []interface{}{})
	if err != nil {
		n.ErrorMsg("GetBlockCount Exception", err)
		n.GetBlockCountErrorTimes++
		return 0, err
	}
	var r BlockCountResult
	err = json.Unmarshal(b, &r)
	if err != nil {
		n.ErrorMsg("GetBlockCount Unmarshal Exception", err)
		n.GetBlockCountErrorTimes++
		return 0, err
	}
	if r.Result < 1 {
		n.GetBlockCountErrorTimes++
		n.ErrorMsg("Result Exception", fmt.Errorf("GetBlockCount Result is 0"))
	}
	n.GetBlockCountErrorTimes = 0
	return r.Result - 1, nil
}

func (n *Node) GetPeers() (int64, error) {
	b, err := n.rpcResult("getPeerInfo", []interface{}{})
	if err != nil {
		n.ErrorMsg("GetPeersErrorTimes Exception", err)
		n.GetPeersErrorTimes++
		return 0, err
	}
	var r PeerInfoResult
	err = json.Unmarshal(b, &r)
	if err != nil {
		n.ErrorMsg("GetPeersErrorTimes Unmarshal Exception", err)
		n.GetPeersErrorTimes++
		return 0, err
	}
	if len(r.Result) < 1 {
		n.GetPeersErrorTimes++
		n.ErrorMsg("Result Exception", fmt.Errorf("GetPeersErrorTimes Result is 0"))
	}
	n.GetPeersErrorTimes = 0
	connected := int64(0)
	for _, v := range r.Result {
		if v.State == "connected" && v.Services != "Relay" {
			connected++
		}
	}
	return connected, nil
}

func (n *Node) GetMempoolCount(retry bool) (int64, error) {
	b, err := n.rpcResult("getMempool", []interface{}{"", false})
	if err != nil {
		n.ErrorMsg("GetMempoolErrorTimes Exception", err)
		n.GetMempoolErrorTimes++
		return 0, err
	}
	var r MempoolResult
	err = json.Unmarshal(b, &r)
	if err != nil {
		n.ErrorMsg("GetMempoolErrorTimes Unmarshal Exception", err)
		n.GetMempoolErrorTimes++
		return 0, err
	}
	if len(r.Result) < 1 && !retry {
		<-time.After(5 * time.Second) // 5s 重试
		return n.GetMempoolCount(true)
	}
	n.GetMempoolErrorTimes = 0
	return int64(len(r.Result)), nil
}

func (n *Node) GetBlockByOrder(order int64) (*BlockOrderResult, error) {
	b, err := n.rpcResult("getBlockByOrder", []interface{}{order, true})
	if err != nil {
		n.ErrorMsg("getBlockByOrder Exception", err)
		n.GetBlockByOrderErrorTimes++
		return nil, err
	}
	var r BlockOrderResult
	err = json.Unmarshal(b, &r)
	if err != nil {
		n.ErrorMsg("getBlockByOrder Unmarshal Exception", err)
		n.GetBlockByOrderErrorTimes++
		return nil, err
	}
	if r.Result.Height < 1 {
		n.GetBlockByOrderErrorTimes++
		return nil, fmt.Errorf("getBlockByOrder %d Result Exception", order)
	}
	n.GetBlockByOrderErrorTimes = 0
	return &r, nil
}

func (n *Node) GetStateRoot(order int64) (*StateRootResult, error) {
	b, err := n.rpcResult("getStateRoot", []interface{}{order, true})
	if err != nil {
		n.ErrorMsg("getStateRoot Exception", err)
		n.GetStateRootErrorTimes++
		return nil, err
	}
	var r StateRootResult
	err = json.Unmarshal(b, &r)
	if err != nil {
		n.ErrorMsg("getStateRoot Unmarshal Exception", err)
		n.GetStateRootErrorTimes++
		return nil, err
	}
	if r.Result.Height < 1 {
		n.GetStateRootErrorTimes++
		return nil, fmt.Errorf("getStateRoot %d Result Exception", order)
	}
	n.GetStateRootErrorTimes = 0
	return &r, nil
}

func (n *Node) ResetPeer() error {
	b, err := n.rpcResultLong("p2p_resetPeers", []interface{}{})
	if err != nil {
		n.ErrorMsg("p2p_resetPeers Exception", err)
		return err
	}
	var r ResetPeersResult
	err = json.Unmarshal(b, &r)
	if err != nil {
		n.ErrorMsg("p2p_resetPeers Unmarshal Exception", err)
		return err
	}
	return nil
}
