package main

import (
	"../common"
	"encoding/json"
	"fmt"
	"github.com/chilts/sid"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

type ResponseData struct {
	Message  string `json:"res"`
	Filename string `json:"filename"`
}

func startServer(port string) {
	router := httprouter.New()
	router.GET(common.CHUNK_UPLOAD_ENDPOINT, handleChunkUpload)
	router.POST(common.FILE_UPLOAD_ENDPOINT, handleFileUpload)
	log.Println("Listening on port", port)
	log.Fatalln(http.ListenAndServe(":"+port, router))

}

func handleFileUpload(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	file, header, _ := req.FormFile("file")
	defer file.Close()
	appConfig := common.GetAppConfig()
	chunkAndSend(appConfig, &file, header)
	respondJson(rw, http.StatusOK, &ResponseData{Filename: header.Filename, Message: "successfully uploaded"})

}

func chunkAndSend(appconfig *common.AppConfig, file *multipart.File, header *multipart.FileHeader) {
	filesize := header.Size
	iterations := filesize / appconfig.Chunk.Size
	fx := header.Filename + sid.Id()
	remainder := filesize % appconfig.Chunk.Size
	total := iterations
	if remainder != 0 {
		total = total + 1
	}
	for i := int64(0); i < iterations; i++ {
		toRead := make([]byte, appconfig.Chunk.Size)
		(*file).ReadAt(toRead, appconfig.Chunk.Size*i)
		common.SendRequest(toRead, appconfig.Chunk.Size*i, i+1, fx, total)
	}
	if remainder != 0 {
		toRead := make([]byte, remainder)
		(*file).ReadAt(toRead, appconfig.Chunk.Size*iterations)
		common.SendRequest(toRead, appconfig.Chunk.Size*iterations, total, fx, total)
	}

}

func handleChunkUpload(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	query := req.URL.Query()
	total, filename := query.Get("total"), query.Get("filename")
	agent := req.Header.Get("User-Agent")
	offset := req.Header.Get("offset")
	var file *os.File
	file, _ = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	chunk, _ := ioutil.ReadAll(req.Body)
	off, _ := strconv.ParseInt(offset, 10, 64)
	file.WriteAt(chunk, off)
	if agent == common.AGENT_CLIENT {
		sendToAllOtherServers(common.GetAppConfig(), req.RequestURI, off, chunk)
	}
	if x := checkIfAllChunksReceived(total, p.ByName("id")); x {
		fmt.Println("All chunks received, fileId : ", filename)
		respondJson(rw, http.StatusOK, &ResponseData{Filename: filename, Message: "successfully uploaded"})
	}
}

func sendToAllOtherServers(config *common.AppConfig, url string, offset int64, body []byte) {
	for i := 0; i < len(config.Server.Addr); i++ {
		url := config.Server.Addr[i] + url
		common.DoRequest(common.GET_METHOD, url, common.AGENT_SERVER, offset, body)
	}
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
