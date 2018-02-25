package main

import (
	"fmt"
	"os"
	"regexp"

	"go/parser"
	"go/token"
	"path/filepath"
)

var docRegex = regexp.MustCompile(">>>\\s?doc\n([\\s\\S]+)\n<<<\\s?doc")

func factoryCommandFiles(root string) ([]string, error) {
	commandFiles := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		commandFiles = append(commandFiles, path)
		return nil
	})
	return commandFiles, err
}

func processComments(out, file string) {
	fmt.Printf("----- %s\n", file)
	t := token.NewFileSet()
	a, err := parser.ParseFile(t, file, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	for _, comments := range a.Comments {
		docs := docRegex.FindAllStringSubmatch(comments.Text(), -1)
		for _, d := range docs {
			fmt.Println(d[1])
		}
	}
}

func main() {
	cwd, _ := os.Getwd()
	in := filepath.Join(cwd, "command")
	out := filepath.Join(cwd, "doc")
	commandFiles, err := factoryCommandFiles(in)
	if err != nil {
		panic(err)
	}
	for _, file := range commandFiles {
		processComments(out, file)
	}
}
