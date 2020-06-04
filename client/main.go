package main

import (
	"big-file-upload/common"
	"github.com/chilts/sid"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	appconfig := common.GetAppConfig()
	err := chunkWholeFile(appconfig)
	if err != nil {
		logrus.Error("Could not chunk file : ", appconfig.Chunk.Filename, err)
		panic(err)
	}
}

func chunkWholeFile(appconfig *common.AppConfig) error {
	file, err := os.Open(appconfig.Chunk.Filename)
	if err != nil {
		return err
	}
	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		return err
	}
	iterations := fileinfo.Size() / appconfig.Chunk.Size
	fx := fileinfo.Name() + sid.Id()
	remainder := fileinfo.Size() % appconfig.Chunk.Size
	total := iterations
	if remainder != 0 {
		total = total + 1
	}
	for i := int64(0); i < iterations; i++ {
		buffer := make([]byte, appconfig.Chunk.Size)
		_, err = file.ReadAt(buffer, i*appconfig.Chunk.Size)
		if err != nil {
			return err
		}
		err = common.SendRequest(buffer, i*appconfig.Chunk.Size, i+1, fx, total)
		if err != nil {
			return err
		}
	}
	if remainder != 0 {
		buffer := make([]byte, remainder)
		_, err = file.ReadAt(buffer, iterations*appconfig.Chunk.Size)
		if err != nil {
			return err
		}
		err = common.SendRequest(buffer, iterations*appconfig.Chunk.Size, total, fx, total)
		if err != nil {
			return err
		}
	}
	return nil
}
