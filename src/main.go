package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/exec"
	// "os/exec"
)

func main() {
	var path string
	var resultPath string
	flag.StringVar(&path, "p", "", "-p <project_dir_path>")      // -p /home/morty/LFX_work/example
	flag.StringVar(&resultPath, "o", "", "-o <result_dir_path>") // -o /home/morty/LFX_work/functions/result_data
	flag.Parse()
	if path == "" {
		log.Fatal("The path cannot be empty")
	} else if resultPath == "" {
		log.Fatal("The resultPath cannot be empty")
	}

	fileNameToContent, err := addCommentToFile(path)
	if err != nil {
		log.Fatalf("addCommentToFile: %s", err)
	}
	out, err := exec.Command("/bin/helm", "template", path, "--debug").Output()
	if err != nil {
		log.Fatalf("Command: %s", err)
	}
	err = writeToFile(resultPath, string(out))
	if err != nil {
		log.Fatalf("writeToFile: %s", err)
	}
	err = parseFile(resultPath)
	if err != nil {
		log.Fatal(err)
	}
	err = deleteCommentInFiles(path, fileNameToContent)
	if err != nil {
		log.Fatal(err)
	}
}

func writeToFile(resultPath string, context string) error {
	resultPath = resultPath + "/example.txt"
	file, err := os.OpenFile(resultPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString(context)

	writer.Flush()
	return nil
}
