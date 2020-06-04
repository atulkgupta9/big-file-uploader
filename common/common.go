package common

import (
	"bytes"
	"github.com/sirupsen/logrus"
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
		logrus.Info("Error reading config file ", err)
	}
	err := viper.Unmarshal(&appConfig)
	if err != nil {
		logrus.Info("Unable to decode into struct ", err)
	}
	return &appConfig
}

func DoRequest(method, url, agent string, offset int64, body []byte) error {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", agent)
	req.Header.Set("offset", strconv.FormatInt(offset, 10))
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func GetMeServerAdd() string {
	rand.Seed(time.Now().UnixNano())
	return appconfig.Server.Addr[rand.Intn(len(appconfig.Server.Addr))]
}

func SendRequest(buffer []byte, offset int64, sequence int64, fx string, total int64) error {
	url := GetMeServerAdd() + CHUNK_UPLOAD_ENDPOINT + strconv.FormatInt(sequence, 10) + "?&filename=" + fx + "&total=" + strconv.FormatInt(total, 10)
	return DoRequest(GET_METHOD, url, AGENT_CLIENT, offset, buffer)
}
