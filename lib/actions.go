package lib

import (
	"fmt"
	"reflect"
	"time"
)

type Scenario struct {
	HostAddress string   `json:"hostAddress" yaml:"hostAddress"`
	Actions     []Action `json:"actions" yaml:"actions"`
}

type Action struct {
	Timer            string `json:"timer" yaml:"timer"`
	ColorTemperature int    `json:"colorTemperature" yaml:"colorTemperature"`
	Brightness       int    `json:"brightness" yaml:"brightness"`
}

var Scenarios []Scenario

func RefreshScenarioState(b BulbData) error {
	for _, s := range Scenarios {
		if s.HostAddress == b.Location.Host {
			if err := executeScenario(s); err != nil {
				return err
			}
		}
	}

	return nil
}

func executeScenario(s Scenario) (err error) {
	var currentAction Action
	currentTime := time.Now()

	timeLayout := "15:04"
	for _, a := range s.Actions {
		parsedTime, err := time.Parse(timeLayout, a.Timer)
		if err != nil {
			return err
		}
		parsedTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, currentTime.Location())
		if currentTime.After(parsedTime) {
			currentAction = a
		}
	}

	if !reflect.DeepEqual(currentAction, Action{}) {
		bulb, err := NewBulb(s.HostAddress)
		if err != nil {
			return err
		}

		if err := bulb.ColorTemp(currentAction.ColorTemperature); err != nil {
			return err
		}

		if err := bulb.Brightness(currentAction.Brightness); err != nil {
			return err
		}

		fmt.Println(time.Now(), "applied action:", "color temp:", currentAction.ColorTemperature, "brightness:", currentAction.Brightness)
	}

	return
}

func PeriodicExecuteScenarios() error {
	for _, s := range Scenarios {
		if err := executeScenario(s); err != nil {
			fmt.Println(err.Error())
			continue
		}
	}
	return nil
}
