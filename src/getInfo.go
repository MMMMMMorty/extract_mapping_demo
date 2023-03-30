package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// change to use go func
func getInfo(dirPath string) error {
	resultPath := dirPath + "/example.yaml"
	lines, _, err := ReadLines(resultPath)
	if err != nil {
		return err
	}

	soueceFileRe := regexp.MustCompile(SOURCE_FILE_PATTERN)
	apiVersionRe := regexp.MustCompile(`apiVersion: (?P<apiVersion>\S*)`)
	kindRe := regexp.MustCompile(`kind: (?P<kind>\S*)`)
	rePath := regexp.MustCompile(`path: (?P<path>\S*)`)
	reType := regexp.MustCompile(`type: '(?P<type>\S*)'`)
	reValue := regexp.MustCompile(`value: (?P<value>\[.+\]|\S*)`)
	reLine := regexp.MustCompile(`This is the (?P<line>\d*) line`)
	lastNumber := -1
	isEmpty := true

	node := new(Node)
	objectID := new(ObjectID)

	for index, line := range lines {
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
			lastNumber = -1
			continue
		} else if apiVersionRe.MatchString(line) {
			apiVersion := extractParameter(apiVersionRe, line, "$apiVersion")
			if apiVersion == "" {
				err := fmt.Errorf("Something wrong when extracting the apiVersion, the line is %s\n", line)
				return err
			}
			objectID.apiVersion = apiVersion
			isEmpty = false
			continue
		} else if kindRe.MatchString(line) {
			kind := extractParameter(kindRe, line, "$kind")
			if kind == "" {
				err := fmt.Errorf("Something wrong when extracting the kind, the line is %s\n", line)
				return err
			}
			objectID.kind = kind
			isEmpty = false
			continue
		} else if isEmpty == false {
			// not sure if it can go to the end
			index = index - 3
			command := `'[..| select(line == ` + strconv.Itoa(index) + `)| {"destpath": path | join("."),"type": type,"value": .}]'`
			yqCommand := fmt.Sprintf("yq %s %s", command, resultPath)
			out, _ := Exec(yqCommand)
			if out != nil {
				output := out.String()
				fmt.Println(output)
				path := extractParameter(rePath, output, "$path")
				//if path is empty, continue
				if path != "" && path != "\"\"" {
					if isEmpty == true {
						err := fmt.Errorf("There is no objectID info")
						return err
					}
					splits := strings.Split(output, "- dest")
					if len(splits) < 2 {
						err := fmt.Errorf("Something wrong with the length of the splits, which is %d", len(splits))
						return err
					} else {
						// cut the redundant one
						splits = splits[1:]
						for _, split := range splits {
							path := extractParameter(rePath, split, "$path")
							pathType := extractParameter(reType, split, "$type")
							mapMatched, err := regexp.MatchString(`!!map`, pathType)
							if err != nil {
								err = fmt.Errorf("regexp.MatchString err: %s", err.Error())
							}
							if mapMatched {
								err := writeNodeToFile(reValue, reLine, split, lastNumber, path, dirPath, node, objectID, true)
								if err != nil {
									err = fmt.Errorf("map type: writeNodeToFile wrong err: %s", err.Error())
								}
							} else {
								err := writeNodeToFile(reValue, reLine, split, lastNumber, path, dirPath, node, objectID, false)
								if err != nil {
									err = fmt.Errorf("not map type: writeNodeToFile wrong err: %s", err.Error())
								}
							}

						}
					}
				}
			}

		}
	}
	return nil
}

func writeNodeToFile(reValue *regexp.Regexp, reLine *regexp.Regexp, split string, lastNumber int, path string, dirPath string, node *Node, objectID *ObjectID, isMapType bool) error {
	value, lineNumber, newLastNumber, err := getInfoFromOne(reValue, reLine, split, lastNumber, isMapType)
	if err != nil {
		err = fmt.Errorf("getInfoFromOne wrong err: %s", err.Error())
		return err
	}
	lastNumber = newLastNumber
	node = writeInfoToNode(node, objectID, path, lineNumber, value)
	err = node.writeNodeToFile(dirPath)
	if err != nil {
		err = fmt.Errorf("node.writeToFile wrong err: %s", err.Error())
		return err
	}
	return nil
}

func getInfoFromOne(reValue *regexp.Regexp, reLine *regexp.Regexp, output string, lastNumber int, isMapType bool) (value string, lineNumber int, newLastNumber int, err error) {
	if isMapType == true {
		value = ""
	} else {
		value = extractParameter(reValue, output, "$value")
	}
	number := extractParameter(reLine, output, "$line")
	if number != "" {
		lineNumber, err = strconv.Atoi(number)
		if err != nil {
			err = fmt.Errorf("strconv.Atoi err: %s", err.Error())
			return "", -1, -1, err
		}
		if isMapType == true {
			lineNumber = lineNumber - 1
		}
		lastNumber = lineNumber
		// save to structure
	} else {
		lineNumber = lastNumber
		// use the last one number
	}
	newLastNumber = lineNumber
	return value, lineNumber, newLastNumber, nil
}

func writeInfoToNode(node *Node, objectID *ObjectID, path string, lineNumber int, value string) *Node {
	node.field = path
	node.line = lineNumber
	node.objectID = objectID
	node.value = value
	return node
}
