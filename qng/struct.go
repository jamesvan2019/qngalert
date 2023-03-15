package qng

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
