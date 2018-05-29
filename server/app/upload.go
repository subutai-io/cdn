package app

import (
	"fmt"
	"io"
	"net/http"
)

type UploadRequest struct {
	file    io.Reader
	token   string
	private string
}

func (request *UploadRequest) ParseRequest(r *http.Request) error {
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		return err
	}
	request.file = io.Reader(file) // multipart.sectionReadCloser
	request.token = r.Header.Get("token")
	if len(request.token) == 0 {
		return fmt.Errorf("token for upload wasn't provided")
	}
	if len(r.MultipartForm.Value["private"]) > 0 {
		request.private = r.MultipartForm.Value["private"][0]
	}
	return nil
}

func Upload(request UploadRequest) {

}
