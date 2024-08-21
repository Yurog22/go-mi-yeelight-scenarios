package lib

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	MaxDatagramSize = 8192
)

type BulbData struct {
	Location         url.URL `json:"location"`
	ColorTemperature int     `json:"colorTemperature"`
	ColorMode        int     `json:"colorMode"`
	Brightness       int     `json:"brightness"`
}

func SendAndReceiveUDPMessage(localAddr string, port int, remoteAddr string, remotePort int, handler func(*net.UDPAddr, int, []byte)) error {
	localUDPAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", localAddr, port))
	if err != nil {
		return fmt.Errorf("failed to resolve local UDP address: %v", err)
	}

	remoteUDPAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", remoteAddr, remotePort))
	if err != nil {
		return fmt.Errorf("failed to resolve remote UDP address: %v", err)
	}
	fmt.Println("periodic udp here 3")

	conn, err := net.ListenUDP("udp4", localUDPAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port: %v", err)
	}
	defer conn.Close()

	message := "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1982\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"ST: wifi_bulb\r\n"

	_, err = conn.WriteToUDP([]byte(message), remoteUDPAddr)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	buffer := make([]byte, 1024)

	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return fmt.Errorf("failed to set read deadline: %v", err)
	}

	numBytes, src, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return fmt.Errorf("failed to read from UDP: %v", err)
	}

	handler(src, numBytes, buffer)

	return nil
}

func ParseResponse(data []byte) (bd BulbData, err error) {
	rawResponse := string(data)
	scanner := bufio.NewScanner(strings.NewReader(rawResponse))
	rightHeaderPassed := false

	rightHeaders := []string{"NOTIFY * HTTP/1.1", "HTTP/1.1 200 OK"}

	for scanner.Scan() {
		line := scanner.Text()

		if !rightHeaderPassed {
			rightHeaderPassed = slices.Contains(rightHeaders, line)
			continue
		}

		if rightHeaderPassed {
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
