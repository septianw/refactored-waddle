package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"sync"
)

// Ab Append buff
// Ab(base,
func Ab(base, in []byte) []byte {
	out := make([]byte, len(base)+len(in)) // len 0, cap sum of base and in

	log.Println("Ab outsize:", len(base)+len(in))
	cpid := copy(out, base)
	log.Println("Ab len copied from base to out:", cpid)

	i := len(base)

	// for _, v := range base {
	// 	out = append(out, v)
	// }
	for _, v := range in {
		out[i] = v
		i++
	}

	return out
}

// Req2Res are converting from request to response
func Req2Res(req Request) Response {
	return Response{
		Id:     req.Id,
		Result: req.Params,
	}
}

func EmptyResponse() []byte {
	out, _ := json.Marshal(Response{
		Id: -1,
		Result: Message{
			Message: "",
		},
	})
	out = Ab(out, []byte("\n"))
	return out
}

// DigestReq are digesting any message coming in,
// if input valid error not nil and response not empty
// [x] validate input by marshalling
// [x] convert Response to json
// [ ] partial message
func ValidateOut(in []byte) (error, []byte) {
	var res Response
	var req Request
	var err error
	var out []byte

	// validate input here, return err immediately if fail.
	err = json.Unmarshal(in, &req)
	if err != nil {
		return err, out
	}
	log.Println("ValidateOut request:", req)

	res = Req2Res(req)

	out, err = json.Marshal(res)
	if err == nil {
		out = Ab(out, []byte("\n"))
	}

	return err, out
}

//
func DigestReq(wg *sync.WaitGroup) {
	buffmut.Lock()

	// read line by line
	// validate each line
	// if invalid sanitize it bring it to buffIn

	for {
		l, err := buffer.ReadBytes(byte(10))
		log.Println("DigestReq: processing item:", string(l))
		if err == io.EOF {
			buffer.Reset()
			break
		}
		if (err != io.EOF) && (err != nil) {
			log.Println("DigestReq: buffer readByte error:", err)
		}

		if json.Valid(l) {
			err, v := ValidateOut(l)
			if err != nil {
				log.Println("DigestReq: ValidateOut error:", err)
				go Sanitize(wg, l)
			}
			if len(v) != 0 {
				buffOut <- v
				// wg.Done()
			}
		} else {
			go Sanitize(wg, l)
		}
	}

	buffmut.Unlock()
}

// Sanitize is new CrumbProc do in goroutine and return to buffer if fail
// if success straight to buffOut
func Sanitize(wg *sync.WaitGroup, in []byte) {
	var bc, bcOi, bcCi int
	// s := []byte("\n")

	for i, v := range in {
		if v == byte(123) { // {
			// if not the first one and prefiously no space. reset to 0
			if (i > 0) && (in[i-1] != byte(32)) {
				bc = 0
			}
			if bc == 0 { // if previously 0 this is first opening
				log.Printf("Sanitize bracket open found at %d containing %s", i, string(in[i]))
				bcOi = i
			}
			bc++
		}
		if v == byte(125) { // }
			bc--
		}
		if bc == 0 { // if this become 0 this is last closing
			log.Printf("Sanitize bracket close found at %d containing %s", i, string(in[i]))
			bcCi = i + 1
			break
		}
	}

	if (bcOi < bcCi) && (len(in[bcOi:bcCi]) != 0) && (json.Valid(in[bcOi:bcCi])) {
		err, v := ValidateOut(in[bcOi:bcCi])
		if err != nil {
			log.Println("Sanitize Validate err:", err)

			buffmut.Lock()
			n, err := buffer.Write(in[bcOi:bcCi])
			buffmut.Unlock()

			if err != nil {
				log.Println("sanitize throwback err:", err)
			}
			log.Println("Sanitize throwback to buff:", n)
			// wg.Done()
		} else {
			buffOut <- v
			// wg.Done()
		}
	}
}

// CrumbProc this function will scan buffer for an object,
// this will block loop of collectingBuff
func CrumbProc() {
	rwcmut.Lock()
	buff := crumb
	log.Println("CrumbProc crumb:", crumb)
	rwcmut.Unlock()
	if len(buff) == 0 {
		return
	}
	var bo, s []byte // buffer out and separator.
	// var err error
	var bracketCount, bracketOidx, bracketCidx int
	s = []byte("\n")

	log.Println("CrumbProc input:", string(buff))
	log.Println("CrumbProc input:", buff)

	for i, v := range buff {
		if v == byte(123) { // {
			if (i > 0) && (buff[i-1] != byte(32)) {
				bracketCount = 0
			}
			if bracketCount == 0 { // if previously 0 this is first opening
				log.Printf("DigestReq bracket open found at %d containing %s", i, string(buff[i]))
				bracketOidx = i
			}
			bracketCount++
		}
		if v == byte(125) { // }
			bracketCount--
		}
		if bracketCount == 0 { // if this become 0 this is last closing
			log.Printf("DigestReq bracket close found at %d containing %s", i, string(buff[i]))
			bracketCidx = i + 1
			break
		}
	}

	if bracketOidx < bracketCidx {
		log.Printf("buff[%d:%d] %s", bracketOidx, bracketCidx, string(buff[bracketOidx:bracketCidx]))
		b := buff[bracketOidx:bracketCidx]

		if len(b) != 0 {
			bs := bytes.Join(bytes.Split(b, s), []byte(""))

			err, bs := ValidateOut(bs)
			if err != nil {
				rwcmut.Lock()
				crumb = bo
				rwcmut.Unlock()
			}

			log.Println("CrumbProc result:", string(bs))
			if len(bs) != 0 {
				buffOut <- bs
			}

			bo = buff[0:bracketOidx]
			bo = Ab(bo, buff[bracketCidx:])
		}
	}

	rwcmut.Lock()
	crumb = bo
	rwcmut.Unlock()
}
