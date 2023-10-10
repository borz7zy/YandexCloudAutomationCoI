package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"time"
	"yandex-cloud-data/LOCAL"
	"yandex-cloud-data/REST_API"
)

func main() {
	folder, _ := ioutil.ReadFile("settings.json")
	fldr := fmt.Sprintf("%s", gjson.Get(string(folder), "folderId"))
	for {
		LOCAL.LoadHW()
		REST_API.GetDataFromAPI(fldr, "")
		time.Sleep(time.Minute * 1)
	}
}
