package lib

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	MaxDatagramSize   = 8192
	MessageBounceTime = 30
)

type BulbData struct {
	Location         url.URL `json:"location"`
	ColorTemperature int     `json:"colorTemperature"`
	ColorMode        int     `json:"colorMode"`
	Brightness       int     `json:"brightness"`
}

var LastMessageTime time.Time

func Listen(address string, handler func(*net.UDPAddr, int, []byte)) {

	fmt.Println("Waiting for bounce time...")
	<-time.After(MessageBounceTime * time.Second)
	fmt.Println("Ready")

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
		buffer := make([]byte, MaxDatagramSize)
		numBytes, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		handler(src, numBytes, buffer)
	}
}

func MsgHandler(src *net.UDPAddr, n int, b []byte) {
	currentMessageTime := time.Now()

	if currentMessageTime.Sub(LastMessageTime).Seconds() > MessageBounceTime {
		LastMessageTime = time.Now()
		var multicastResponse BulbData
		multicastResponse, err := parseResponse(b)
		if err != nil {
			fmt.Println(err.Error())
		}

		if err := RefreshScenarioState(multicastResponse); err != nil {
			fmt.Println(err.Error())
		}
	}
}

func parseResponse(data []byte) (bd BulbData, err error) {
	rawResponse := string(data)
	scanner := bufio.NewScanner(strings.NewReader(rawResponse))
	notifyHeaderPassed := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "NOTIFY * HTTP/1.1" && !notifyHeaderPassed {
			notifyHeaderPassed = true
			continue
		}

		if notifyHeaderPassed {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				switch argument := parts[0]; argument {
				case "Location":
					hostAddress, parseErr := url.Parse(parts[1])
					if parseErr != nil {
						err = parseErr
						return
					}
					bd.Location.Host = hostAddress.Hostname()
				case "ct":
					parseInt, parseErr := strconv.Atoi(parts[1])
					if parseErr != nil {
						err = parseErr
						return
					}
					bd.ColorTemperature = parseInt
				case "bright":
					parseInt, parseErr := strconv.Atoi(parts[1])
					if parseErr != nil {
						err = parseErr
						return
					}
					bd.Brightness = parseInt
				case "color_mode":
					parseInt, parseErr := strconv.Atoi(parts[1])
					if parseErr != nil {
						err = parseErr
						return
					}
					bd.ColorMode = parseInt
				}
			}
		}
	}
	return
}
