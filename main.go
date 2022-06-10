package main

import (
	"flag"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"gorm.io/gorm"
)

var config Config

//Cluster enum
type Cluster string

var database *gorm.DB
var dbMtx *sync.Mutex

const useGCloudDB = true

//Cluster enum
const (
	MainnetBeta Cluster = "MainnetBeta"
	Testnet             = "Testnet"
	Devnet              = "Devnet"
)

var userInputClusterMode string

type ClustersToRun string

const (
	RunMainnetBeta ClustersToRun = "MainnetBeta"
	RunTestnet                   = "Testnet"
	RunDevnet                    = "Devnet"
	RunAllClusters               = "All"
)

func init() {

	flag.StringVar(&userInputClusterMode,
		"run-cluster-mode",
		RunAllClusters,
		"specify which cluster (MainnetBeta/Testnet/Devnet/All) to run.")

	log.Println("--- Config Start --- ")
	config = loadConfig()
	log.Println("ClusterCLIConfig Mainnet:", config.ClusterCLIConfig.ConfigMain)
	log.Println("ClusterCLIConfig Testnet:", config.ClusterCLIConfig.ConfigTestnet)
	log.Println("ClusterCLIConfig Devnet:", config.ClusterCLIConfig.ConfigDevnet)
	log.Println("Mainnet Config:", config.Mainnet)
	log.Println("Testnet Config:", config.Testnet)
	log.Println("Devnet Config:", config.Devnet)

}

func main() {
	flag.Parse()
	clustersToRun := flag.Arg(0)
	if !(strings.Compare(clustersToRun, string(RunMainnetBeta)) == 0 ||
		strings.Compare(clustersToRun, string(RunTestnet)) == 0 ||
		strings.Compare(clustersToRun, string(RunDevnet)) == 0) {
		clustersToRun = RunAllClusters
	}
	log.Println("Ping Service will run clusters:", clustersToRun)
	go launchWorkers(ClustersToRun(clustersToRun))
	for {
		time.Sleep(10 * time.Second)
	}
}
