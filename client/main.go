package main

import (
	"../common"
	"github.com/chilts/sid"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func main() {
	appconfig := common.GetAppConfig()
	filename := "/home/use/Flipkart-Labels-03-Jan-2020-02-42.pdf"
	chunkWholeFile(filename, appconfig)
}

func getMeServerAdd(appconfig *common.AppConfig) string {
	rand.Seed(time.Now().UnixNano())
	return appconfig.Server.Addr[rand.Intn(2)]
}

func chunkWholeFile(filename string, appconfig *common.AppConfig) {
	file, _ := os.Open(filename)
	defer file.Close()
	fileinfo, _ := file.Stat()
	filesize := int(fileinfo.Size())
	iterations := filesize / appconfig.Chunk.Size
	fx := fileinfo.Name() + sid.Id()
	remainder := filesize % appconfig.Chunk.Size
	total := iterations
	if remainder != 0 {
		total = total + 1
	}
	for i := 0; i < iterations; i++ {
		sendRequest(appconfig, file, appconfig.Chunk.Size, (int64)(i*appconfig.Chunk.Size), i+1, fx, total)
	}
	if remainder != 0 {
		sendRequest(appconfig, file, remainder, (int64)(iterations*appconfig.Chunk.Size), total, fx, total)
	}
}

func sendRequest(appconfig *common.AppConfig, file *os.File, size int, offset int64, sequence int, fx string, total int) {
	buffer := make([]byte, size)
	file.ReadAt(buffer, offset)
	// initialize global pseudo random generator
	url := getMeServerAdd(appconfig) + "/upload/" + strconv.Itoa(sequence) + "?&filename=" + fx + "&total=" + strconv.Itoa(total)
	common.DoRequest("GET", url, "client", offset, buffer)
}
