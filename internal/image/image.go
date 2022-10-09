package image

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

var (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	wordBank = []string{"Swing", "Sick", "Unaware", "Ear", "Frame",
		"Ignorant", "Insistence", "detective", "Dorm",
		"Theory", "Fresh", "Tract", "Nest", "Mars", "Shatter",
		"Draw", "Crash", "Past", "Vegetable", "Impact", "Is",
		"Cut", "Pull", "Thought", "Tenant", "Broadcast", "Selection",
		"Second", "Wrong", "Stop", "Dinner", "Net", "terrify", "Cheap",
		"Conservative", "Throne", "symptom", "eyebrow", "Helpless",
		"Reliable", "Compose", "Partnership", "asylum", "Fool",
		"Credit", "Angel", "Established", "Hard", "Press", "Grass",
	}
)

func RandString() string {
	rand.Seed(time.Now().UnixNano())
	res := ""
	for i := 0; i < 64; i++ {
		res += string(alphabet[rand.Intn(len(alphabet))])
	}

	return res
}

// GenerateName generates a random name containing various words.
func GenerateName() string {
	rand.Seed(time.Now().UnixNano())
	res := ""
	for i := 0; i < 6; i++ {
		res += wordBank[rand.Intn(len(wordBank))]
	}

	return res
}

type Sizer interface {
	Size() int64
}

type File struct {
	Name        string
	Data        io.Reader
	Size        int64
	ContentType string
}

func Read(file multipart.File, multiHeader *multipart.FileHeader, r *http.Request) (*File, error) {
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			zap.S().Warnf("Failed to close file: %s", err)
		}
	}(file)

	fileHeader := make([]byte, 512)

	if _, err := file.Read(fileHeader); err != nil {
		zap.S().Warnf("Failed to read file header: %s", err)
		return nil, err
	}

	if _, err := file.Seek(0, 0); err != nil {
		zap.S().Warnf("Failed to seek file: %s", err)
		return nil, err
	}

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, file); err != nil {
		zap.S().Warnf("Failed to copy file: %s", err)
		return nil, err
	}

	ext := filepath.Ext(multiHeader.Filename)
	t := mime.TypeByExtension(ext)
	if t == "" {
		t = http.DetectContentType(fileHeader)
	}
	if t == "application/octet-stream" {
		t2 := r.Header.Get("Content-Type")
		if t2 != "" {
			t = t2
		}
	}

	f := File{
		Name:        GenerateName(),
		Data:        bytes.NewBuffer(buf.Bytes()),
		Size:        file.(Sizer).Size(),
		ContentType: t,
	}

	if len(f.ContentType) <= 1 {
		return nil, errors.New("unknown file type")
	}

	return &f, nil
}
