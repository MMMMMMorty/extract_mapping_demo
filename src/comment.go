package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func addCommentToFile(dir_path string) (map[string]string, error) {
	templatePath := dir_path + "/templates"
	fileNameToContent := make(map[string]string)
	err := filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".yaml") {
			fileName := info.Name()
			content, err := addCommentInLine(path, fileName)
			if err != nil {
				log.Fatalf("addCommentInLine: %s: %s", fileName, err)
				return err
			}
			fileNameToContent[fileName] = content
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileNameToContent, nil
}

func addCommentByLine(filePath string, fileName string) (string, error) {
	lines, contents, err := ReadLines(filePath)
	if err != nil {
		return "", err
	}
	// lineBytes, err := ioutil.ReadFile(filePath)
	// var lines []string
	// if err != nil {
	// 	return "", err
	// }
	// contents := string(lineBytes)
	// lines = strings.Split(contents, "\n")

	var newLines []string

	for index, line := range lines {
		comment := "#This is the " + strconv.Itoa(index+1) + " line in " + fileName
		newLines = append(newLines, comment)
		newLines = append(newLines, line)
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
	defer file.Close()
	_, err = file.WriteString(strings.Join(newLines, "\n"))
	if err != nil {
		return "", err
	}
	return contents, nil
}

func addCommentInLine(filePath string, fileName string) (string, error) {
	lineBytes, err := ioutil.ReadFile(filePath)
	var lines []string
	if err != nil {
		log.Fatal(err)
	}
	contents := string(lineBytes)
	lines = strings.Split(contents, "\n")

	var newLines []string

	for index, line := range lines {
		comment := " #This is the " + strconv.Itoa(index+1) + " line in " + fileName
		newLines = append(newLines, line+comment)
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(newLines, "\n"))
	if err != nil {
		return "", err
	}
	return contents, nil
}

func deleteCommentInFiles(dir_path string, fileNameToContent map[string]string) error {
	for fileName, content := range fileNameToContent {
		filePath := dir_path + "/templates/" + fileName
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = file.WriteString(content)
		if err != nil {
			return err
		}
	}
	return nil
}
