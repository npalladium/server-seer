package sender

import (
	"../logger"
	"../storage"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Sender struct {
	ApiUrl string
}

func (self Sender) SendEntries(entries []storage.OutputEntry) {
	entriesJSON, _ := json.Marshal(entries)
	req, err := http.NewRequest("POST", self.ApiUrl, bytes.NewBuffer(entriesJSON))
	if err != nil {
		logger.Logger.Log(
			fmt.Sprintf("Error while sending request", err),
		)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}
