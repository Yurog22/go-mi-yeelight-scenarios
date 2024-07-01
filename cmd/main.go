package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
	"yeelight/lib"
)

var config mainConfig

type mainConfig struct {
	UdpAddr string `yaml:"udpAddress"`
}

func main() {
	if err := getConfigurations(); err != nil {
		panic(err)
	}
	lib.LastMessageTime = time.Now()

	go lib.Listen(config.UdpAddr, lib.MsgHandler)

	go func() {
		for {
			if err := lib.PeriodicExecuteScenarios(); err != nil {
				fmt.Println(err.Error())
			}
			time.Sleep(10 * time.Minute)
		}
	}()

	select {}

}

func getConfigurations() (err error) {
	yamlFile, err := os.ReadFile("./config/scenarios.yaml")
	err = yaml.Unmarshal(yamlFile, &lib.Scenarios)

	yamlFile, err = os.ReadFile("./config/main.yaml")
	err = yaml.Unmarshal(yamlFile, &config)

	return
}
