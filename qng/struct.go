package qng

import "sync"

type BlockCountResult struct {
	Result int64 `json:"result"`
}

type PeerInfoResult struct {
	Result []PeerInfo `json:"result"`
}

type NodeInfoResult struct {
	Result NodeInfo `json:"result"`
}

type MempoolResult struct {
	Result []string `json:"result"`
}

type GraphState struct {
	MainOrder int `json:"mainorder"`
}

type PeerInfo struct {
	Address    string     `json:"address"`
	State      bool       `json:"state"`    //connected
	Active     bool       `json:"active"`   //connected
	Version    string     `json:"version"`  //connected
	Services   string     `json:"services"` // Relay | Full|Bloom|CF
	GraphState GraphState `json:"graphstate"`
}

type NodeInfo struct {
	Buildversion string `json:"buildversion"`
	ID           string `json:"ID"`
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

type ResetPeersResult struct {
	Result interface{} `json:"result"`
}

type StateRootObjStruct struct {
	StateRoots    map[int64]StateRoot
	StateRootsArr []int64
	lock          sync.Mutex
}

var StateRootObj StateRootObjStruct
