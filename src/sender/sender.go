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
	ApplicationKey string
	ServerHandler  string
	ApiUrl         string
}

type SendEntries struct {
	ApplicationKey string
	ServerHandler  string
	Entries        []storage.OutputEntry
}

type ResponseData struct {
	Status string `json:'status'`
}

func (self Sender) SendEntries(entries []storage.OutputEntry) bool {
	sendEntriesData := SendEntries{
		ApplicationKey: self.ApplicationKey,
		ServerHandler:  self.ServerHandler,
		Entries:        entries,
	}
	// entriesJSON, _ := json.Marshal(entries)
	sendEntriesDataJSON, _ := json.Marshal(sendEntriesData)
	req, err := http.NewRequest("POST", self.ApiUrl, bytes.NewBuffer(sendEntriesDataJSON))
	if err != nil {
		logger.Logger.Log(
			fmt.Sprintf("Error while sending request", err),
		)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	responseData := ResponseData{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		logger.Logger.Log(
			fmt.Sprintf("Response from consuming server failed to be parsed: %s", err),
		)
	}

	logger.Logger.Log(
		fmt.Sprintf("Response from consuming server received: %s", string(body)),
	)

	if responseData.Status == "OK" {
		return true
	}

	return false

}
