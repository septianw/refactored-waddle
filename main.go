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
	[ ] stream
	[ ] incomplete
	[ ] duplicate
*/

package main

import (
	// "bytes"
	"log"
	"net"
	"os"

	"strings"
	"sync"
	"time"
)

var buffOut = make(chan []byte, 100)
var buffIn = make(chan []byte, 100)
var resOut = make(chan net.Conn, 100)
var resWaitCount int = 0
var rwcmut sync.Mutex
var crumb []byte
var timer *time.Timer
var ta time.Time

func collectingBuff() {
	var buffer, lastBuff []byte
	buffer = make([]byte, 0, 5242880) // 5MB 1024*1024*5
	log.Println("collectingBuff goroutine start.")
	for {
		select {
		case in := <-buffIn:
			log.Println("collectingBuff: buffIn received.")

			log.Println("buffIn string:", string(in))
			// log.Println("buffIn bytes:", in)

			if len(in) != len(lastBuff) {
				log.Println("collectingBuff: appending to buffer.")
				t1 := time.Now()
				buffer = Ab(buffer, in)
				log.Println("Append duration:", time.Now().Sub(t1))
				lastBuff = in
			} else {
				tt0 := time.Now()
				// eq := bytes.Compare(in, lastBuff)
				eq := strings.Compare(string(in), string(lastBuff))
				log.Println("compare duration:", time.Now().Sub(tt0))
				if eq != 0 {
					log.Println("collectingBuff: appending to buffer.")
					t1 := time.Now()
					buffer = Ab(buffer, in)
					log.Println("Append duration:", time.Now().Sub(t1))
					lastBuff = in
				} else {
					log.Println("collectingBuff: Duplicate message found, ignore.")
				}
			}

			// log.Println("buffIn content predigest:", string(buffer))
			// log.Println("buffIn content predigest:", buffer)

			// sanitize, crop, validate
			buffer = DigestReq(buffer)

			// log.Println("buffIn content postdigest:", string(buffer))
			// log.Println("buffIn content postdigest:", buffer)
		default:
			continue
		}
	}
}

func collectingResult() {
	log.Println("collectingResult goroutine start.")
	for {
		select {
		case out := <-buffOut:
			log.Println("collectingResult: buffOut received.")

			// log.Println("buffOut string:", string(out))
			// log.Println("buffOut bytes:", out)

			// Make sure response are terminated by \n
			// return only single line unformatted
			log.Println("===>", string(out))
			c := <-resOut
			_, err := c.Write(out) // send response
			log.Println("request duration:", time.Now().Sub(ta))
			if err != nil {
				log.Fatal("collectingResult: write error", err)
			}
			timer.Stop()
		default:
			continue
		}

	}
}

func echoServer(c net.Conn) {
	log.Println("echoServer goroutine start.")
	for {
		// log.Println("echoServer: default select")
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]

		ta = time.Now()
		log.Println("<===", string(data)) // receive request here
		if (len(data) == 1) && (data[0] == byte(10)) {
			log.Println("Only endline, ignore")
			continue
		}
		err, res := ValidateOut(data)
		if err != nil {
			log.Println("echoServer: invalid json, buffering.")

			// count request waiting
			rwcmut.Lock()
			resWaitCount++
			rwcmut.Unlock()
			log.Println("echoServer: resWaitCount", resWaitCount)

			buffIn <- data
			resOut <- c

			// _, err = c.Write(EmptyResponse())
			// if err != nil {
			// 	log.Fatal("write error", err)
			// }
			// c.Close()
		} else {
			log.Println("===>", string(res))
			_, err = c.Write(res) // send response here
			log.Println("request duration:", time.Now().Sub(ta))

			if err != nil {
				log.Fatal("write error", err)
			}
			timer.Stop()
		}
	}

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
	if err != nil {
		log.Fatal("net listen error", err)
	}

	go collectingBuff()
	go collectingResult()

	for {
		fd, err := l.Accept()
		timer = time.NewTimer(2 * time.Second)
		go timerf()

		if err != nil {
			log.Fatal("net listen accept error", err)
		}

		go echoServer(fd)
	}
}
