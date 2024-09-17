package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

type Status struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Note struct {
	Status    string   `json:"status"`
	Medias    []string `json:"media_ids"`
	Sensitive bool     `json:"sensitive"`
}

var auth string
var hc http.Client

func init_mastodon() {
	auth = fmt.Sprintf("Bearer %s", config.Mastodon_Access_Key)
	log.Println("Mastodon siap!")
}

func keluarkan(text string, mediaID *string, spoiler *bool) (*http.Response, error) {
	n := Note{
		Status: text,
	}

	if mediaID != nil {
		n.Medias = append(n.Medias, *mediaID)
	}

	if spoiler != nil {
		n.Sensitive = *spoiler
	}

	j := bytes.Buffer{}
	json.NewEncoder(&j).Encode(n)

	req, err := http.NewRequest("POST", config.Mastodon_Host_Url+"/api/v1/statuses", &j)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	return hc.Do(req)
}

func masto_postMultipart(mp *multipart.Writer, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", config.Mastodon_Host_Url+"/api/v2/media", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", mp.FormDataContentType())

	return hc.Do(req)
}

func getPostURL(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	var status Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", err
	}

	if len(status.URL) < 1 {
		return "", errors.New(
			fmt.Sprintf("Status code: %d", resp.StatusCode),
		)
	}

	return status.URL, nil
}
