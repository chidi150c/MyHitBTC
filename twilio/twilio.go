package twilio

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TwilioAPI struct {
	AccountSid string
	AuthToken  string
	UrlStr     string
	MsgData    url.Values
	StartTime  time.Time
}

func NewTwilioAPI(myphone string) *TwilioAPI {
	t := &TwilioAPI{
		// Set account keys & information
		AccountSid: "AC0f5b6938d0dd092d0823f8fbccda30f7",
		AuthToken:  "3ef748eafff2c62e78843f462231e2e5",
		MsgData:    url.Values{},
	}
	t.UrlStr = "https://api.twilio.com/2010-04-01/Accounts/" + t.AccountSid + "/Messages.json"
	t.MsgData.Set("To", myphone)
	t.MsgData.Set("From", "(909) 639-4683")
	return t
}

func (t *TwilioAPI) TwilioTextMessage(TextMessage string) {
	// Build out the data for our message
	t.MsgData.Set("Body", TextMessage)
	msgDataReader := *strings.NewReader(t.MsgData.Encode())
	// Create Client
	client := &http.Client{}
	req, err := http.NewRequest("POST", t.UrlStr, &msgDataReader)
	if err != nil {
		fmt.Printf("Twilio: Fatal: %v", err)
		return
	}
	req.SetBasicAuth(t.AccountSid, t.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	twilioResp, _ := client.Do(req)
	if twilioResp.StatusCode >= 200 && twilioResp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(twilioResp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println(data["sid"])
		}
	} else {
		fmt.Println(twilioResp.Status)
	}
}

func (t *TwilioAPI) Inform(message string) {
	if time.Since(t.StartTime) > time.Duration(time.Hour*1) {
		t.StartTime = time.Now()
		log.Println(message)
		fmt.Println()
		t.TwilioTextMessage(message + " Time: " + t.StartTime.Format("2/1/2006 15:04:05"))
	} else if strings.Contains(message, "Alive") {
		log.Println(message)
		fmt.Println()
		t.TwilioTextMessage(message + " Time: " + t.StartTime.Format("2/1/2006 15:04:05"))
	}
	return
}
