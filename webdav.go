package livekitminio

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/studio-b12/gowebdav"
)

type WebDavUploader struct {
	username string
	password string
	endpoint string
}

func NewWebDavUploader(username, password, endpoint string) *WebDavUploader {
	return &WebDavUploader{
		username: username,
		password: password,
		endpoint: endpoint,
	}
}

func (w *WebDavUploader) Upload(r io.Reader, remotePath string) error {
	url := w.endpoint + remotePath
	req, err := http.NewRequest("PUT", url, r)
	if err != nil {
		return err
	}

	req.SetBasicAuth(w.username, w.password)

	// I tried a streaming approach but it seems to misbehave,
	// returning Access Denied before the stream is over.

	data, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println("error reading file", err)
		return err
	}
	client := gowebdav.NewClient(w.endpoint, w.username, w.password)

	err = client.Write(remotePath, data, 0644)
	if err != nil {
		fmt.Println("error uploading file", err)
		return err
	}
	return nil
}
