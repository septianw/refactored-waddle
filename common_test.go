package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReq2Res(t *testing.T) {
	var req Request
	var res Response
	var msg Message
	var d json.RawMessage

	msg.Message = "this is test"

	d = json.RawMessage("42")
	req.Id = &d
	req.Method = "test"
	req.Params = msg

	res = Req2Res(req)

	t.Log("req", req)
	t.Log("res", res)

	d = json.RawMessage("-1")
	req.Id = &d

	res = Req2Res(req)

	t.Log("req", req)
	t.Log("res", res)

	if strings.Compare(req.Params.Message, res.Result.Message) != 0 {
		t.Fail()
	}

	if req.Id != res.Id {
		t.Fail()
	}
}

func TestValidateOut(t *testing.T) {
	reqSuccess := []byte(`{"id": 42, "method": "echo", "params": {"message": "Hello"}}`)
	reqSuccess2 := []byte(`{"id": "test", "method": "echo", "params": {"message": "Hello"}}`)
	reqFail1 := []byte(`{"id": 42, "method": "echo", "params": 34}`)
	reqFail2 := []byte(`help`)

	err, result := ValidateOut(reqSuccess)
	if err != nil {
		t.Fail()
	}
	t.Log(err)
	t.Log(string(result))
	t.Log(result)
	t.Log(string(reqSuccess))

	err, result = ValidateOut(reqSuccess2)
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
	t.Log(string(reqSuccess2))

	err, result = ValidateOut(reqFail1)
	if err == nil {
		t.Fail()
	}
	t.Log(err)
	t.Log(result)
	t.Log(string(reqFail1))

	err, result = ValidateOut(reqFail2)
	if err == nil {
		t.Fail()
	}
	t.Log(err)
	t.Log(result)
	t.Log(string(reqFail2))
}

func TestDigestReq(t *testing.T) {
	// var input [][]byte
	// input = append(input, []byte(`{"id": 0, "method": "echo",{"id": 0, "method": "echo","params": {"message": "Hello"}}`))
	// input = append(input, []byte(`{"id": 0, "method": "echo",
	// "params": {"message": "Hello"
	// }}`))

	// for i, v := range input {
	// t.Log("input", string(v))
	// DigestReq()
	// t.Log("index", i, "output", string(o))
	// }

	// t.Log([]byte("\n"))
}
