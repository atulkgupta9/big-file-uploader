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
)

func startServer(port string) {
	router := httprouter.New()

	router.GET("/upload/", handleFilePut)

	log.Println("Listening on port", port)
	log.Fatalln(http.ListenAndServe(":"+port, router))
}

func handleFilePut(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	ch, err := ioutil.ReadAll(req.Body)
	if err == nil {
		fmt.Println("chunk received")
		fmt.Println(string(ch))
	}
	respondJson(rw, http.StatusOK, &SqlData{Res: req.Host})
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
