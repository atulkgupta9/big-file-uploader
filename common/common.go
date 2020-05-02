package common

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const GET_METHOD = "GET"

const CHUNK_UPLOAD_ENDPOINT = "/upload/:id"
const FILE_UPLOAD_ENDPOINT = "/file"

const AGENT_CLIENT = "client"
const AGENT_SERVER = "server"

var appconfig = GetAppConfig()

type AppConfig struct {
	Chunk  Chunk
	Server Server
}
type Chunk struct {
	Size     int64
	Filename string
}
type Server struct {
	Addr []string
}

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
	fmt.Printf("%+v\n", appConfig)
	return &appConfig
}

func DoRequest(method, url, agent string, offset int64, body []byte) error {
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("User-Agent", agent)
	req.Header.Set("offset", strconv.FormatInt(offset, 10))
	res, err := client.Do(req)
	if err != nil {
		fmt.Errorf("here err", err.Error())
		panic(err)
	}
	defer res.Body.Close()
	return err
}

func GetMeServerAdd() string {
	rand.Seed(time.Now().UnixNano())
	return appconfig.Server.Addr[rand.Intn(len(appconfig.Server.Addr))]
}

func SendRequest(buffer []byte, offset int64, sequence int64, fx string, total int64) {
	url := GetMeServerAdd() + CHUNK_UPLOAD_ENDPOINT + strconv.FormatInt(sequence, 10) + "?&filename=" + fx + "&total=" + strconv.FormatInt(total, 10)
	DoRequest(GET_METHOD, url, AGENT_CLIENT, offset, buffer)
}
