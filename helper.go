package main

// https://gist.github.com/mattetti/5914158/f4d1393d83ebedc682a3c8e7bdc6b49670083b84

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (string, error) {

	file, _ := ioutil.ReadFile(path)
	// Create a Resty Client
	client := resty.New()

	resp, err := client.R().
		SetFileReader("file", "test-img.png", bytes.NewReader(file)).
		SetFormData(params).
		Post(uri)

	return resp.String(), err

}

func destObjectRewrite(source, format string) (string, error) {
	// dir-1/...
	// /mnt/data/chat/6286828821/FAWJGASLJG.go
	// [ mnt data chat 6286828821 FAWJGASLJG.go]
	sourceArr := strings.Split(source, "/")
	formatArr := strings.Split(format, "/")

	objName := []string{}

	if len(sourceArr) >= len(formatArr) {
		latestDirIdx := 0

		for _, path := range formatArr {

			if path == ".." {
				if latestDirIdx < len(sourceArr) {
					objName = append(objName, sourceArr[latestDirIdx])
					latestDirIdx++
				}
			} else if len(path) >= 4 {
				switch os := path[0:4]; os {
				case "dir-":
					dirIdx, err := strconv.Atoi(strings.Replace(path, "dir-", "", 1))
					if err != nil {
						panic(err)
					}

					if len(sourceArr) > 1 && dirIdx < len(sourceArr) {
						latestDirIdx = dirIdx

						objName = append(objName, sourceArr[dirIdx])
					}
				case "file":
					objName = append(objName, sourceArr[len(sourceArr)-1])
				}
			}
		}
	} else {
		return "", errors.New("path not fullfil the rule")
	}

	object := strings.Join(objName, "/")
	if object[len(object)-1:] == "/" {
		object = trimSuffix(object, "/")
	}
	return object, nil
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
