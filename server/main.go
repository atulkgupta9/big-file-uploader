package main

import (
	"../common"
	"encoding/json"
	"fmt"
	"github.com/chilts/sid"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type ResponseData struct {
	Message  string `json:"res"`
	Filename string `json:"filename"`
}

var config = common.GetAppConfig()

func startServer(port string) {
	router := httprouter.New()
	router.GET(common.CHUNK_UPLOAD_ENDPOINT, handleChunkUpload)
	router.POST(common.FILE_UPLOAD_ENDPOINT, handleFileUpload)
	router.GET("/index", handleIndex)
	router.ServeFiles("/ui/*filepath", http.Dir("../static"))
	log.Println("Listening on port", port)
	log.Fatalln(http.ListenAndServe(":"+port, router))

}

func handleIndex(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	t, err := template.ParseFiles("../static/index.html")
	if err != nil {
		fmt.Errorf("error serving file", err.Error())
		panic(err)
	}
	t.Execute(writer, nil)
}
func handleFileUpload(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	t1 := time.Now()
	file, header, _ := req.FormFile("file")
	defer file.Close()
	chunkAndSend(config, &file, header)
	respondJson(rw, http.StatusOK, &ResponseData{Filename: header.Filename, Message: "successfully uploaded"})
	fmt.Println("total time taken in serving request file size ", header.Size/(1024*1024), time.Now().Sub(t1))

}

/*
	sending all chunks concurrently using a buffered channel to limit maximum no of FileDescriptors,
	not using buffered channel and spawning thousands of goroutines caused "tcp too many open connections"

*/
func chunkAndSend(appconfig *common.AppConfig, file *multipart.File, header *multipart.FileHeader) {
	filesize := header.Size
	iterations := filesize / appconfig.Chunk.Size
	fx := sid.IdBase64() + header.Filename
	remainder := filesize % appconfig.Chunk.Size
	total := iterations
	if remainder != 0 {
		total = total + 1
	}
	wg := sync.WaitGroup{}
	wg.Add(int(total))
	//maximum we want to process 100 goroutines at once to hit server to upload chunk
	maxChan := make(chan bool, 100)
	for i := int64(0); i < iterations; i++ {
		//will be able to send only if there are less then equal to 100 entries in buffered channel
		//will be blocking if already 100 are there in channel waiting to process
		maxChan <- true
		go readAndSend(&wg, maxChan, appconfig.Chunk.Size, appconfig.Chunk.Size*i, i, total, file, fx)
	}
	if remainder != 0 {
		maxChan <- true
		go readAndSend(&wg, maxChan, remainder, appconfig.Chunk.Size*iterations, total, total, file, fx)
	}
	fmt.Println("goroutines spawned", runtime.NumGoroutine())
	wg.Wait()
}

func readAndSend(wg *sync.WaitGroup, maxChan chan bool, toReadBytes, offset, i, total int64, file *multipart.File, fx string) {
	defer wg.Done()
	//freeing up buffer as execution of this method completes
	defer func(maxChan chan bool) { <-maxChan }(maxChan)
	toRead := make([]byte, toReadBytes)
	(*file).ReadAt(toRead, offset)
	common.SendRequest(toRead, offset, i+1, fx, total)
}

//this function is called by all the chunks which hit the endpoint handleChunkUpload
func handleChunkUpload(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	query := req.URL.Query()
	total, filename := query.Get("total"), query.Get("filename")
	agent := req.Header.Get("User-Agent")
	offset := req.Header.Get("offset")
	var file *os.File
	file, _ = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	chunk, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	off, _ := strconv.ParseInt(offset, 10, 64)
	file.WriteAt(chunk, off)
	if agent == common.AGENT_CLIENT {
		sendToAllOtherServers(config, req.RequestURI, off, chunk)
	}
	if x := checkIfAllChunksReceived(total, p.ByName("id")); x {
		fmt.Println("All chunks received, fileId : ", filename)
		respondJson(rw, http.StatusOK, &ResponseData{Filename: filename, Message: "successfully uploaded"})
	}
}

func sendToAllOtherServers(config *common.AppConfig, url string, offset int64, body []byte) {
	wg := sync.WaitGroup{}
	wg.Add(len(config.Server.Addr))
	for i := 0; i < len(config.Server.Addr); i++ {
		url := config.Server.Addr[i] + url
		//sending across each server asynchronously
		go wrapDoRequest(&wg, url, offset, body)
	}
	wg.Wait()
}

func wrapDoRequest(wg *sync.WaitGroup, url string, offset int64, body []byte) {
	common.DoRequest(common.GET_METHOD, url, common.AGENT_SERVER, offset, body)
	wg.Done()
}


func checkIfAllChunksReceived(total string, id string) bool {
	return id == total
}

func respondJson(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func main() {
	startServerAny()
}

func startServerAny() {
	if len(os.Args) < 2 {
		startServer("3001")
		return
	}
	startServer(os.Args[1])
}
