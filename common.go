package main

import (
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
// [ ] convert Response to json
func DigestReq(in []byte) (error, []byte) {
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
