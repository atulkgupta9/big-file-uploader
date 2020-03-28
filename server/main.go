package main

import (
	"../common"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type SqlData struct {
	Res string `json:"res"`
}

func startServer(port string) {
	router := httprouter.New()
	router.GET("/upload/:id", handleFilePut)
	log.Println("Listening on port", port)
	log.Fatalln(http.ListenAndServe(":"+port, router))
}

func handleFilePut(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	query := req.URL.Query()
	total, filename := query.Get("total"), query.Get("filename")
	agent := req.Header.Get("User-Agent")
	offset := req.Header.Get("offset")
	fmt.Println("agent ", agent, filename)
	var file *os.File
	file, _ = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	ch, _ := ioutil.ReadAll(req.Body)
	cf, _ := strconv.ParseInt(offset, 10, 64)
	file.WriteAt(ch, cf)
	if agent == "client" {
		sendToAllOtherServers(common.GetAppConfig(), req.RequestURI, cf, ch)
	}
	if x := checkIfAllChunksReceived(total, p.ByName("id")); x {
		fmt.Println("done completed : ")
		return
	}
	respondJson(rw, http.StatusOK, &SqlData{Res: req.Host})
}

func sendToAllOtherServers(config *common.AppConfig, url string, offset int64, body []byte) {
	for i := 0; i < len(config.Server.Addr); i++ {
		url := config.Server.Addr[i] + url
		common.DoRequest("GET", url, "server", offset, body)
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
