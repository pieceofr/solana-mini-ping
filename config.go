package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ConnectionMode is a type of client connection
type ConnectionMode string

const (
	HTTP  ConnectionMode = "http"
	HTTPS ConnectionMode = "https"
	BOTH  ConnectionMode = "both"
)

type SolanaConfig struct {
	JsonRPCURL    string
	WebsocketURL  string
	KeypairPath   string
	AddressLabels map[string]string
	Commitment    string
}

type SolanaConfigInfo struct {
	Dir           string
	MainnetPath   string
	TestnetPath   string
	DevnetPath    string
	ConfigMain    SolanaConfig
	ConfigTestnet SolanaConfig
	ConfigDevnet  SolanaConfig
}

type PingConfig struct {
	Receiver                string
	NumWorkers              int
	BatchCount              int
	BatchInverval           int
	TxTimeout               int64
	WaitConfirmationTimeout int64
	StatusCheckInterval     int64
	MinPerPingTime          int64
	MaxPerPingTime          int64
	RequestUnits            uint32
	ComputeUnitPrice        uint32
	TxLogOn                 bool
}

type EndPoint struct {
	Mainnet string
	Testnet string
	Devnet  string
}

type SolanaPing struct {
	AlternativeEnpoint EndPoint
	Clusters           []Cluster
	PingConfig
}

type Config struct {
	HostName string
	SolanaConfigInfo
	SolanaPing
}

func loadConfig() Config {
	c := Config{}
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("./")
	v.ReadInConfig()
	v.AutomaticEnv()
	host, err := os.Hostname()
	if err != nil {
		c.HostName = ""
	}

	c.HostName = host
	c.SolanaConfigInfo = SolanaConfigInfo{
		Dir:         v.GetString("SolanaConfig.Dir"),
		MainnetPath: v.GetString("SolanaConfig.MainnetPath"),
		TestnetPath: v.GetString("SolanaConfig.TestnetPath"),
		DevnetPath:  v.GetString("SolanaConfig.DevnetPath"),
	}
	if len(c.SolanaConfigInfo.MainnetPath) > 0 {
		sConfig, err := ReadSolanaConfigFile(c.SolanaConfigInfo.Dir + c.SolanaConfigInfo.MainnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.SolanaConfigInfo.ConfigMain = sConfig
	}
	if len(c.SolanaConfigInfo.TestnetPath) > 0 {
		sConfig, err := ReadSolanaConfigFile(c.SolanaConfigInfo.Dir + c.SolanaConfigInfo.TestnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.SolanaConfigInfo.ConfigTestnet = sConfig
	}
	if len(c.SolanaConfigInfo.DevnetPath) > 0 {
		sConfig, err := ReadSolanaConfigFile(c.SolanaConfigInfo.Dir + c.SolanaConfigInfo.DevnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.SolanaConfigInfo.ConfigDevnet = sConfig
	}
	// SolanaPing
	c.SolanaPing = SolanaPing{
		AlternativeEnpoint: EndPoint{
			Mainnet: v.GetString("SolanaPing.AlternativeEnpoint.Mainnet"),
			Testnet: v.GetString("SolanaPing.AlternativeEnpoint.Testnet"),
			Devnet:  v.GetString("SolanaPing.AlternativeEnpoint.Devnet"),
		},
		PingConfig: PingConfig{
			Receiver:                v.GetString("SolanaPing.PingConfig.Receiver"),
			NumWorkers:              v.GetInt("SolanaPing.PingConfig.NumWorkers"),
			BatchCount:              v.GetInt("SolanaPing.PingConfig.BatchCount"),
			BatchInverval:           v.GetInt("SolanaPing.PingConfig.BatchInverval"),
			TxTimeout:               v.GetInt64("SolanaPing.PingConfig.TxTimeout"),
			WaitConfirmationTimeout: v.GetInt64("SolanaPing.PingConfig.WaitConfirmationTimeout"),
			StatusCheckInterval:     v.GetInt64("SolanaPing.PingConfig.StatusCheckInterval"),
			MinPerPingTime:          v.GetInt64("SolanaPing.PingConfig.MinPerPingTime"),
			MaxPerPingTime:          v.GetInt64("SolanaPing.PingConfig.MaxPerPingTime"),
			RequestUnits:            v.GetUint32("SolanaPing.PingConfig.RequestUnits"),
			ComputeUnitPrice:        v.GetUint32("SolanaPing.PingConfig.ComputeUnitPrice"),
			TxLogOn:                 v.GetBool("SolanaPing.PingConfig.TxLogOn"),
		},
	}
	c.SolanaPing.Clusters = []Cluster{}
	for _, e := range v.GetStringSlice("SolanaPing.Clusters") {
		c.SolanaPing.Clusters = append(c.SolanaPing.Clusters, Cluster(e))
	}
	// SlackReport
	sCluster := []Cluster{}
	for _, e := range v.GetStringSlice("SlackReport.Clusters") {
		sCluster = append(sCluster, Cluster(e))
	}

	return c
}

func ReadSolanaConfigFile(filepath string) (SolanaConfig, error) {
	configmap := make(map[string]string, 1)
	addressmap := make(map[string]string, 1)

	f, err := os.Open(filepath)
	if err != nil {
		log.Printf("error opening file: %v\n", err)
		return SolanaConfig{}, err
	}
	r := bufio.NewReader(f)
	line, _, err := r.ReadLine()
	for err == nil {
		k, v := ToKeyPair(string(line))
		if k == "address_labels" {
			line, _, err := r.ReadLine()
			lKey, lVal := ToKeyPair(string(line))
			if err == nil && string(line)[0:1] == " " {
				if len(lKey) > 0 && len(lVal) > 0 {
					addressmap[lKey] = lVal
				}
			} else {
				configmap[k] = v
			}
		} else {
			configmap[k] = v
		}

		line, _, err = r.ReadLine()
	}
	return SolanaConfig{
		JsonRPCURL:    configmap["json_rpc_url"],
		WebsocketURL:  configmap["websocket_url:"],
		KeypairPath:   configmap["keypair_path"],
		AddressLabels: addressmap,
		Commitment:    configmap["commitment"],
	}, nil
}

func ToKeyPair(line string) (key string, val string) {
	noSpaceLine := strings.TrimSpace(string(line))
	idx := strings.Index(noSpaceLine, ":")
	if idx == -1 || idx == 0 { // not found or only have :
		return "", ""
	}
	if (len(noSpaceLine) - 1) == idx { // no value
		return strings.TrimSpace(noSpaceLine[0:idx]), ""
	}
	return strings.TrimSpace(noSpaceLine[0:idx]), strings.TrimSpace(noSpaceLine[idx+1:])
}
