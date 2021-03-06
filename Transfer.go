package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/cmptbdgprog"
	"github.com/portto/solana-go-sdk/program/memoprog"
	"github.com/portto/solana-go-sdk/program/sysprog"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
)

var (
	txTimeoutDefault               = 10 * time.Second
	waitConfirmationTimeoutDefault = 50 * time.Second
	statusCheckTimeDefault         = 1 * time.Second
)

func Transfer(c *client.Client, sender types.Account, feePayer types.Account, receiverPubkey string, txTimeout time.Duration) (txHash string, err error) {
	// to fetch recent blockhash
	res, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		//log.Println("get recent block hash error, err:", err)
		return "", err
	}
	// create a message
	message := types.NewMessage(types.NewMessageParam{
		FeePayer:        feePayer.PublicKey,
		RecentBlockhash: res.Blockhash, // recent blockhash
		Instructions: []types.Instruction{
			sysprog.Transfer(sysprog.TransferParam{
				From:   sender.PublicKey,                           // from
				To:     common.PublicKeyFromString(receiverPubkey), // to
				Amount: 1,                                          //  SOL
			}),
		},
	})
	log.Println("tx message:", message)
	// create tx by message + signer
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{feePayer, sender},
	})

	if err != nil {
		return "", err
	}
	// send tx
	if txTimeout <= 0 {
		txTimeout = time.Duration(txTimeoutDefault)
	}
	ctx, _ := context.WithTimeout(context.TODO(), txTimeout)
	txHash, err = c.SendTransaction(ctx, tx)

	if err != nil {
		return "", err
	}
	return txHash, nil
}

type SendPingTxParam struct {
	Client              *client.Client
	Ctx                 context.Context
	FeePayer            types.Account
	RequestComputeUnits uint32
	ComputeUnitPrice    uint32
}

func SendPingTx(param SendPingTxParam) (string, error) {
	latestBlockhashResponse, err := param.Client.GetLatestBlockhash(param.Ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get latest blockhash, err: %v", err)
	}

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{param.FeePayer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        param.FeePayer.PublicKey,
			RecentBlockhash: latestBlockhashResponse.Blockhash,
			Instructions: []types.Instruction{
				cmptbdgprog.RequestUnits(cmptbdgprog.RequestUnitsParam{
					Units:         param.RequestComputeUnits,
					AdditionalFee: (param.RequestComputeUnits * param.ComputeUnitPrice) / 1_000_000,
				}),
				memoprog.BuildMemo(memoprog.BuildMemoParam{
					Memo: []byte("ping"),
				}),
			},
		}),
	})
	if err != nil {
		return "", fmt.Errorf("failed to new a tx, err: %v", err)
	}

	txhash, err := param.Client.SendTransaction(param.Ctx, tx)
	if err != nil {
		return "", fmt.Errorf("failed to send a tx, err: %v", err)
	}

	return txhash, nil
}

func waitConfirmation(c *client.Client, txHash string, timeout time.Duration, requestTimeout time.Duration, queryTime time.Duration) error {
	if timeout <= 0 {
		timeout = waitConfirmationTimeoutDefault
		log.Println("timeout is not set! Use default timeout", timeout, " sec")
	}

	ctx, _ := context.WithTimeout(context.TODO(), requestTimeout)
	elapse := time.Now()
	for {
		resp, err := c.GetSignatureStatus(ctx, txHash)
		now := time.Now()
		if err != nil {
			if now.Sub(elapse).Seconds() < timeout.Seconds() {
				continue
			} else {
				return err
			}
		}
		if resp != nil {
			if *resp.ConfirmationStatus == rpc.CommitmentConfirmed || *resp.ConfirmationStatus == rpc.CommitmentFinalized {
				log.Println(txHash, " confirmed/finalized")
				return nil
			}
		}
		if now.Sub(elapse).Seconds() > timeout.Seconds() {
			return err
		}

		if queryTime <= 0 {
			queryTime = statusCheckTimeDefault
		}
		time.Sleep(queryTime)
	}
}

func getConfigKeyPair(cluster Cluster) (types.Account, error) {
	var c SolanaCLIConfig
	switch cluster {
	case MainnetBeta:
		c = config.ClusterCLIConfig.ConfigMain
	case Testnet:
		c = config.ClusterCLIConfig.ConfigTestnet
	case Devnet:
		c = config.ClusterCLIConfig.ConfigDevnet
	default:
		log.Println("StatusNotFound Error:", cluster)
		return types.Account{}, errors.New("Invalid Cluster")
	}
	body, err := ioutil.ReadFile(c.KeypairPath)
	if err != nil {

	}
	key := []byte{}
	err = json.Unmarshal(body, &key)
	if err != nil {
		return types.Account{}, err
	}

	acct, err := types.AccountFromBytes(key)
	if err != nil {
		return types.Account{}, err
	}
	return acct, nil

}
