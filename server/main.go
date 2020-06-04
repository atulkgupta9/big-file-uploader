package main

import (
	"encoding/json"
	"github.com/atulkgupta9/big-file-uploader/common"
	"github.com/chilts/sid"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"html/template"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
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
	router.ServeFiles("/ui/*filepath", http.Dir("/home/use/GolandProjects/big-file-upload/static"))
	logrus.Info("Listening on port ", port)
	logrus.Fatalln(http.ListenAndServe(":"+port, router))

}

func handleIndex(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	t, err := template.ParseFiles("/home/use//GolandProjects/big-file-upload/static/index.html")
	if err != nil {
		logrus.Error("error serving file", err.Error())
		panic(err)
	}
	t.Execute(writer, nil)
}
func handleFileUpload(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	t1 := time.Now()
	file, header, err := req.FormFile("file")
	if err != nil {
		logrus.Error("could not read multipart file ", err)
		respondJson(rw, http.StatusInternalServerError, &ResponseData{Filename: "", Message: "could not upload file"})
		return
	}
	defer file.Close()
	err = chunkAndSend(config, &file, header)
	if err != nil {
		logrus.Error("Could not chunk the file ", err)
		respondJson(rw, http.StatusInternalServerError, &ResponseData{Filename: "", Message: "could not upload file"})
		return
	}
	logrus.Info("total time taken in serving request file size ", header.Size/(1024*1024), time.Now().Sub(t1))
	respondJson(rw, http.StatusOK, &ResponseData{Filename: header.Filename, Message: "successfully uploaded"})

}

/*
	sending all chunks concurrently using a buffered channel to limit maximum no of FileDescriptors,
	not using buffered channel and spawning thousands of goroutines caused "tcp too many open connections"

*/
func chunkAndSend(appconfig *common.AppConfig, file *multipart.File, header *multipart.FileHeader) error {
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
	wg.Wait()
	return nil
}

func readAndSend(wg *sync.WaitGroup, maxChan chan bool, toReadBytes, offset, i, total int64, file *multipart.File, fx string) error {
	//freeing up buffer as execution of this method completes
	defer wg.Done()
	defer func(maxChan chan bool) { <-maxChan }(maxChan)
	toRead := make([]byte, toReadBytes)
	_, err := (*file).ReadAt(toRead, offset)
	if err != nil {
		return err
	}
	return common.SendRequest(toRead, offset, i+1, fx, total)
}

//this function is called by all the chunks which hit the endpoint handleChunkUpload
func handleChunkUpload(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	query := req.URL.Query()
	total, filename := query.Get("total"), query.Get("filename")
	agent := req.Header.Get("User-Agent")
	offset := req.Header.Get("offset")
	var file *os.File
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logrus.Error("Could not open a file", err)
		respondJson(rw, http.StatusInternalServerError, &ResponseData{Filename: filename, Message: "could not upload file"})
	}
	defer file.Close()
	chunk, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logrus.Error("Could not read request body")
		respondJson(rw, http.StatusInternalServerError, &ResponseData{Filename: filename, Message: "could not upload file"})
	}
	defer req.Body.Close()
	off, err := strconv.ParseInt(offset, 10, 64)
	if err != nil {
		logrus.Error("Could not parse offset as integer in header")
		respondJson(rw, http.StatusInternalServerError, &ResponseData{Filename: filename, Message: "could not upload file"})
	}
	_, err = file.WriteAt(chunk, off)
	if err != nil {
		logrus.Error("Could not write chunk at file")
		respondJson(rw, http.StatusInternalServerError, &ResponseData{Filename: filename, Message: "could not upload file"})
	}
	if agent == common.AGENT_CLIENT {
		err := sendToAllOtherServers(config, req.RequestURI, off, chunk)
		if err != nil {
			logrus.Error("Could not send chunk to all servers")
			respondJson(rw, http.StatusInternalServerError, &ResponseData{Filename: "", Message: "could not upload file"})
		}
	}
	if x := checkIfAllChunksReceived(total, p.ByName("id")); x {
		logrus.Info("All chunks received, fileId : ", filename)
		respondJson(rw, http.StatusOK, &ResponseData{Filename: filename, Message: "successfully uploaded"})
	}
}

func sendToAllOtherServers(config *common.AppConfig, url string, offset int64, body []byte) error {
	wg := sync.WaitGroup{}
	wg.Add(len(config.Server.Addr))
	defer wg.Wait()
	for i := 0; i < len(config.Server.Addr); i++ {
		url := config.Server.Addr[i] + url
		//sending across each server asynchronously
		go wrapDoRequest(&wg, url, offset, body)
	}
	return nil
}

func wrapDoRequest(wg *sync.WaitGroup, url string, offset int64, body []byte) error {
	defer wg.Done()
	return common.DoRequest(common.GET_METHOD, url, common.AGENT_SERVER, offset, body)
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
	initializeLogger()
	startServerAny()
}

func startServerAny() {
	if len(os.Args) < 2 {
		startServer("3001")
		return
	}
	startServer(os.Args[1])
}

func initializeLogger() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2020-06-04 15:04:05"
	logrus.SetFormatter(customFormatter)
	logrus.SetReportCaller(true)
	customFormatter.FullTimestamp = true
}
