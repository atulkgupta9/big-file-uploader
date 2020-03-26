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

func startServer(port string) {
	router := httprouter.New()

	router.GET("/upload/:id", handleFilePut)

	log.Println("Listening on port", port)
	log.Fatalln(http.ListenAndServe(":"+port, router))
}

var file *os.File

func handleFilePut(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	query := req.URL.Query()
	total, filename := query.Get("total"), query.Get("filename")
	agent := req.Header.Get("User-Agent")
	fmt.Println("agent ", agent, filename)
	file, _ = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	ch, err := ioutil.ReadAll(req.Body)
	off, _ := strconv.Atoi(p.ByName("id"))
	cf := (int64)(off-1) * 1024
	n, err := file.WriteAt(ch, cf)
	fmt.Println("n, err", n, err)
	if x := checkIfAllChunksReceived(total, p.ByName("id")); x {
		fmt.Println("done completed : ")
		return
	}
	if agent == "client" {
		sendToAllOtherServers(common.GetAppConfig(), req.RequestURI, ch)
	}
	respondJson(rw, http.StatusOK, &SqlData{Res: req.Host})
}

func sendToAllOtherServers(config *common.AppConfig, url string, body []byte) {
	for i := 0; i < len(config.Server.Addr); i++ {
		url := config.Server.Addr[i] + url
		fmt.Println("url in server ", url)
		common.DoRequest("GET", url, "server", body)
	}
}

func checkIfAllChunksReceived(total string, id string) bool {
	return id == total
}

type SqlData struct {
	Res string `json:"res"`
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
	appconfig := common.GetAppConfig()
	fmt.Println(appconfig)
	startServerAny()
}

func startServerAny() {
	if len(os.Args) < 2 {
		startServer("3001")
		return
	}
	startServer(os.Args[1])
}
