/*

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

*/

package main

import (
	"log"
	"net"
	"os"
)

func echoServer(c net.Conn) {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]

		// if receive invalid request, disconnect client immediately.
		log.Println("Data received by server", string(data)) // receive request here
		err, res := DigestReq(data)
		if err != nil {
			log.Println("digest error:", err)
			// c.Close()
		} else {
			// Make sure response are terminated by \n
			// return only single line unformatted
			_, err = c.Write(res) // send response here

			if err != nil {
				log.Fatal("write error", err)
			}
		}

	}

}

func main() {
	l, err := net.Listen("unix", os.Args[1])
	if err != nil {
		log.Fatal("net listen error", err)
	}

	for {
		fd, err := l.Accept()
		if err != nil {
			log.Fatal("net listen accept error", err)
		}

		go echoServer(fd)
	}
}
