package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Status struct {
	ID string `json:"id"`
}

var auth string
var hc http.Client

func init_mastodon() {
	auth = fmt.Sprintf("Bearer %s", config.Mastodon_Access_Key)
	log.Println("Mastodon siap!")
}

func keluarkan(text string) (*http.Response, error) {
	p := url.Values{}
	p.Set("status", text)

	req, err := http.NewRequest("POST", config.Mastodon_Host_Url+"/api/v1/statuses", strings.NewReader(p.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return hc.Do(req)
}

func getPostURL(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	var status Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", err
	}

	if len(status.ID) < 1 {
		return "", errors.New(
			fmt.Sprintf("Status code: %d", resp.StatusCode),
		)
	}

	return config.Mastodon_User_Url + "/" + status.ID, nil
}
