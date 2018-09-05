package radosgw

import (
	"encoding/json"
	"testing"

	"logger"
)

const (
	testEndpoint  = "http://10.190.75.12:8080"
	testAccessKey = "3AZRWFYUYKRJHKJ4YI29"
	testSecretKey = "NZnrUDMOOKnPeQQCki8ANc07FLhpCpDDK9M3ZjeX"
)

var testClient *Client = nil

func init() {
	var err error
	testClient, err = NewClient(testEndpoint, testAccessKey, testSecretKey)
	if err != nil {
		panic("create radosgw client failed:" + err.Error())
	}
	logger.SetLogHandler(logger.STDERR)
	logger.SetLogLevel(logger.DEBUG)
}

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Log("error occurs:", err)
		t.Fail()
	}
}

func TestUser(t *testing.T) {
	status, data, err := testClient.CreateUser("oshyn", "Oshyn Song", "", 2)
	checkError(t, err)
	t.Log("status=", status)
	t.Logf("data=%+v", data)

	status, data, err = testClient.CreateUser("", "Oshyn Song", "xxxxxx", 2000)
	t.Log(err)
	t.Log("status=", status)
	t.Logf("data=%+v", data)

	status, data, err = testClient.GetUser("oshyn")
	checkError(t, err)
	t.Log("status=", status)
	t.Logf("data=%+v", data)

	status, data, err = testClient.UpdateUser("oshyn", "oshyn", "abc@baidu.com", 20, false)
	checkError(t, err)
	t.Log("status=", status)
	t.Logf("data=%+v", data)

	status, err = testClient.DeleteUser("oshyn", true)
	checkError(t, err)
	t.Log("status=", status)
}

func TestBucket(t *testing.T) {
	s, err := testClient.CreateBucket("bucket1", "", "")
	t.Log(err, s)
	s, err = testClient.CreateBucket("bucket2", "", "")
	t.Log(s)
	s, err = testClient.CreateBucket("bucket3", "", "")
	t.Log(s)
	testClient.DeleteBucket("bucket1", false)
	testClient.DeleteBucket("bucket2", false)
	testClient.DeleteBucket("bucket3", false)
}

func TestQuota(t *testing.T) {
	s, q, e := testClient.GetQuota("oshyn", "user")
	t.Log(s, q, e)
	s, e = testClient.SetQuota("oshyn", "user", 104857600, -1, true, "")
	t.Log(s, e)
	s, q, e = testClient.GetQuota("oshyn", "user")
	t.Log(s, q, e)
	s, e = testClient.SetQuota("oshyn", "user", -1, 9999999, true, "")
	t.Log(s, e)
	s, q, e = testClient.GetQuota("oshyn", "user")
	t.Log(s, q, e)
}

func TestUsage(t *testing.T) {
	//status, data, err := testClient.GetUser("oshyn")
	//if err != nil || status != 200 {
	//	_, data, _ = testClient.CreateUser("oshyn", "Oshyn Song", "abc@baidu.com", 10)
	//}
	//t.Log(data)

	status, usage, err := testClient.GetUsage("ssy", nil, nil, false, false)
	t.Log(status, err)
	raw, _ := json.MarshalIndent(usage, "", "  ")
	t.Log(string(raw))
}

func TestGetBucket(t *testing.T) {
	s, data, err := testClient.GetBucket("", "", true)
	t.Log(err, s)
	t.Log(len(data))

	s, policy, err := testClient.GetPolicy("ssy", "")
	t.Log(err, s)
	t.Log(string(policy))
}
