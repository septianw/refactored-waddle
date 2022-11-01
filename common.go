package main

import (
	_ "bytes"
	"encoding/json"
	"log"
)

// Req2Res are converting from request to response
func Req2Res(req Request) Response {
	var res Response

	res.Id = req.Id
	res.Result = req.Params

	return res
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
		out = append(out, 10)
	}

	return err, out
}

// ScanMsg this function will scan buffer for an object,
// this will block loop of collectingBuff
func DigestReq(buff []byte) []byte {
	// msgs := bytes.Split(buffer, []byte("\n"))
	var bo, s []byte // buffer out and separator.
	var err error
	s = []byte("\n")

	b := bytes.NewBuffer(buff)

	bo = buff

	return bo
}
