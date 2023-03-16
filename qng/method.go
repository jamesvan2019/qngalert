package qng

import (
	"encoding/json"
	"fmt"
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
