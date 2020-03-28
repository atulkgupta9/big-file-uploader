package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GetAppConfig() *AppConfig {
	var appConfig AppConfig
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("..")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&appConfig)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
	return &appConfig
}

type AppConfig struct {
	Chunk  Chunk
	Server Server
}
type Chunk struct {
	Size int
}
type Server struct {
	Addr []string
}
type ServerResp struct {
	code   int    `json:"-"`
	ID     string `json:"id,omitempty"`
	Offset int64  `json:"offset,omitempty"`
	Bytes  int64  `json:"bytes,omitempty"`
	Name   string `json:"name,omitempty"`
}

func DoRequest(method, url, agent string, offset int64, body []byte) (ServerResp, error) {
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("User-Agent", agent)
	req.Header.Set("offset", strconv.FormatInt(offset, 10))
	if resp, err := client.Do(req); err != nil {
		return ServerResp{}, err
	} else {
		defer resp.Body.Close()

		if rbody, err := ioutil.ReadAll(resp.Body); err != nil {
			return ServerResp{}, err
		} else {
			sresp := ServerResp{}

			if err := json.Unmarshal(rbody, &sresp); err != nil {
				return ServerResp{}, err
			}
			return sresp, nil
		}
	}
	return ServerResp{}, nil
}
