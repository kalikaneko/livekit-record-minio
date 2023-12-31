package livekitminio

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	webdavRoot       = "/remote.php/dav/files/"
	recordingsFolder = "/recordings/"
)

func UploadFileToWebDAV(filePath string) error {
	webdavUser := os.Getenv("NEXTCLOUD_USER")
	if webdavUser == "" {
		log.Fatal("Need to configure NEXTCLOUD_USER")
	}
	webdavUser = strings.TrimSpace(webdavUser)

	webdavPassword := os.Getenv("NEXTCLOUD_PASS")
	if webdavPassword == "" {
		log.Fatal("Need to configure NEXTCLOUD_PASS")
	}
	webdavPassword = strings.TrimSpace(webdavPassword)

	webdavInstance := os.Getenv("NEXTCLOUD_API")
	if webdavInstance == "" {
		log.Fatal("Need to configure NEXTCLOUD_API")
	}
	webdavInstance = fmt.Sprintf("http://%s", strings.TrimSpace(webdavInstance))

	folder := os.Getenv("S3_FOLDER")

	fileName := filepath.Base(filePath)

	objectName := folder + "/" + fileName

	reader, err := GetMinIOObject(objectName)
	if err != nil {
		log.Printf("minio error: %s", err.Error())
		return err
	}

	uploader := NewWebDavUploader(webdavUser, webdavPassword, webdavInstance)
	webdavFilePath := webdavRoot + webdavUser + recordingsFolder + fileName

	fmt.Println("uploading to:", webdavFilePath)

	err = uploader.Upload(reader, webdavFilePath)
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}

	log.Println("File uploaded successfully!")
	return nil
}

const (
	TypeUser  = 0
	TypeGroup = 1
	TypeTalk  = 10
)

func getShareRequestURL(user, pass, domain string) string {
	apiSharePath := "ocs/v2.php/apps/files_sharing/api/v1/shares"
	shareRequestURL := fmt.Sprintf(
		"http://%s:%s@%s/%s",
		user, pass, domain, apiSharePath)
	return shareRequestURL
}

type Share struct {
	Filename  string
	Type      uint8
	ShareWith string
}

func (s *Share) DoShare() error {
	form := map[string]string{
		"path":      s.Filename,
		"shareType": strconv.Itoa(int(s.Type)),
		"shareWith": s.ShareWith,
	}

	ct, body, err := createForm(form)
	if err != nil {
		return err
	}

	user := os.Getenv("NEXTCLOUD_USER")
	if user == "" {
		log.Fatal("Need to configure NEXTCLOUD_USER")
	}
	pass := os.Getenv("NEXTCLOUD_PASS")
	if pass == "" {
		log.Fatal("Need to configure NEXTCLOUD_PASS")
	}
	instance := os.Getenv("NEXTCLOUD_API")
	if instance == "" {
		log.Fatal("Need to configure NEXTCLOUD_API")
	}

	req, err := http.NewRequest(http.MethodPost, getShareRequestURL(user, pass, instance), body)
	if err != nil {
		return err
	}
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Content-Type", ct)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	log.Println("share api result:", res)
	return nil
}

func createForm(form map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range form {
		mp.WriteField(key, val)
	}
	return mp.FormDataContentType(), body, nil
}
