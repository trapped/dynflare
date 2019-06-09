package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type myIPcomResponse struct {
	IP string `json:"ip"`
}

type MyIPCom struct{}

var (
	myIPcomHost = "https://api.myip.com/"
)

func (mipc *MyIPCom) Name() string {
	return "MyIP.com"
}

func (ip *MyIPCom) Fetch() (string, error) {
	resp, err := http.Get(myIPcomHost)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch %v response: %v", ip.Name(), err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read %v response: %v", ip.Name(), err)
	}
	myIPcomResp := myIPcomResponse{}
	err = json.Unmarshal(data, &myIPcomResp)
	return myIPcomResp.IP, err
}
