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
	// for _, c := range config.ClusterPing.Clusters {
	// 	for i := 0; i < config.PingConfig.NumWorkers; i++ {
	// 		go pingDataWorker(c)
	// 		time.Sleep(5 * time.Second)
	// 	}

	// }
	go pingDataWorker(config.Devnet)
}

func createRPCClient(config ClusterConfig) *client.Client {
	var c *client.Client
	if len(config.AlternativeEnpoint) > 0 {
		c = client.NewClient(config.AlternativeEnpoint)
		log.Println(c, " use alternative endpoint:", config.AlternativeEnpoint)
	} else {
		c = client.NewClient(rpc.TestnetRPCEndpoint)
	}
	return c
}

func pingDataWorker(config ClusterConfig) {
	log.Println(">> Solana DataPoint1MinWorker for ", config.Cluster, " start!")
	defer log.Println(">> Solana DataPoint1MinWorker for ", config.Cluster, " end!")
	c := createRPCClient(config)
	for {
		if c == nil {
			c = createRPCClient(config)
		}
		result, err := Ping(c, DataPoint1Min, config)
		if err != nil {
			continue
		}
		waitTime := config.ClusterPing.PingConfig.MinPerPingTime - (result.TakeTime / 1000)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}
}
