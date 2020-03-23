package common

import (
	"fmt"
	"github.com/spf13/viper"
)

func GetAppConfig() *AppConfig {
	var appConfig AppConfig
	//viper.SetConfigName("config-local")
	viper.SetConfigName("config")

	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
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
