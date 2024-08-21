package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
	"yeelight/lib"
)

const (
	MaxDatagramSize = 8192
)

func main() {
	if err := getConfigurations(); err != nil {
		panic(err)
	}

	udpAddr := lib.Config.UdpAddr
	stopChan := make(chan struct{})
	doneChan := make(chan struct{})

	for {
		go lib.ListenUDPWithStop(udpAddr, lib.MsgHandler, stopChan, doneChan)

		time.Sleep(time.Duration(lib.Config.RefreshTimer) * time.Minute)

		close(stopChan)

		<-doneChan
		err := lib.PeriodicExecuteScenarios()
		if err != nil {
			fmt.Println("Error executing scenarios:", err)
		}

		stopChan = make(chan struct{})
		doneChan = make(chan struct{})
	}
}

func getConfigurations() (err error) {
	yamlFile, err := os.ReadFile("./config/scenarios.yaml")
	err = yaml.Unmarshal(yamlFile, &lib.Scenarios)

	yamlFile, err = os.ReadFile("./config/main.yaml")
	err = yaml.Unmarshal(yamlFile, &lib.Config)

	return
}
