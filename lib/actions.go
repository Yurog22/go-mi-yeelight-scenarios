package lib

import (
	"context"
	"fmt"
	"log"
	"net"
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

var Config mainConfig

type mainConfig struct {
	UdpAddr      string `yaml:"udpAddress"`
	LocalAddr    string `yaml:"localAddress"`
	RefreshTimer int    `yaml:"refreshTimer"`
}

func RefreshScenarioState(b BulbData) error {
	for _, s := range Scenarios {
		if s.HostAddress == b.Location.Host {
			if err := executeScenario(s, b); err != nil {
				return err
			}
		}
	}

	return nil
}

func executeScenario(s Scenario, b BulbData) (err error) {
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

	fmt.Printf("Current scenario state is: \n  Timer:%s\n  Color temperature: %d\n  Brightness: %d\n\n", currentAction.Timer, currentAction.ColorTemperature, currentAction.Brightness)

	sameState := (b.Brightness == currentAction.Brightness) && (b.ColorTemperature == currentAction.ColorTemperature)

	if !reflect.DeepEqual(currentAction, Action{}) && !sameState {
		bulb, err := NewBulb(s.HostAddress)
		if err != nil {
			fmt.Println(err)
			return err
		}
		if err := bulb.ColorTemp(currentAction.ColorTemperature); err != nil {
			fmt.Println(err)
			return err
		}
		if err := bulb.Brightness(currentAction.Brightness); err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("[APPLIED] %s Color temperature: %d Brightness: %d\n", time.Now().Format("01.02.2006 15:04:05"), currentAction.ColorTemperature, currentAction.Brightness)

		err = bulb.conn.Close()
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return
}

func ListenUDPWithStop(udpAddr string, handler func(*net.UDPAddr, int, []byte), stopChan chan struct{}, doneChan chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-stopChan
		cancel()
	}()

	ListenUDP(udpAddr, handler, ctx)

	close(doneChan)
}

func ListenUDP(address string, handler func(*net.UDPAddr, int, []byte), ctx context.Context) {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadBuffer(MaxDatagramSize)

	for {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		select {
		case <-ctx.Done():
			return
		default:
			buffer := make([]byte, MaxDatagramSize)
			numBytes, src, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					continue
				}
				if ctx.Err() != nil {
					return
				}
				log.Fatal("ReadFromUDP failed:", err)
			}

			handler(src, numBytes, buffer)
		}
	}
}

func PeriodicExecuteScenarios() error {
	err := SendAndReceiveUDPMessage(Config.LocalAddr, 8001, "239.255.255.250", 1982, MsgHandler)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func MsgHandler(src *net.UDPAddr, n int, b []byte) {
	var multicastResponse BulbData
	multicastResponse, err := ParseResponse(b)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Parsed response is: \n  Host:%s\n  Color temperature: %d\n  Brightness: %d\n\n", multicastResponse.Location.Host, multicastResponse.ColorTemperature, multicastResponse.Brightness)

	if err := RefreshScenarioState(multicastResponse); err != nil {
		fmt.Println(err.Error())
	}
}
