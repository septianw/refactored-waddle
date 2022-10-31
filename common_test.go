package main

import (
	"strings"
	"testing"
)

func TestReq2Res(t *testing.T) {
	var req Request
	var msg Message

	msg.Message = "this is test"

	req.Id = 42
	req.Method = "test"
	req.Params = msg

	res := Req2Res(req)

	t.Log("req", req)
	t.Log("res", res)

	if strings.Compare(req.Params.Message, res.Result.Message) != 0 {
		t.Fail()
	}

	if req.Id != res.Id {
		t.Fail()
	}
}

func TestDigestReq(t *testing.T) {
	reqSuccess := []byte(`{"id": 42, "method": "echo", "params": {"message": "Hello"}}`)
	reqFail1 := []byte(`{"id": 42, "method": "echo", "params": 34}`)
	reqFail2 := []byte(`help`)

	err, result := DigestReq(reqSuccess)
	if err != nil {
		t.Fail()
	}
	// if result.Id != 42 {
	// 	t.Fail()
	// }
	// if strings.Compare(result.Result.Message, "Hello") != 0 {
	// 	t.Fail()
	// }
	t.Log(err)
	t.Log(string(result))
	t.Log(result)
	t.Log(string(reqSuccess))

	err, result = DigestReq(reqFail1)
	if err == nil {
		t.Fail()
	}
	t.Log(err)
	t.Log(result)
	t.Log(string(reqFail1))

	err, result = DigestReq(reqFail2)
	if err == nil {
		t.Fail()
	}
	t.Log(err)
	t.Log(result)
	t.Log(string(reqFail2))
}
