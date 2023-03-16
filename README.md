# This is a Demo

## Problem

Mapping file contains mapping of every field in the object from the templated chart to the original file and line.

Source: [issue](https://github.com/helm/helm/issues/11552)

## Run

Go version: 1.20

    go run . -p <project_dir_path> -o <result_dir_path>

## Process

- Create a command line tool which adds comments to all template files (templates directory)

- Run the tool on the chart directory
  
- Run helm template (save the output in a file)

- Create another tool which does two things with the templated chart: Creates a mapping file and Removes our comments

## Result

Add "value"

    {
        "field": "metadata.labels.app.kubernetes.io/name",
        "file": "service.yaml",
        "line": 5,
        "objectID": "{\"apiVersion\":\"v1\",\"kind\":\"Service\"}",
        "value": "example"
    }

## addCommentByLine vs addCommentInLine

### addCommentByLine

    #This is the 1 line in deployment.yaml
    apiVersion: apps/v1
    #This is the 2 line in deployment.yaml
    kind: Deployment

### addCommentInLine

    apiVersion: apps/v1 #This is the 1 line in deployment.yaml
    kind: Deployment #This is the 2 line in deployment.yaml

## Regular expression

Basically used Regular expression in the main part

Use space in the line to calculate the hierarchy, each two space adding one hierarchy

### valueLineCommentRe

example:

    app.kubernetes.io/managed-by: Helm #This is the 7 line in serviceaccount.yaml

It will be the end part of the chain, add to the file

### valueLineRe

example:

    helm.sh/chart: example-0.1.0

It will be the end part of the chain, add to the file

### LineCommentRe

example:

    metadata: #This is the 4 line in serviceaccount.yaml

It must be the middle part in the chain, don't have to write to File

## Special condition

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

## Problems

1. A smarter way to get the fields in result
2. unit test
3. performance
4. just baesd on the example, need modification for the general
