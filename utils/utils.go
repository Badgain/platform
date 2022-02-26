package utils

import (
	"fmt"
	"os"
)

func OpenOrCreateFile(filename string) (*os.File, error) {
	var file *os.File
	var err error
	if file, err = os.OpenFile(filename, 0666, os.FileMode(os.O_RDWR)); err != nil && !os.IsExist(err) {
		fmt.Println(os.IsExist(err))
		if file, err = os.Create(filename); err != nil {
			return nil, err
		}
		return file, nil

	}
	return file, nil
}

