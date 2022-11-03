package main

import (
	"bytes"
	"net"
	"sync"
)

type Message struct {
	Message string `json:"message"`
}

// {"id": 42, "method": "echo", "params": {"message": "Hello"}}
type Request struct {
	Id     int     `json:"id"`
	Method string  `json:"method"`
	Params Message `json:"params"`
}

// {"id": 42, "result": {"message": "Hello"}}
type Response struct {
	Id     int     `json:"id"`
	Result Message `json:"result"`
}

type Particle struct {
	Buff *bytes.Buffer
	Conn *net.Conn
	Wg   *sync.WaitGroup
}
