package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"qngalert/notify"
	"strings"
	"sync"
	"time"
)

func SendTransactions(client *ethclient.Client, privkey, to string, input []byte, value *big.Int) string {
	// private key
	privateKey, err := crypto.HexToECDSA(privkey)
	if err != nil {
		log.Println("私钥错误", err)
		return ""
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Println("私钥pubkey错误", err)
		return ""
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Println(err)
		return ""
	}
	fmt.Println(fromAddress.String(), "nonce:", nonce)
	gasV := 300000           //gas limit
	gasLimit := uint64(gasV) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Println(err)
		return ""
	}
	toAddress := common.HexToAddress(to)
	baseTx := &types.LegacyTx{
		To:       &toAddress,
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		Value:    value,
		Data:     input,
	}
	tx := types.NewTx(baseTx)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(813)), privateKey)
	if err != nil {
		log.Println(err)
		return ""
	}
	// return signedTx.Hash().Hex()
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Println(err)
		return ""
	}
	txHash := signedTx.Hash().Hex()
	return txHash
}

func ListenSendTx(ctx context.Context, wg *sync.WaitGroup, notifyClients notify.Clients) {
	defer wg.Done()
	log.Println("start ListenSendTx Service")
	node := "https://evm-dataseed1.meerscan.com"
	cli, err := ethclient.Dial(node)
	if err != nil {
		log.Println(err)
		return
	}
	times := 0
	txids := make([]string, 0)
	for {
		select {
		case <-ctx.Done():
			log.Println("stop ListenSendTx Service,exit...")
			return
		default:
			if times > 10 {
				// 累计错误10个  停止发送
				notifyClients.Send("node Exception", "Can not broadcast txs:"+strings.Join(txids, ",")+" IN https://evm-dataseed1.meerscan.com")
				return
			}
			txid := SendTransactions(cli, "ca39be758c5264b15113f434300e9abee8ecb57e0f533ea2db89ad58703024f9",
				"0x91e9bcE3D9FD5b1aEd8dA2243f5bE15aF149C51a", []byte{}, big.NewInt(1))
			if txid != "" {
				if !CheckTx(ctx, cli, txid) {
					times++
					log.Println(txid, "tx Send Failed! Check node")
					txids = append(txids, txid)
					continue
				}
				txids = make([]string, 0)
				times = 0
				log.Println(txid, "tx Send Succ! ")
			}
		}
	}
}

func CheckTx(ctx context.Context, client *ethclient.Client, tx string) bool {
	times := 0
	for {
		select {
		case <-ctx.Done():
			return false
		default:
			if times >= 30 {
				// 300s 5min
				log.Println(tx, "Send Failed,Not packed within 5m")
				return false
			}
			<-time.After(10 * time.Second)
			times++
			txD, err := client.TransactionReceipt(ctx, common.HexToHash(tx))
			if err != nil {
				log.Println(tx, "TransactionReceipt Not Found Need Wait...")
				continue
			}
			if txD != nil {
				if txD.Status == uint64(0x1) {
					return true
				}
				return false
			}
		}
	}
}
