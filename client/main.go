package main

import (
	"../common"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	appconfig := common.GetAppConfig()
	file, err := os.Open("/home/use/PERSONAL_CODES/242C.cpp")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	buffer := make([]byte, appconfig.Chunk.Size)
	_, err = file.ReadAt(buffer, 0)
	if err != nil {
		fmt.Println("error while reading", err)
	}

	if resp, err := doRequest("GET", buffer); err != nil {
		log.Fatalf("Failed to upload chunk: %s\n", err)
	} else {
		log.Printf("Got response: %+v\n", resp)
	}

}

func doRequest(method string, body []byte) (common.ServerResp, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, "http://localhost:3001/upload/", bytes.NewReader(body))

	if err != nil {
		return common.ServerResp{}, err
	}
	if resp, err := client.Do(req); err != nil {
		return common.ServerResp{}, err
	} else {
		defer resp.Body.Close()

		if rbody, err := ioutil.ReadAll(resp.Body); err != nil {
			return common.ServerResp{}, err
		} else {
			sresp := common.ServerResp{}

			if err := json.Unmarshal(rbody, &sresp); err != nil {
				return common.ServerResp{}, err
			}
			return sresp, nil
		}
	}
	return common.ServerResp{}, nil
}
