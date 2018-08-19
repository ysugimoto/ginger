// +build ignore
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"go/parser"
	"go/token"
	"io/ioutil"
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

func processComments(file string) []string {
	fmt.Printf("Processing %s...\n", file)
	t := token.NewFileSet()
	a, err := parser.ParseFile(t, file, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	apis := []string{}
	for _, comments := range a.Comments {
		docs := docRegex.FindAllStringSubmatch(comments.Text(), -1)
		for _, d := range docs {
			apis = append(apis, d[1])
		}
	}
	return apis
}

func main() {
	cwd, _ := os.Getwd()
	in := filepath.Join(cwd, "command")
	commandFiles, err := factoryCommandFiles(in)
	if err != nil {
		panic(err)
	}
	apis := []string{}
	for _, file := range commandFiles {
		a := processComments(file)
		apis = append(apis, a...)
	}

	out := filepath.Join(cwd, "docs/command.md")
	body := strings.Join(apis, "\n")
	if err := ioutil.WriteFile(out, []byte("<!-- This document generated automatically -->\n"+body), 0666); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
