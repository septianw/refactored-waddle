package main

import (
	"bytes"
	"encoding/json"
	"net"
	"sync"
)

type Message struct {
	Message string `json:"message"`
}

// {"id": 42, "method": "echo", "params": {"message": "Hello"}}
type Request struct {
	Id     *json.RawMessage `json:"id"`
	Method string           `json:"method"`
	Params Message          `json:"params"`
}

// {"id": 42, "result": {"message": "Hello"}}
type Response struct {
	Id     *json.RawMessage `json:"id"`
	Result Message          `json:"result"`
}

type Particle struct {
	Buff *bytes.Buffer
	Conn *net.Conn
	Wg   *sync.WaitGroup
}
