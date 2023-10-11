package main

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"time"
	"yandex-cloud-data/LOCAL"
	"yandex-cloud-data/REST_API"
)

func main() {
	folder, err := ioutil.ReadFile("settings.json")
	if err != nil {
		log.Fatalf("Failed to read settings.json: %v", err)
	}

	fldr := gjson.GetBytes(folder, "folderId").String()

	for {
		LOCAL.LoadHW()
		REST_API.GetDataFromAPI(fldr, "")
		time.Sleep(time.Minute * 1)
	}
}
