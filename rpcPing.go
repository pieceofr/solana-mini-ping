package main

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/lib/pq"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/types"
)

type TakeTime struct {
	Times []int64
	Start int64
	End   int64
}

//PingResult is a struct to store ping result and database structure
type PingResult struct {
	TimeStamp int64 `gorm:"autoIncrement:false"`
	Cluster   string
	Hostname  string
	PingType  string `gorm:"NOT NULL"`
	Submitted int    `gorm:"NOT NULL"`
	Confirmed int    `gorm:"NOT NULL"`
	Loss      float64
	Max       int64
	Mean      int64
	Min       int64
	Stddev    int64
	TakeTime  int64
	Error     pq.StringArray `gorm:"type:text[];"NOT NULL"`
	CreatedAt time.Time      `gorm:"type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP" json:"created_at,omitempty"`
	UpdatedAt time.Time      `gorm:"type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP" json:"updated_at,omitempty"`
}

func Ping(cluster Cluster, c *client.Client, host string, pType PingType, config PingConfig) (PingResult, error) {
	var configAcct types.Account
	resultErrs := []string{}
	timer := TakeTime{}
	result := PingResult{
		Cluster:  string(cluster),
		Hostname: host,
		PingType: string(pType),
	}

	configAcct, err := getConfigKeyPair(cluster)
	if err != nil {
		result.Error = resultErrs
		return result, err
	}
	confirmedCount := 0
	for i := 0; i < config.BatchCount; i++ {
		if i > 0 {
			time.Sleep(time.Duration(config.BatchInverval))
		}
		timer.TimerStart()
		var hash string
		if cluster == Testnet {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TxTimeout)*time.Second)
			defer cancel()
			hash, err = SendPingTx(SendPingTxParam{
				Client:              c,
				Ctx:                 ctx,
				FeePayer:            configAcct,
				RequestComputeUnits: config.RequestUnits,
				ComputeUnitPrice:    config.ComputeUnitPrice,
			})
			if err != nil {
				timer.TimerStop()
				log.Println("SendPingTx error:", err)
				resultErrs = append(resultErrs, err.Error())
				continue
			}
			log.Println("SendPingTx:", hash)
		} else {
			hash, err = Transfer(c, configAcct, configAcct, config.Receiver, time.Duration(config.TxTimeout)*time.Second)
			if err != nil {
				timer.TimerStop()
				resultErrs = append(resultErrs, err.Error())
				continue
			}
		}
		err = waitConfirmation(c, hash, time.Duration(config.WaitConfirmationTimeout)*time.Second, time.Duration(config.TxTimeout)*time.Second, time.Duration(config.StatusCheckInterval)*time.Second)
		timer.TimerStop()
		if err != nil {
			resultErrs = append(resultErrs, err.Error())
			log.Println("waitConfirmation error:", err)
			continue
		}
		timer.Add()
		confirmedCount++
	}
	result.TimeStamp = time.Now().UTC().Unix()
	result.Submitted = config.BatchCount
	result.Confirmed = confirmedCount
	result.Loss = (float64(result.Submitted-result.Confirmed) / float64(result.Submitted)) * 100
	max, mean, min, stdDev, total := timer.Statistic()
	result.Max = max
	result.Mean = int64(mean)
	result.Min = min
	result.Stddev = int64(stdDev)
	result.TakeTime = total
	result.Error = resultErrs
	return result, nil
}

func (t *TakeTime) TimerStart() {
	t.Start = time.Now().UTC().UnixMilli()
}

func (t *TakeTime) TimerStop() {
	t.End = time.Now().UTC().UnixMilli()
}

func (t *TakeTime) Add() {
	t.Times = append(t.Times, (t.End - t.Start))
}

func (t *TakeTime) AddTime(ts int64) {
	t.Times = append(t.Times, ts)
}

func (t *TakeTime) TotalTime() int64 {
	sum := int64(0)
	for _, ts := range t.Times {
		sum += ts
	}
	return sum
}

func (t *TakeTime) Statistic() (max int64, mean float64, min int64, stddev float64, sum int64) {
	count := 0
	for _, ts := range t.Times {
		if ts <= 0 { // do not use 0 data because it is the bad data
			continue
		}
		if max == 0 {
			max = ts
		}
		if min == 0 {
			min = ts
		}
		if ts >= max {
			max = ts
		}
		if ts <= min {
			min = ts
		}
		sum += ts
		count++

	}
	if count > 0 {
		mean = float64(sum) / float64(count)
		for _, ts := range t.Times {
			if ts > 0 { // if ts = 0 , ping fail.
				stddev += math.Pow(float64(ts)-mean, 2)
			}
		}
		stddev = math.Sqrt(stddev / float64(count))
	}
	return
}
