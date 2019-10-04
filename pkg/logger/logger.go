package logger

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// Read is reading file from logs.
func Read(fileName string) (*os.File, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path.Join(dir, "logs", fileName))
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Write is writing to logs.
func Write(fileName string, text string) error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	folder := path.Join(dir, "logs")
	if _, err = os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0744)
	}

	file, err := os.OpenFile(path.Join(folder, fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(fmt.Sprintf("%s\n", text)); err != nil {
		return err
	}

	return nil
}
