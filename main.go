package main

import (
	"log"
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

func init() {
	log.Println("--- Config Start --- ")
	config = loadConfig()
	log.Println("SolanaConfig/Dir:", config.SolanaConfigInfo.Dir, "\n",
		" SolanaConfig/Mainnet", config.SolanaConfigInfo.MainnetPath, "\n",
		" SolanaConfig/Testnet", config.SolanaConfigInfo.TestnetPath, "\n",
		" SolanaConfig/Devnet", config.SolanaConfigInfo.DevnetPath)
	log.Println("SolanaConfigFile/Mainnet:", "\n", config.SolanaConfigInfo.ConfigMain)
	log.Println("SolanaConfigFile/Testnet:", "\n", config.SolanaConfigInfo.ConfigTestnet)
	log.Println("SolanaConfigFile/Devnet:", "\n", config.SolanaConfigInfo.ConfigDevnet)
	log.Println("SolanaPing:", config.SolanaPing)

}

func main() {
	go launchWorkers()
	for {
		time.Sleep(10 * time.Second)
	}
}
