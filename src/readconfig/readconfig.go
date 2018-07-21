// Package readconfig implements a functions to read the config file which is in json format
package readconfig

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

type FullJson struct {
	Sources       SourceName        `json:"Sources"`
	Client        ClientInfo        `json:"Client"`
	Server        ServerInfo        `json:"Server"`
	Miscellaneous MiscellaneousInfo `json:"Miscellaneous"`
}

type SourceName struct {
	NetDania     SourcesInfo `json:"NetDania"`
	OpenExchange SourcesInfo `json:"OpenExchange"`
}

type SourcesInfo struct {
	Username  string        `json:"Username"`
	Password  string        `json:"Password"`
	Frequency time.Duration `json:"Frequency"`
	Status    bool          `json:"Status"`
}

type ClientInfo struct {
	Token string   `json:"Token"`
	Ips   []string `json:"Ips"`
}

type ServerInfo struct {
	IP            string        `json:"IP"`
	Port          string        `json:"Port"`
	ServerTimeout time.Duration `json:"timeout"`
}

type MiscellaneousInfo struct {
	DebugRequest bool `json:"DebugRequest"`
	DebugSource  bool `json:"DebugSource"`
	ServerDebug  bool `json:"ServerDebug"`
	Precision    int  `json:"Precision"`
}

//Returns Informatoin collected from the json file
func GetConfigInfo() *FullJson {
	f, err := os.Open("../tmp/configFile")
	if err != nil {
		panic(err)
	}

	readBuf, _ := ioutil.ReadAll(f)
	defer f.Close()
	unmarshalledJson := &FullJson{}

	err = json.Unmarshal(readBuf, unmarshalledJson)

	if err != nil {
		panic(err)
	}

	return unmarshalledJson
}
