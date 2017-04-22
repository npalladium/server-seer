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
	ApiKey string
	ApiUrl string
}

type SendEntries struct {
	ApiKey  string
	Entries []storage.OutputEntry
}

func (self Sender) SendEntries(entries []storage.OutputEntry) {
	sendEntriesData := SendEntries{
		ApiKey:  self.ApiKey,
		Entries: entries,
	}
	// entriesJSON, _ := json.Marshal(entries)
	sendEntriesDataJSON, _ := json.Marshal(sendEntriesData)
	req, err := http.NewRequest("POST", self.ApiUrl, bytes.NewBuffer(sendEntriesDataJSON))
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

	fmt.Println("request Body:", sendEntriesData)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

}
