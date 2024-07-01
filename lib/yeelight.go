package lib

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
)

type command struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type response struct {
	ID     int      `json:"id"`
	Result []string `json:"result"`
}

type Method string

var (
	MethodSetCTABX      Method = "set_ct_abx"
	MethodSetRGB        Method = "set_rgb"
	MethodSetHSV        Method = "set_hsv"
	MethodSetBrightness Method = "set_bright"
	MethodSetPower      Method = "set_power"
	MethodToggle        Method = "toggle"
)

func (m *Method) String() string {
	if m == nil {
		return ""
	}
	return string(*m)
}

type Bulb struct {
	mu    sync.Mutex
	cmdID int
	conn  net.Conn
}

func NewBulb(address string) (*Bulb, error) {
	if !strings.Contains(address, ":") {
		address = address + ":55443"
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("could not dial address: %+v", err)
	}
	return &Bulb{
		conn: conn,
	}, nil
}

func (b *Bulb) Send(method Method, args ...interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	cmd := command{
		ID:     b.cmdID,
		Method: method.String(),
		Params: args,
	}

	err := json.NewEncoder(b.conn).Encode(cmd)
	if err != nil {
		return fmt.Errorf("cannot write json: %+v", err)
	}

	_, err = fmt.Fprint(b.conn, "\r\n")
	if err != nil {
		return fmt.Errorf("cannot write trailer: %+v", err)
	}

	var resp response
	err = json.NewDecoder(b.conn).Decode(&resp)
	if err != nil {
		return fmt.Errorf("receiving response: %+v", err)
	}

	b.cmdID++
	return nil
}

func (b *Bulb) TurnOn() error {
	return b.Send(MethodSetPower, "on")
}

func (b *Bulb) TurnOff() error {
	return b.Send(MethodSetPower, "off")
}

func (b *Bulb) ColorTemp(temp int) error {
	switch {
	case temp < 1700:
		temp = 1700
	case temp > 6500:
		temp = 6500
	}
	return b.Send(MethodSetCTABX, temp)
}

func (b *Bulb) RGB(red, green, blue int) error {
	return b.Send(MethodSetRGB, red<<16+green<<8+blue)
}

func (b *Bulb) Brightness(brightness int) error {
	switch {
	case brightness > 100:
		brightness = 100
	case brightness < 1:
		brightness = 1
	}
	return b.Send(MethodSetBrightness, brightness)
}
