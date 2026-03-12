package superlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// ConsoleShipper — JSON line to stdout
type ConsoleShipper struct{}

func (s *ConsoleShipper) Ship(event LogEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

// FileShipper — appends JSON lines to a file
type FileShipper struct {
	FilePath string
}

func (s *FileShipper) Ship(event LogEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(s.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file")
		}
	}(file)
	_, err = file.Write(append(jsonData, '\n'))
	return err
}

// HTTPShipper — POSTs JSON to a remote endpoint
type HTTPShipper struct {
	URL string
}

func (s *HTTPShipper) Ship(event LogEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = http.Post(s.URL, "application/json", bytes.NewBuffer(jsonData))
	return err
}
