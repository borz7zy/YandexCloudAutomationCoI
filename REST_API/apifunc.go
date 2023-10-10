package REST_API

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"yandex-cloud-data/LOCAL"
)

func GetDataFromAPI(folder string, newToken string) {
	sliceURLS := []string{
		"https://compute.api.cloud.yandex.net/compute/v1/disks/?folderId=",
		"https://compute.api.cloud.yandex.net/compute/v1/instances/?folderId=",
		"https://iam.api.cloud.yandex.net/iam/v1/serviceAccounts?folderId=",
	}
	var srvcUsrTkn string
	if newToken == "" {
		token, _ := ioutil.ReadFile("settings.json")
		srvcUsrTkn = "Bearer " + fmt.Sprintf("%s", gjson.Get(string(token), "token"))
	} else {
		srvcUsrTkn = "Bearer " + newToken
	}

	for i := 0; i < len(sliceURLS); i++ {
		req, _ := http.NewRequest("GET", sliceURLS[i]+folder, nil)
		u, _ := url.ParseRequestURI(sliceURLS[i])
		req.Host = fmt.Sprintf("%s", u.Host)
		req.Header.Add("Authorization", srvcUsrTkn)

		client := &http.Client{}
		resp, _ := client.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)
		if fmt.Sprintf("%s", gjson.Get(string(body), "code")) == "16" {
			cmd := exec.Command("yc", "iam create-token")
			out, err := cmd.Output()
			if err != nil {
				log.Println(err)
			}
			GetDataFromAPI(folder, string(out))
		}

		LOCAL.ControlData(body)
	}

}
