package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	SOURCE_FILE_PATTERN        = `(?m)Source: (\S*)/(?P<fileName>[\w-_]*\.yaml)$`
	API_VERSION_PATTERN        = `(?m)apiVersion: (?P<apiVersion>\S*) #This is the (?P<line>[0-9]+) line in (?P<fileName>\S+)\.yaml(.*)$`
	KIND_PATTERN               = `(?m)kind: (?P<kind>\S*) #This is the (?P<line>[0-9]+) line in (?P<fileName>\S+)\.yaml(.*)$`
	VALUE_LINE_COMMENT_PATTERN = `(?m)(?P<hierarchy>[ |-]*)(?P<key>\S+): (?P<value>\S+) #This is the (?P<line>[0-9]+) line in (?P<fileName>\S+).yaml(.*)$`
	LINE_COMMENT_PATTERN       = `(?m)(?P<hierarchy>[ |-]*)(?P<key>\S+): #This is the (?P<line>[0-9]+) line in (?P<fileName>\S+).yaml(.*)$`
	VALUE_LINE_PATTERN         = `(?m)(?P<hierarchy>[ |-]*)(?P<key>\S+): (?P<value>\S+)(.*)$`
)

type ObjectID struct {
	apiVersion string
	kind       string
}

type Node struct {
	objectID *ObjectID
	field    string
	value    string
	file     string
	line     int
}

func (node *Node) writeNote(re *regexp.Regexp, line string, fields []string, objectID *ObjectID, index int, isValue bool, isComment bool) error {
	node.field = strings.Join(fields[:(index+1)], ".")
	if isComment == true {
		lineNumber := extractParameter(re, line, "$line")
		lineNum, err := strconv.Atoi(lineNumber)
		if err != nil {
			return err
		}
		node.line = lineNum
	}
	if isValue == true {
		value := extractParameter(re, line, "$value")
		if value == "" {
			err := fmt.Errorf("Something wrong when extracting the value, the line is %s\n", line)
			return err
		}
		node.value = value
	} else {
		node.value = ""
	}
	node.objectID = objectID
	return nil
}

func (node *Node) writeNodeToFile(path string) error {
	path = path + "/result.jsonl"
	objectIDMapping := map[string]string{
		"apiVersion": node.objectID.apiVersion,
		"kind":       node.objectID.kind,
	}
	objectJson, err := json.Marshal(objectIDMapping)
	if err != nil {
		err := fmt.Errorf("json_Marshal: %w", err)
		return err
	}
	nodeMapping := map[string]interface{}{
		"objectID": string(objectJson),
		"field":    node.field,
		"value":    node.value,
		"file":     node.file,
		"line":     node.line,
	}
	nodeJson, err := json.Marshal(nodeMapping)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := file.Write(nodeJson); err != nil {
		return err
	}
	if _, err := file.Write([]byte("\n")); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

func parseFile(path string) error {
	resultPath := path + "/example.txt"
	lines, _, err := ReadLines(resultPath)
	if err != nil {
		return err
	}
	isEmpty := true
	node := new(Node)
	objectID := new(ObjectID)
	// fields := make(map[int]string)
	fields := make([]string, 10)
	realLen := 0
	for _, line := range lines {
		soueceFileRe := regexp.MustCompile(SOURCE_FILE_PATTERN)
		apiVersionRe := regexp.MustCompile(API_VERSION_PATTERN)
		kindRe := regexp.MustCompile(KIND_PATTERN)
		valueLineCommentRe := regexp.MustCompile(VALUE_LINE_COMMENT_PATTERN)
		lineCommentRe := regexp.MustCompile(LINE_COMMENT_PATTERN)
		valueLineRe := regexp.MustCompile(VALUE_LINE_PATTERN)
		// map the "# Source: xx.yaml" fileName = xx isEmpty = true
		if soueceFileRe.MatchString(line) {
			fileName := extractParameter(soueceFileRe, line, "$fileName")
			if fileName == "" {
				err := fmt.Errorf("Something wrong when extracting the fileName, the line is %s\n", line)
				return err
			}
			node = new(Node)
			objectID = new(ObjectID)
			node.file = fileName
			isEmpty = true
			continue
		}
		// map apiVersion: %w* #This is %d line in %w.yaml
		if apiVersionRe.MatchString(line) {
			apiVersion := extractParameter(apiVersionRe, line, "$apiVersion")
			if apiVersion == "" {
				err := fmt.Errorf("Something wrong when extracting the apiVersion, the line is %s\n", line)
				return err
			}
			objectID.apiVersion = apiVersion
			isEmpty = false
			continue
		}
		// map kind: %w* #This is %d line in %w.yaml OpenID = isEmpty = false
		if kindRe.MatchString(line) {
			kind := extractParameter(kindRe, line, "$kind")
			if kind == "" {
				err := fmt.Errorf("Something wrong when extracting the kind, the line is %s\n", line)
				return err
			}
			objectID.kind = kind
			isEmpty = false
			continue
		}
		if valueLineCommentRe.MatchString(line) { //not the end
			if isEmpty == true {
				err := fmt.Errorf("There is no objectID info")
				return err
			}
			index, key, err := GetIndexAndKey(valueLineCommentRe, line)
			if err != nil {
				return err
			}
			// check index == real_len(fields) + 1
			// add the new
			// index <= real_len(fields)
			// delete the old and add the new
			// write to the file
			if index == realLen {
				fields[index] = key
				realLen = realLen + 1
			} else if index <= realLen-1 {
				tempField := make([]string, 10)
				copy(tempField, fields[0:index])
				fields = tempField
				fields[index] = key
				realLen = index + 1
			} else if index == realLen+1 { // - name:
				fields[index-1] = key
				realLen = realLen + 1
			} else {
				err := fmt.Errorf("Something wrong with the yaml format")
				return err
			}
			err = node.writeNote(valueLineCommentRe, line, fields, objectID, index, true, true)
			if err != nil {
				err = fmt.Errorf("valueLineCommentRe Error: %w", err)
				return err
			}
			err = node.writeNodeToFile(path)
			if err != nil {
				err = fmt.Errorf("writeNodeToFile Error: %w", err)
				return err
			}
			continue
		} else if lineCommentRe.MatchString(line) {
			if isEmpty == true {
				err := fmt.Errorf("There is no objectID info")
				return err
			}
			index, key, err := GetIndexAndKey(lineCommentRe, line)
			if err != nil {
				return err
			}
			// check index == real_len(fields)
			// add the new
			// index <= real_len(fields) -1
			// delete the old and add the new
			if index == realLen {
				fields[index] = key
				realLen = realLen + 1
			} else if index <= realLen-1 {
				tempField := make([]string, 10)
				copy(tempField, fields[0:index])
				fields = tempField
				fields[index] = key
				realLen = index + 1
			} else {
				err := fmt.Errorf("Something wrong with the yaml format")
				return err
			}
			err = node.writeNote(lineCommentRe, line, fields, objectID, index, false, true)
			if err != nil {
				err = fmt.Errorf("lineCommentRe Error: %w", err)
				return err
			}
			err = node.writeNodeToFile(path)
			if err != nil {
				err = fmt.Errorf("writeNodeToFile Error: %w", err)
				return err
			}
			continue
		} else if valueLineRe.MatchString(line) {
			//if length equals, then write to the value, else(return err)
			index, key, err := GetIndexAndKey(valueLineRe, line)
			if err != nil {
				return err
			}
			if isEmpty == true {
				err := fmt.Errorf("There is no objectID info")
				return err
			}
			// check index = real_len(fields) +1 || index = real_len(fields)
			//writeNode
			//writeToFile
			//else something wrong with the file
			if index == realLen {
				fields[index] = key
				realLen = realLen + 1
			} else if index == realLen-1 {
				tempField := make([]string, 10)
				copy(tempField, fields[0:index])
				fields = tempField
				fields[index] = key
				realLen = index + 1
			} else {
				err := fmt.Errorf("Something wrong with the yaml format")
				return err
			}
			err = node.writeNote(valueLineRe, line, fields, objectID, index, true, false)
			if err != nil {
				err = fmt.Errorf("valueLineRe Error: %w", err)
				return err
			}
			err = node.writeNodeToFile(path)
			if err != nil {
				err = fmt.Errorf("writeNodeToFile Error: %w", err)
				return err
			}
			continue
		} else {
			// for the debug
			log.Printf("Passed the unimportant line: %s", line)
		}
	}
	return nil
}

func extractParameter(re *regexp.Regexp, line string, keyword string) string {
	submatch := re.FindStringSubmatchIndex(line)
	result := []byte{}
	result = re.ExpandString(result, keyword, line, submatch)
	parameter := string(result)
	return parameter
}

func exactHierarchyAndKey(re *regexp.Regexp, line string) (string, string) {
	hierarchy := extractParameter(re, line, "$hierarchy")
	key := extractParameter(re, line, "$key")
	return hierarchy, key
}

func checkHierarchy(hierarchy string) int {
	return int(len(hierarchy) / 2)
}

func GetIndexAndKey(re *regexp.Regexp, line string) (int, string, error) {
	hierarchy, key := exactHierarchyAndKey(re, line)
	// it is possible hierarchy equals to ""
	if hierarchy == "" && key == "" {
		err := fmt.Errorf("Something wrong when extracting the hierarchy or key, the line is %s\n", line)
		return 0, "", err
	}
	index := checkHierarchy(hierarchy)
	return index, key, nil
}
