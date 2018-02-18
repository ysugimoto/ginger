package entity

import (
	"os"

	"io/ioutil"
)

type StorageObject struct {
	Data []byte
	Key  string
	Info os.FileInfo
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
	return err
}
