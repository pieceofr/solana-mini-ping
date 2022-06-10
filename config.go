package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// ConnectionMode is a type of client connection
type ConnectionMode string

const (
	HTTP  ConnectionMode = "http"
	HTTPS ConnectionMode = "https"
	BOTH  ConnectionMode = "both"
)

type SolanaCLIConfig struct {
	JsonRPCURL    string
	WebsocketURL  string
	KeypairPath   string
	AddressLabels map[string]string
	Commitment    string
}

type ClusterCLIConfig struct {
	Dir           string
	MainnetPath   string
	TestnetPath   string
	DevnetPath    string
	ConfigMain    SolanaCLIConfig
	ConfigTestnet SolanaCLIConfig
	ConfigDevnet  SolanaCLIConfig
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
type ClusterPing struct {
	AlternativeEnpoint string
	Clusters           []Cluster
	PingConfig
}

type ClusterConfig struct {
	Cluster
	HostName string
	ClusterPing
}

type Config struct {
	Mainnet ClusterConfig
	Testnet ClusterConfig
	Devnet  ClusterConfig
	ClusterCLIConfig
}

func loadConfig() Config {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelTrace)
	c := Config{}
	v := viper.New()
	v.AddConfigPath("./")
	v.AutomaticEnv()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	// setup config.yaml
	v.SetConfigName("config")
	v.ReadInConfig()
	c.ClusterCLIConfig = ClusterCLIConfig{
		Dir:         v.GetString("SolanaCliFile.Dir"),
		MainnetPath: v.GetString("SolanaCliFile.MainnetPath"),
		TestnetPath: v.GetString("SolanaCliFile.TestnetPath"),
		DevnetPath:  v.GetString("SolanaCliFile.DevnetPath"),
	}

	if len(c.ClusterCLIConfig.MainnetPath) > 0 {
		sConfig, err := ReadSolanaCliConfigFile(c.ClusterCLIConfig.Dir + c.ClusterCLIConfig.MainnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.ClusterCLIConfig.ConfigMain = sConfig
	}
	if len(c.ClusterCLIConfig.TestnetPath) > 0 {
		sConfig, err := ReadSolanaCliConfigFile(c.ClusterCLIConfig.Dir + c.ClusterCLIConfig.TestnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.ClusterCLIConfig.ConfigTestnet = sConfig
	}
	if len(c.ClusterCLIConfig.DevnetPath) > 0 {
		sConfig, err := ReadSolanaCliConfigFile(c.ClusterCLIConfig.Dir + c.ClusterCLIConfig.DevnetPath)
		if err != nil {
			log.Fatal(err)
		}
		c.ClusterCLIConfig.ConfigDevnet = sConfig
	}

	// Read Each Cluster Configurations
	// setup config.yaml for mainnet
	configMainnetFile := v.GetString("ClusterConfigFile.Mainnet")
	v.SetConfigName(configMainnetFile)
	v.ReadInConfig()
	c.Mainnet = ClusterConfig{
		Cluster:     MainnetBeta,
		HostName:    hostname,
		ClusterPing: ReadClusterPingConfig(v),
	}
	configTestnetFile := v.GetString("ClusterConfigFile.Testnet")
	v.SetConfigName(configTestnetFile)
	v.ReadInConfig()
	c.Testnet = ClusterConfig{
		Cluster:     Testnet,
		HostName:    hostname,
		ClusterPing: ReadClusterPingConfig(v),
	}
	configDevnetFile := v.GetString("ClusterConfigFile.Devnet")
	v.SetConfigName(configDevnetFile)
	v.ReadInConfig()
	c.Devnet = ClusterConfig{
		Cluster:     Devnet,
		HostName:    hostname,
		ClusterPing: ReadClusterPingConfig(v),
	}
	return c
}

func ReadSolanaCliConfigFile(filepath string) (SolanaCLIConfig, error) {
	configmap := make(map[string]string, 1)
	addressmap := make(map[string]string, 1)

	f, err := os.Open(filepath)
	if err != nil {
		log.Printf("error opening file: %v\n", err)
		return SolanaCLIConfig{}, err
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
	return SolanaCLIConfig{
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

func ReadClusterPingConfig(v *viper.Viper) ClusterPing {
	return ClusterPing{
		AlternativeEnpoint: v.GetString("AlternativeEnpoint"),
		PingConfig: PingConfig{
			Receiver:                v.GetString("PingConfig.Receiver"),
			NumWorkers:              v.GetInt("PingConfig.NumWorkers"),
			BatchCount:              v.GetInt("PingConfig.BatchCount"),
			BatchInverval:           v.GetInt("PingConfig.BatchInverval"),
			TxTimeout:               v.GetInt64("PingConfig.TxTimeout"),
			WaitConfirmationTimeout: v.GetInt64("PingConfig.WaitConfirmationTimeout"),
			StatusCheckInterval:     v.GetInt64("PingConfig.StatusCheckInterval"),
			MinPerPingTime:          v.GetInt64("PingConfig.MinPerPingTime"),
			MaxPerPingTime:          v.GetInt64("PingConfig.MaxPerPingTime"),
			RequestUnits:            v.GetUint32("PingConfig.RequestUnits"),
			ComputeUnitPrice:        v.GetUint32("PingConfig.ComputeUnitPrice"),
		},
	}
}
