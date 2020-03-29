package main

import (
	"../common"
	"github.com/chilts/sid"
	"os"
)

func main() {
	appconfig := common.GetAppConfig()
	chunkWholeFile(appconfig)
}

func chunkWholeFile(appconfig *common.AppConfig) {
	file, _ := os.Open(appconfig.Chunk.Filename)
	defer file.Close()
	fileinfo, _ := file.Stat()
	iterations := fileinfo.Size() / appconfig.Chunk.Size
	fx := fileinfo.Name() + sid.Id()
	remainder := fileinfo.Size() % appconfig.Chunk.Size
	total := iterations
	if remainder != 0 {
		total = total + 1
	}
	for i := int64(0); i < iterations; i++ {
		buffer := make([]byte, appconfig.Chunk.Size)
		file.ReadAt(buffer, i*appconfig.Chunk.Size)
		common.SendRequest(buffer, i*appconfig.Chunk.Size, i+1, fx, total)
	}
	if remainder != 0 {
		buffer := make([]byte, remainder)
		file.ReadAt(buffer, iterations*appconfig.Chunk.Size)
		common.SendRequest(buffer, iterations*appconfig.Chunk.Size, total, fx, total)
	}
}
