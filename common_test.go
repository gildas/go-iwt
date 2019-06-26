package iwt_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gildas/go-logger"
)

func LoadData(filename string) ([]byte, error) {
	if strings.HasPrefix(filename, "/") {
		return ioutil.ReadFile(filename)
	}
	return ioutil.ReadFile(filepath.Join(".", "testdata", filename))
}

func Load(filename string, object interface{}) (err error) {
	if len(filename) == 0 {
		return nil
	}
	var payload []byte

	if payload, err = LoadData(filename); err != nil {
		return err
	}
	if err = json.Unmarshal(payload, &object); err != nil {
		return err
	}
	return nil
}

func CreateLogger(filename string) *logger.Logger {
	folder := filepath.Join(".", "log")
	os.MkdirAll(folder, os.ModePerm)
	return logger.CreateWithDestination("test", "file://"+filepath.Join(folder, filename))
}