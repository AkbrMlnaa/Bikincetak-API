package erpnext

import (
	"bytes"
	"io"
	"net/http"
	"os"
)

func ERPNextReq(method, endpoint string, payload []byte) ([]byte, error)  {
	url:= os.Getenv("ERPNEXT_URL") + endpoint
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+os.Getenv("ERPNEXT_API_KEY")+":"+os.Getenv("ERPNEXT_API_SECRET"))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return io.ReadAll(res.Body)
}