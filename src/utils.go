package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func ReadLines(filePath string) ([]string, string, error) {
	lineBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
		return nil, "", err
	}
	contents := string(lineBytes)
	lines := strings.Split(contents, "\n")
	return lines, contents, nil
}
