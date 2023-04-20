# This is a Demo

This demo is not general now, it just for the template created by "helm create <project_name>"

## Problem

Mapping file contains mapping of every field in the object from the templated chart to the original file and line.

Source: [issue](https://github.com/helm/helm/issues/11552)

## Run

Go version: 1.20

yq version v4.33.1

    snap install yq

    # install helm first

    helm create <project_name>

    git clone git@github.com:MMMMMMorty/extract_mapping_demo.git

    cd src

    go run . -p <project_dir_path> -o <result_dir_path>

## Process

- Create a command line tool which adds comments to all template files (templates directory)

- Run the tool on the chart directory
  
- Run helm template (save the output in a file)

- Create another tool which does two things with the templated chart: Creates a mapping file and Removes our comments

## Version 1

Using regular expression to calculate the path, it is complex, so change to version 2.
### Result

Add "value"

    {
        "field": "metadata.labels.app.kubernetes.io/name",
        "file": "service.yaml",
        "line": 5,
        "objectID": "{\"apiVersion\":\"v1\",\"kind\":\"Service\"}",
        "value": "example"
    }

### addCommentByLine vs addCommentInLine

#### addCommentByLine

    #This is the 1 line in deployment.yaml
    apiVersion: apps/v1
    #This is the 2 line in deployment.yaml
    kind: Deployment

#### addCommentInLine

    apiVersion: apps/v1 #This is the 1 line in deployment.yaml
    kind: Deployment #This is the 2 line in deployment.yaml

### Regular expression

Basically used Regular expression in the main part

Use space in the line to calculate the hierarchy, each two space adding one hierarchy

#### valueLineCommentRe

example:

    app.kubernetes.io/managed-by: Helm #This is the 7 line in serviceaccount.yaml

It will be the end part of the chain, add to the file

#### valueLineRe

example:

    helm.sh/chart: example-0.1.0

It will be the end part of the chain, add to the file

#### LineCommentRe

example:

    metadata: #This is the 4 line in serviceaccount.yaml

It must be the middle part in the chain, don't have to write to File

### Special condition

1. too many line numbers

        app.kubernetes.io/managed-by: Helm #This is the 7 line in serviceaccount.yaml #This is the 11 line in serviceaccount.yaml #This is the 12 line in serviceaccount.yaml

    only take the first number as lineNumber

2. different yaml format

       spec: #This is the 7 line in service.yaml
         type: ClusterIP #This is the 8 line in service.yaml
         ports: #This is the 9 line in service.yaml
          - port: 80 #This is the 10 line in service.yaml
            targetPort: http #This is the 11 line in service.yaml

Generated answer:

    {"field":"spec.ports.port.","file":"service.yaml","line":10,"objectID":"{\"apiVersion\":\"v1\",\"kind\":\"Service\"}","value":"80"}

    {"field":"spec.ports.port.targetPort","file":"service.yaml","line":11,"objectID":"{\"apiVersion\":\"v1\",\"kind\":\"Service\"}","value":"http"}

The second one should be "spec.ports.targetPort", but now it is "spec.ports.port.targetPort"

### Problems

1. A smarter way to get the fields in result
2. unit test
3. performance
4. just baesd on the example, need modification for the general

## Version 2

In this version, I used [yq](https://github.com/mikefarah/yq) to get the path.

- Drawbacks: Slow. lower accracy of line number

- Benefit: high accuracy of the path

### Ideas about this version

There are two main types of returns based on yq, map type or not map type.

* Map type

            - destpath: metadata.annotations
              type: '!!map'
              value:
                prometheus.io/scrape: 'true' #This is the 10 line in service.yaml 
            - destpath: metadata.annotations.prometheus.io/scrape
              type: '!!str'
              value: 'true' #This is the 10 line in service.yaml


- Seq type (new one, need to fix)

              - destpath: spec.ports
                  type: '!!seq'
                  value:
              - name: "http" #This is the 18 line in service.yaml 
                protocol: TCP #This is the 19 line in service.yaml
                port: 8080 #This is the 20 line in service.yaml #This is the 23 line in service.yaml
                targetPort: 8080 #This is the 24 line in service.yaml 
                #This is the 33 line in service.yaml  #This is the 36 line in service.yaml #This is the 36 line in service.yaml #This is the 42 line in service.yaml #This is the 47 line in service.yaml
              - destpath: spec.ports.0
                  type: '!!map'
                  value:
                    name: "http" #This is the 18 line in service.yaml 
                    protocol: TCP #This is the 19 line in service.yaml
                    port: 8080 #This is the 20 line in service.yaml #This is the 23 line in service.yaml
                    targetPort: 8080 #This is the 24 line in service.yaml
                    #This is the 33 line in service.yaml #This is the 36 line in service.yaml #This is the 42 line in service.yaml #This is the 47 line in service.yaml #This is the 47 line in service.yaml
              - destpath: spec.ports.0.name
                  type: '!!str'
                  value: "http" #This is the 18 line in service.yaml

* Not Map type

        - destpath: spec.ports.0.protocol
            type: '!!str'
            value: TCP #This is the 19 line in service.yaml #This is the 19 line in service.yaml #This is the 19 line in service.yaml #This is the 19 line in service.yaml

### What causes the lower accuracy of line number

    - destpath: metadata.annotations
      type: '!!map'
      value:
        prometheus.io/scrape: 'true' #This is the 10 line in service.yaml

For map type, it doesn't show the line number of the map but the line number of its value. So here I used below expression to calculate:
    line number of the map = line number of its first value - 1

When there are more than 2 map types connecting to each other

    template: #This is the 14 line in deployment.yaml
    metadata: #This is the 15 line in deployment.yaml #This is the 19 line in deployment.yaml
    labels: #This is the 20 line in deployment.yaml

    Here the line number of matedata will be 19 instead of 15, pointing to the end of the metadata

    15  metadata:
    16  {{- with .Values.podAnnotations }}
    17  annotations:
    18    {{- toYaml . | nindent 8 }}
    19  {{- end }}

### The potential problem

1. Program might not reach the last line, if there are no extra comments at the end of the file
    This is because the line number of the call to yq does not match the line number of the actual lines. Further testing is needed here.


### Next

1. to solve the general problem, testing on more templates to fix some small bugs.
2. use go func (concurrency) to improve the speed of the program
3. try to extract the functional part of yq