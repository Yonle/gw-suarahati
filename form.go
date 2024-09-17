package main

import (
	"bytes"
	"io"
	"mime/multipart"
)

func createForm(filename string, f io.Reader) (*multipart.Writer, io.Reader, error) {
	body := bytes.Buffer{}
	mp := multipart.NewWriter(&body)
	defer mp.Close()

	fileWriter, err := mp.CreateFormFile("file", filename)
	if err != nil {
		panic(err)
	}
	io.Copy(fileWriter, f)

	return mp, &body, nil
}
