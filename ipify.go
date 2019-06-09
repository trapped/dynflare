package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ipifyResponse struct {
	IP string `json:"ip"`
}

type IPify struct{}

var (
	ipifyHost = "https://api.ipify.org/?format=json"
)

func (ip *IPify) Name() string {
	return "IPify.org"
}

func (ip *IPify) Fetch() (string, error) {
	resp, err := http.Get(ipifyHost)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch %v response: %v", ip.Name(), err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read %v response: %v", ip.Name(), err)
	}
	ipifyResp := ipifyResponse{}
	err = json.Unmarshal(data, &ipifyResp)
	return ipifyResp.IP, err
}
