package main

import (
	"bytes"
	"encoding/json"
	"log"
)

// Ab Append buff
// Ab(base,
func Ab(base, in []byte) []byte {
	out := make([]byte, 0, len(base)+len(in)) // len 0, cap sum of base and in
	out = base
	// for _, v := range base {
	// 	out = append(out, v)
	// }
	for _, v := range in {
		out = append(out, v)
	}

	return out
}

// Req2Res are converting from request to response
func Req2Res(req Request) Response {
	var res Response

	res.Id = req.Id
	res.Result = req.Params

	return res
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
	log.Println(req)

	res = Req2Res(req)

	out, err = json.Marshal(res)
	if err == nil {
		out = Ab(out, []byte("\n"))
	}

	return err, out
}

//
func DigestReq(buff []byte) []byte {
	s := []byte("\n") // buffer out and separator

	idx := bytes.Index(buff, s)
	if idx != -1 {
		bs := bytes.Split(buff, s)

		for _, v := range bs {
			err, v := ValidateOut(v)
			if err != nil {
				rwcmut.Lock()
				crumb = Ab(crumb, v)
				rwcmut.Unlock()
				go CrumbProc()
			}
			if len(v) != 0 {
				buffOut <- v
			}
		}
		return []byte("")
	}

	return buff
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
