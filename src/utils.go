package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
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

func Exec(cmd string) (*bytes.Buffer, error) {
	command := exec.Command("sh")
	in := bytes.NewBuffer(nil)
	out := bytes.NewBuffer(nil)
	errbytes := bytes.NewBuffer(nil)
	command.Stdin = in
	command.Stdout = out
	command.Stderr = errbytes
	in.WriteString(cmd)
	// in.WriteString("exit\n")
	err := command.Run()
	if err != nil {
		err = fmt.Errorf("%v : %s ", err, errbytes.String())
		return out, err
	}
	return out, nil
}
