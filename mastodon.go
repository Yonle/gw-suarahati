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
	"time"
)

type Status struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Error string `json:"error"`
}

type Note struct {
	Status    string   `json:"status"`
	Medias    []string `json:"media_ids"`
	Sensitive bool     `json:"sensitive"`
}

type InstanceInfo struct {
	Title         string                `json:"title"`
	Configuration InstanceConfiguration `json:"configuration"`
}

type InstanceConfiguration struct {
	Statuses InstanceStatusesConfiguration `json:"statuses"`
}

type InstanceStatusesConfiguration struct {
	Max_Characters int64 `json:"max_characters"`
}

var auth string
var hc http.Client
var instance InstanceInfo

func init_mastodon() {
	auth = fmt.Sprintf("Bearer %s", config.Mastodon_Access_Key)
	if err := getInstance(); err != nil {
		panic(err)
	}

	go func() {
		for {
			time.Sleep(time.Minute)
			getInstance()
		}
	}()

	log.Println("Mastodon siap!")
}

func getInstance() error {
	resp, err := http.Get(config.Mastodon_Host_Url + "/api/v2/instance")
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(&instance)
}

func isTextTooLong(text string) bool {
	return int64(len(text)) > instance.Configuration.Statuses.Max_Characters
}

func masto_make_request(method string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, config.Mastodon_Host_Url+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", auth)

	return req, err
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

	req, err := masto_make_request("POST", "/api/v1/statuses", &j)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return hc.Do(req)
}

func masto_postMultipart(mp *multipart.Writer, body io.Reader) (*http.Response, error) {
	req, err := masto_make_request("POST", "/api/v2/media", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", mp.FormDataContentType())

	return hc.Do(req)
}

func getPostURL(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	var status Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", err
	}

	if len(status.Error) > 0 {
		return "", errors.New(status.Error)
	}

	return status.URL, nil
}

func masto_wait_media_ready(id string) error {
	for {
		req, err := masto_make_request("GET", "/api/v1/media/"+id, nil)
		if err != nil {
			return err
		}

		resp, err := hc.Do(req)
		if err != nil {
			return err
		}

		resp.Body.Close()

		if resp.StatusCode == 206 {
			time.Sleep(time.Second * 2)
			continue
		}

		if resp.StatusCode >= 400 {
			return errors.New(fmt.Sprintf("Status code: %d", resp.StatusCode))
		}

		return nil
	}
}
