package qng

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"qngalert/config"
	"qngalert/notify"
	"time"
)

type Node struct {
	Cfg                       *config.Node
	GetBlockCountErrorTimes   int64
	MempoolEmptyTimes         int64
	GetPeersErrorTimes        int64
	GetMempoolErrorTimes      int64
	GetBlockByOrderErrorTimes int64
	GetStateRootErrorTimes    int64
	GetMinerErrorTimes        int64
	LastestOrder              int64
	NotifyClients             notify.Clients
	ReqTimes                  int64
	lastReset                 int64
	zhangben                  int64
}

func (n *Node) Init(cfg config.Node, ns notify.Clients) error {
	n.Cfg = &cfg
	n.NotifyClients = ns
	return nil
}

func (n *Node) ErrorMsg(str string, err error) {
	log.Println("[node]", n.Cfg.Rpc, "msg", str, "err", err)
}

func (n *Node) ErrorMsgFormat(str string, err error) string {
	return fmt.Sprintf("[node]%s [msg] %s [err] %s", n.Cfg.Rpc, str, err.Error())
}

func (n *Node) Msg(str string) {
	log.Println(fmt.Sprintf("[node]%s [msg]%s", n.Cfg.Rpc, str))
}

func (n *Node) rpcResult(method string, params []interface{}) ([]byte, error) {
	// Create the new HTTP client that is configured according to the user-
	// specified options and submit the request.
	return n.reqNode(time.Duration(10), method, params)
}

func (n *Node) rpcResultLong(method string, params []interface{}) ([]byte, error) {
	// Create the new HTTP client that is configured according to the user-
	// specified options and submit the request.
	return n.reqNode(time.Duration(3600), method, params)
}

func (n *Node) reqNode(timeout time.Duration, method string, params []interface{}) ([]byte, error) {
	paramStr, err := json.Marshal(params)
	if err != nil {
		n.ErrorMsg("rpc params error:", err)
		return nil, err
	}
	id := rand.Int31()
	jsonStr := []byte(fmt.Sprintf(`{"jsonrpc": "2.0", "method": "%s", "params":%s, "id": %d}`, method, string(paramStr), id))
	bodyBuff := bytes.NewBuffer(jsonStr)
	httpRequest, err := http.NewRequest("POST", n.Cfg.Rpc, bodyBuff)
	if err != nil {
		n.ErrorMsg("rpc connect failed", err)
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Connection", "close")
	// Configure basic access authorization.
	httpRequest.SetBasicAuth(n.Cfg.User, n.Cfg.Pass)
	httpClient := http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			ForceAttemptHTTP2: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
	}
	httpClient.Timeout = timeout * time.Second
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		n.ErrorMsg("rpc request faild", err)
		return nil, err
	}
	defer func() {
		_ = httpResponse.Body.Close()
	}()
	body, err := io.ReadAll(httpResponse.Body)

	if err != nil {
		n.ErrorMsg("error reading json reply:", err)
		return nil, err
	}

	if httpResponse.Status != "200 OK" {
		err = fmt.Errorf("%s:%s", httpResponse.Status, body)
		n.ErrorMsg("error http response", err)
		return nil, err
	}
	return body, nil
}
