package main

import (
	// "bytes"
	"encoding/json"
	"errors"
	"time"

	// "io"
	"log"
	// "sync"
)

// Ab Append buff
// Ab(base,
func Ab(base, in []byte) []byte {
	out := make([]byte, len(base)+len(in)) // len 0, cap sum of base and in

	log.Println("Ab outsize:", len(base)+len(in))
	tb := time.Now()
	cpid := copy(out, base)
	log.Println("Ab: cost of copy(out, base)", time.Now().Sub(tb))
	log.Println("Ab len copied from base to out:", cpid)

	i := len(base)

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

// DigestReq are digesting any message coming in,
// if input valid error not nil and response not empty
// [x] validate input by marshalling
// [x] convert Response to json
// [x] partial message
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

// Sanitize is new CrumbProc do in goroutine and return to buffer if fail
// if success straight to buffOut
func Sanitize(in []byte) (error, []byte) {
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
		return nil, in[bcOi:bcCi]
	}

	return errors.New("Sanitize fail, JSON not found."), []byte("")
}
