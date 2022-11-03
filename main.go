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

func collectingResult(wg *sync.WaitGroup) {
	log.Println("collectingResult goroutine start.")
	for {
		select {
		case out := <-buffOut:
			log.Println("collectingResult: buffOut received.")

			// Make sure response are terminated by \n
			// return only single line unformatted
			log.Println("===>", string(out))
			log.Println("===>", out)

			c := <-resOut
			_, err := c.Write(out) // send response
			log.Println("request duration:", time.Now().Sub(ta))
			if err != nil {
				log.Fatal("collectingResult: write error", err)
			}
			timer.Stop()
			// wg.Done()
		default:
			continue
		}

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
			nr, err := buffer.Write(d)
			buffmut.Unlock()

			errLog(err)
			// log.Println("echoServer: buff bytes written:", nr)
			// log.Println("echoServer: len buffer b", buffer.Len())
			// log.Println("echoServer: written buffer", buffer.String())
			// log.Println("echoServer: written buffer", buffer.Bytes())
			// log.Println("echoServer: len buffer a", buffer.Len())

			if bytes.Index(d, []byte("\n")) != -1 {
				buffmut.Lock()
				nb := bytes.NewBuffer(buffer.Bytes())
				buffer.Reset()
				buffmut.Unlock()

				// log.Println("echoServer: len buffer b", buffer.Len())
				// log.Println("echoServer: content buffer", buffer.String())
				// log.Println("echoServer: content buffer", buffer.Bytes())
				// log.Println("echoServer: len buffer a", buffer.Len())

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

			// buffIn <- data
			// resOut <- c

			// _, err = c.Write(EmptyResponse())
			// if err != nil {
			// 	log.Fatal("write error", err)
			// }
			// c.Close()
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
		// log.Println("echoServer: default select")
		buf := make([]byte, 2048)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		lb := bytes.NewBuffer(buf[0:nr])
		// log.Println("echoServer: buff", data.String())
		// log.Println("echoServer: buff", data.Bytes())

		ta = time.Now()
		// log.Println("<===", data.String()) // receive request here
		// log.Println("<===", data.Bytes())  // receive request here
		// if (len(data) == 1) && (data[0] == byte(10)) {
		// 	log.Println("Only endline, ignore")
		// 	continue
		// }
		// if (len(data.Bytes()) == 1) && (data.Bytes()[0] == byte(10)) {
		// 	continue
		// }
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

	// go collectingBuff(&wg)
	// go collectingResult(&wg)

	for {
		log.Println("accept in.")
		fd, err := l.Accept()

		timer = time.NewTimer(5 * time.Second)
		go timerf()

		if err != nil {
			log.Fatal("net listen accept error", err)
		}

		// wg.Add(1)
		go echoServer(fd)
		// wg.Wait()
	}

}
