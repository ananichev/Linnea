package models

import "encoding/json"

type File struct {
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	ContentType string `json:"type"`
}

func (f File) ToString() (string, error) {
	b, err := json.Marshal(f)
	return string(b), err
}