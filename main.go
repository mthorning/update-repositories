package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
)

func runCommand(path string) {
	cmd := exec.Command("npm install")
	cmd.Dir = path
	cmd.Run()
}

func main() {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}
	absPath, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(absPath)
	fmt.Println(len(files))

	for _, f := range files {
		fmt.Println(f.Name())
		runCommand(absPath + "/" + f.Name())
	}
}
