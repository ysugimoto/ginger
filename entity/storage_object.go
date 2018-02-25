package entity

import (
	"os"
	"strings"

	"io/ioutil"
	"net/http"
)

type StorageObject struct {
	Data     []byte
	Key      string
	MimeType string
	Info     os.FileInfo
}

func NewStorageObject(key string, info os.FileInfo) *StorageObject {
	return &StorageObject{
		Data: make([]byte, 0),
		Key:  key,
		Info: info,
	}
}

func (s *StorageObject) Load(path string) (err error) {
	s.Data, err = ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	mime := http.DetectContentType(s.Data)
	if index := strings.Index(mime, ";"); index != -1 {
		mime = mime[0:index]
	}
	s.MimeType = mime
	return nil
}
