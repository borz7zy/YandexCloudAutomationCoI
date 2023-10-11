package REST_API

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"yandex-cloud-data/LOCAL"
)

func GetDataFromAPI(folder string, newToken string) {
	sliceURLS := []string{
		"https://compute.api.cloud.yandex.net/compute/v1/disks/?folderId=",
		"https://compute.api.cloud.yandex.net/compute/v1/instances/?folderId=",
		"https://iam.api.cloud.yandex.net/iam/v1/serviceAccounts?folderId=",
	}

	tokenData, _ := ioutil.ReadFile("settings.json")
	defaultToken := gjson.Get(string(tokenData), "token").String()
	srvcUsrTkn := "Bearer " + defaultToken

	if newToken != "" {
		srvcUsrTkn = "Bearer " + newToken
	}

	client := &http.Client{}

	for _, apiUrl := range sliceURLS {
		req, err := http.NewRequest("GET", apiUrl+folder, nil)
		if err != nil {
			log.Printf("Error creating request: %s", err)
			continue
		}

		req.Header.Add("Authorization", srvcUsrTkn)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making request: %s", err)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %s", err)
			continue
		}

		if gjson.Get(string(body), "code").String() == "16" && newToken == "" {
			cmd := exec.Command("yc", "iam", "create-token")
			out, err := cmd.Output()
			if err != nil {
				log.Println(err)
			}
			GetDataFromAPI(folder, string(out))
			return
		}

		LOCAL.ControlData(body)
	}
}
