package Const

import (
	"encoding/json"
	"log"
)

var DeploymentType struct {
	Arena       int `json:"Arena"`
	Auto        int `json:"Auto"`
	Camp        int `json:"Camp"`
	Center      int `json:"Center"`
	Circle      int `json:"Circle"`
	Custom      int `json:"Custom"`
	Edge        int `json:"Edge"`
	Line        int `json:"Line"`
	LineBack    int `json:"LineBack"`
	LineForward int `json:"LineForward"`
	Pincer      int `json:"Pincer"`
	Random      int `json:"Random"`
}

func init() {
	data := `{"Arena":10,"Auto":0,"Camp":8,"Center":6,"Circle":5,"Custom":11,"Edge":7,"Line":1,"LineBack":2,"LineForward":3,"Pincer":4,"Random":9}`
	if err := json.Unmarshal([]byte(data), &DeploymentType); err != nil {
		log.Fatalf("Unmarshal DeploymentType error: %v", err)
	}
}
