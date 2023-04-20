package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
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

func configurePrinterWriter(out io.Writer) yqlib.PrinterWriter {
	var printerWriter yqlib.PrinterWriter
	printerWriter = yqlib.NewSinglePrinterWriter(out)

	return printerWriter
}

func configureEncoder() yqlib.Encoder {
	indent := 2
	colorsEnabled := false
	yqlibEncoder := yqlib.NewYamlEncoder(indent, colorsEnabled, yqlib.ConfiguredYamlPreferences)
	return yqlibEncoder
}

func configureDecoder(evaluateTogether bool) yqlib.Decoder {
	prefs := yqlib.ConfiguredYamlPreferences
	prefs.EvaluateTogether = evaluateTogether
	yqlibDecoder := yqlib.NewYamlDecoder(prefs)
	return yqlibDecoder
}
