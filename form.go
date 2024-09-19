package main

import (
	"io"
	"mime/multipart"
)

func writeMultipartBody(filename string, f io.Reader, mp *multipart.Writer, w *io.PipeWriter) {
	defer w.Close()
	defer mp.Close()
	fileWriter, err := mp.CreateFormFile("file", filename)
	if err != nil {
		return
	}
	io.Copy(fileWriter, f)
}

func createForm(filename string, f io.Reader) (*multipart.Writer, *io.PipeReader, error) {
	r, w := io.Pipe()
	mp := multipart.NewWriter(w)

	go writeMultipartBody(filename, f, mp, w)

	return mp, r, nil
}
