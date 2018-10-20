package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const fileNameBase = "nexcenter"

type Folder struct {
	base, platform, app, path string
}

func (f *Folder) checkoutDevelop() {
	fmt.Printf("%s: checking out develop\n", f.app)
	checkout := []string{"checkout", "develop"}
	f.runCommand(f.path, "git", checkout)
}

func (f *Folder) pullDevelop() {
	fmt.Printf("%s: pulling develop\n", f.app)
	pull := []string{"pull"}
	f.runCommand(f.path, "git", pull)
}

func (f *Folder) npmInstall() {
	fmt.Printf("%s: installing node modules\n", f.app)
	install := []string{"install"}
	f.runCommand(f.path, "npm", install)
}

func (f *Folder) build() {
	cmd := "node_modules/.bin/webpack"
	args := []string{"--config", "webpack.dev.js", "--progress"}
	f.runCommand(f.path, cmd, args)
}

func (f *Folder) runCommand(path, command string, args []string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	stdin, err := cmd.StdinPipe()
	checkError(err)
	go func() {
		defer stdin.Close()
	}()
	out, err := cmd.CombinedOutput()
	checkError(err)
	fmt.Printf("%s: %s\n", f.app, out)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getFolders(absPath string, files []os.FileInfo) map[string]Folder {
	folders := make(map[string]Folder)
	for _, f := range files {
		// Only send folders we want to runCommand
		splitName := strings.Split(f.Name(), "-")
		base := splitName[0]
		platform := splitName[1]
		app := splitName[2]
		if base == fileNameBase {
			path := filepath.Join(absPath, f.Name())
			folders[app] = Folder{base, platform, app, path}
		}
	}
	return folders
}

//ignore cores
//npm link cores

func main() {
	files, err := ioutil.ReadDir("./")
	checkError(err)
	absPath, err := filepath.Abs("./")
	checkError(err)
	for _, folder := range getFolders(absPath, files) {
		folder.checkoutDevelop()
		folder.pullDevelop()
		folder.npmInstall()
		folder.build()
	}
}
