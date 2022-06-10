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

func launchWorkers(c ClustersToRun) {
	runCluster := func(clusterConf ClusterConfig) {
		for i := 0; i < clusterConf.PingConfig.NumWorkers; i++ {
			log.Println("	go pingDataWorker", clusterConf.Cluster, " n:", clusterConf.PingConfig.NumWorkers)
			go pingDataWorker(clusterConf)
			time.Sleep(2 * time.Second)
		}
	}
	switch c {
	case RunMainnetBeta:
		runCluster(config.Mainnet)
	case RunTestnet:
		runCluster(config.Testnet)
	case RunDevnet:
		runCluster(config.Devnet)
	case RunAllClusters:
		runCluster(config.Mainnet)
		runCluster(config.Testnet)
		runCluster(config.Devnet)
	default:
		panic(InvalidCluster)
	}
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
