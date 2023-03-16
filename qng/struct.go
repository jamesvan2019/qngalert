package qng

import "sync"

type BlockCountResult struct {
	Result int64 `json:"result"`
}

type BlockOrderResult struct {
	Result struct {
		Timestamp    string        `json:"timestamp"`
		Height       int64         `json:"height"`
		Order        int64         `json:"order"`
		TxValid      bool          `json:"txsvalid"`
		Transactions []interface{} `json:"transactions"`
	} `json:"result"`
}

type StateRoot struct {
	Hash      string `json:"Hash"`
	Order     int    `json:"Order"`
	Height    int    `json:"Height"`
	Valid     bool   `json:"Valid"`
	StateRoot string `json:"StateRoot"`
	Number    int    `json:"Number"`
	Node      string `json:"-"`
}

type StateRootResult struct {
	Result StateRoot `json:"result"`
}

type StateRootObjStruct struct {
	StateRoots    map[int64]StateRoot
	StateRootsArr []int64
	lock          sync.Mutex
}

var StateRootObj StateRootObjStruct
