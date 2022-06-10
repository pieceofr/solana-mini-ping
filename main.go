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
	log.Println("ClusterCLIConfig:", config.ClusterCLIConfig)
	log.Println("Mainnet Config:", config.Mainnet)
	log.Println("Testnet Config:", config.Mainnet)
	log.Println("Devnet Config:", config.Devnet)

}

func main() {
	go launchWorkers()
	for {
		time.Sleep(10 * time.Second)
	}
}
