package main

import (
	"log"
	"time"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
)

type PingType string

const DefaultAlertThredHold = 20
const (
	Report        PingType = "report"
	DataPoint1Min PingType = "datapoint1min"
)

func launchWorkers() {
	for _, c := range config.SolanaPing.Clusters {
		for i := 0; i < config.PingConfig.NumWorkers; i++ {
			go pingDataWorker(c)
			time.Sleep(5 * time.Second)
		}

	}
}

func createRPCClient(cluster Cluster) (*client.Client, error) {
	var c *client.Client
	switch cluster {
	case MainnetBeta:
		if len(config.SolanaPing.AlternativeEnpoint.Mainnet) > 0 {
			c = client.NewClient(config.SolanaPing.AlternativeEnpoint.Mainnet)
			log.Println(c, " use alternative endpoint:", config.SolanaPing.AlternativeEnpoint.Mainnet)
		} else {
			c = client.NewClient(rpc.MainnetRPCEndpoint)
		}
	case Testnet:
		if len(config.SolanaPing.AlternativeEnpoint.Testnet) > 0 {
			c = client.NewClient(config.SolanaPing.AlternativeEnpoint.Testnet)
			log.Println(c, " use alternative endpoint:", config.SolanaPing.AlternativeEnpoint.Testnet)
		} else {
			c = client.NewClient(rpc.TestnetRPCEndpoint)
		}
	case Devnet:
		if len(config.SolanaPing.AlternativeEnpoint.Devnet) > 0 {
			c = client.NewClient(config.SolanaPing.AlternativeEnpoint.Devnet)
			log.Println(c, " use alternative endpoint:", config.SolanaPing.AlternativeEnpoint.Devnet)
		} else {
			c = client.NewClient(rpc.DevnetRPCEndpoint)
		}
	default:
		log.Fatal("Invalid Cluster")
		return nil, InvalidCluster
	}
	return c, nil
}

func pingDataWorker(cluster Cluster) {
	log.Println(">> Solana DataPoint1MinWorker for ", cluster, " start!")
	defer log.Println(">> Solana DataPoint1MinWorker for ", cluster, " end!")
	c, err := createRPCClient(cluster)
	if err != nil {
		return
	}
	for {
		if c == nil {
			c, err = createRPCClient(cluster)
			if err != nil {
				return
			}
		}
		result, err := Ping(cluster, c, config.HostName, DataPoint1Min, config.SolanaPing.PingConfig)
		if err != nil {
			continue
		}
		waitTime := config.SolanaPing.PingConfig.MinPerPingTime - (result.TakeTime / 1000)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}
