package main

import (
	"../common"
	"fmt"
	"github.com/chilts/sid"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type chunk struct {
	bufsize int
	offset  int64
	seq     int
}

func main() {
	appconfig := common.GetAppConfig()
	filename := "/home/use/PERSONAL_CODES/242C.cpp"
	chunkWholeFile(filename, appconfig)

}

func getMeServerAdd(appconfig *common.AppConfig) string {
	rand.Seed(time.Now().UnixNano())
	return appconfig.Server.Addr[rand.Intn(1)]
}

func chunkWholeFile(filename string, appconfig *common.AppConfig) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	filesize := int(fileinfo.Size())

	// Number of go routines we need to spawn.
	concurrency := filesize / appconfig.Chunk.Size
	// buffer sizes that each of the go routine below should use. ReadAt
	// returns an error if the buffer size is larger than the bytes returned
	// from the file.
	chunksizes := make([]chunk, concurrency)

	fx := fileinfo.Name() + sid.Id()
	// All buffer sizes are the same in the normal case. Offsets depend on the
	// index. Second go routine should start at 100, for example, given our
	// buffer size of 100.
	for i := 0; i < concurrency; i++ {
		chunksizes[i].bufsize = appconfig.Chunk.Size
		chunksizes[i].offset = int64(appconfig.Chunk.Size * i)
		chunksizes[i].seq = i + 1
	}

	// check for any left over bytes. Add the residual number of bytes as the
	// the last chunk size.
	if remainder := filesize % appconfig.Chunk.Size; remainder != 0 {
		c := chunk{bufsize: remainder, offset: int64(concurrency * appconfig.Chunk.Size)}
		concurrency++
		c.seq = concurrency
		chunksizes = append(chunksizes, c)
	}
	wg := sync.WaitGroup{}
	wg.Add(concurrency)

	wg2 := sync.WaitGroup{}
	wg2.Add(concurrency)
	t1 := time.Now()
	fmt.Println("total chunks :", concurrency)
	_ = filename + sid.Id()
	for i := 0; i < concurrency; i++ {
		fmt.Printf("%d thread has been started \n", i)
		wg2.Add(1)
		go func(chunksizes []chunk, i int) {
			t1 := time.Now()
			defer wg.Done()
			chunk := chunksizes[i]
			buffer := make([]byte, chunk.bufsize)
			_, err := file.ReadAt(buffer, chunk.offset)

			if err != nil {
				fmt.Println(err)
				return
			}
			// initialize global pseudo random generator
			url := getMeServerAdd(appconfig) + "/upload/" + strconv.Itoa(i+1) + "?&filename=" + fx + "&total=" + strconv.Itoa(concurrency)
			fmt.Println("url ", url)
			common.DoRequest("GET", url, "client", buffer)
			fmt.Println("thread has been completed", i, time.Now().Sub(t1))
		}(chunksizes, i)
	}
	wg.Wait()
	wg2.Done()
	fmt.Println("total time taken in reading writing file size ", float64(fileinfo.Size())/(1024*1024), time.Now().Sub(t1))

}
