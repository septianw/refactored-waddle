/*
	# Problem 1
    [x] Each message is a JSON object contained in a single line.
    [x] Every message is terminated with the new line character \n (ASCII: 10).
    [x] Every message from a client is called a request.
    [x] Every message from the server is a response to a request previously sent by a client.
    [x] Every request must have these attributes:
        - id is an arbitrary string or number.
        - method is a string that indicates which type of request it is.
        - params is a JSON object that contains the parameters for the request. All questions will refer to attributes in this object as "parameters" for the method.
    [x] A response from the server must have these attributes:
        - id: identify which request it is for.
        - result: A JSON object that contains the responses for the corresponding request.
    [x] A client can send many requests concurrently and the server does not have to reply to them in order. This is because we can pair a response with a request using the id attribute.
    [x] A server will NEVER be expected to respond to an invalid message. It is allowed to disconnect when a client sends something invalid.

For this problem, implement only the request type: echo. For this message type, the request will also contains the param message which is an arbitrary string. The server's response must contain:

    message: The same content as the message attribute in the request.

    For example, given the request: {"id": 42, "method": "echo", "params": {"message": "Hello"}}. The correct response is: {"id": 42, "result": {"message": "Hello"}}.

	# Problem 2
	[x] stream
	[x] incomplete
	[x] duplicate
*/

package main

import (
	"bytes"
	"encoding/json"
	"io"

	// "bytes"
	"log"
	"net"
	"os"

	// "strings"
	"sync"
	"time"
)

var buffOut = make(chan []byte, 100)
var buffIn = make(chan []byte, 100)
var resOut = make(chan net.Conn, 100)
var resWaitCount int = 0
var rwcmut, buffmut sync.Mutex
var crumb []byte
var timer *time.Timer
var ta time.Time
var buffer *bytes.Buffer

func errLog(err error) {
	if err != nil {
		log.Println("error:", err)
	}
}

func readLine(lb *bytes.Buffer, c net.Conn) {
	// lb := bytes.NewBuffer(data.Bytes())

	for {
		log.Println("echoServer: start readline")

		d, err := lb.ReadBytes(10)
		log.Println("echoServer: len(d)", len(d))
		log.Println("echoServer: err", err)

		if err != nil {
			if (err == io.EOF) && (len(d) == 0) {
				lb.Reset()
				break
			} else {
				errLog(err)
			}
		}
		log.Println("echoServer: process line:", string(d))

		if !json.Valid(d) {
			log.Println("echoServer: invalid json, buffering.", string(d))

			// count request waiting
			rwcmut.Lock()
			resWaitCount++
			rwcmut.Unlock()
			log.Println("echoServer: resWaitCount", resWaitCount)

			buffmut.Lock()
			_, err := buffer.Write(d)
			buffmut.Unlock()

			errLog(err)

			if bytes.Index(d, []byte("\n")) != -1 {
				buffmut.Lock()
				nb := bytes.NewBuffer(buffer.Bytes())
				buffer.Reset()
				buffmut.Unlock()

				err, r := Sanitize(nb.Bytes())
				if err == nil {
					_, res := ValidateOut(r)
					log.Println("===>", string(res))
					log.Println("===>", res)
					_, err = c.Write(res) // send response here
					log.Println("request duration:", time.Now().Sub(ta))

					if err != nil {
						log.Fatal("write error", err)
					}
					timer.Stop()
				}
				errLog(err)

				// go readLine(nb, c)
			}
			continue

		} else {
			_, res := ValidateOut(d)
			log.Println("===>", string(res))
			log.Println("===>", res)
			_, err = c.Write(res) // send response here
			log.Println("request duration:", time.Now().Sub(ta))

			if err != nil {
				log.Fatal("write error", err)
			}
			timer.Stop()
			// wg.Done()
		}

		log.Println("echoServer: end readline")
	}
}

func echoServer(c net.Conn) {
	// var err error
	log.Println("echoServer goroutine start.")
	for {
		buf := make([]byte, 2048)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		lb := bytes.NewBuffer(buf[0:nr])

		ta = time.Now()

		readLine(lb, c)
	}

	log.Println("echoServer goroutine end.")
}

func timerf() {
	for {
		select {
		case <-timer.C:
			log.Println("Timeout.")
			log.Println("all duration:", time.Now().Sub(ta))
			// os.Exit(2)
		}
	}
}

func main() {
	l, err := net.Listen("unix", os.Args[1])
	// var wg sync.WaitGroup
	buffer = bytes.NewBuffer([]byte(""))
	buffer.Grow(5242880) // 5MB 1024×1024×5

	if err != nil {
		log.Fatal("net listen error", err)
	}

	for {
		log.Println("accept in.")
		fd, err := l.Accept()

		timer = time.NewTimer(5 * time.Second)
		go timerf()

		if err != nil {
			log.Fatal("net listen accept error", err)
		}

		go echoServer(fd)
	}

}
