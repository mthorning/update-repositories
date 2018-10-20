package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const folderNameBase = "nexcenter"

type Folder struct {
	base, platform, app, path string
}

func (f *Folder) checkoutDevelop() {
	fmt.Printf("%s app: checking out develop\n", f.app)
	args := []string{"checkout", "develop"}
	f.runCommand(f.path, "git", args)
}

func (f *Folder) pullDevelop() {
	fmt.Printf("%s app: pulling develop\n", f.app)
	f.runCommand(f.path, "git", []string{"pull"})
}

func (f *Folder) npmInstall() {
	fmt.Printf("%s app: installing node modules, please wait\n", f.app)
	f.runCommand(f.path, "npm", []string{"install"})
}

func (f *Folder) build() {
	cmd := "node_modules/.bin/webpack"
	args := []string{"--config", "webpack.dev.js", "--progress"}
	f.runCommand(f.path, cmd, args)
}

func (f *Folder) linkToCore() {
	core := fmt.Sprintf("%s-%s-core", f.base, f.platform)
	args := []string{"link", core}
	f.runCommand(f.path, "npm", args)
	fmt.Printf("%s app: linked to %s\n", f.app, core)
}

func (f *Folder) runCommand(path, command string, args []string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = path
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to run command %s (StdinPipe): %s", command, err)
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run command %s (CombinedOutput): %s", command, err)
	}
	fmt.Printf("%s app: %s\n", f.app, out)
}

func getFolders(absPath string, files []os.FileInfo) map[string]Folder {
	folders := make(map[string]Folder)
	for _, f := range files {
		// Only send folders of the nexcenter app
		var platform, app string
		splitName := strings.Split(f.Name(), "-")
		base := splitName[0]
		if len(splitName) == 3 && base == folderNameBase {
			platform = splitName[1]
			app = splitName[2]
			path := filepath.Join(absPath, f.Name())
			folders[app] = Folder{base, platform, app, path}
		}
	}
	return folders
}

func processApp(f Folder) {
	f.checkoutDevelop()
	f.pullDevelop()
	f.npmInstall()
	// f.build()
	f.linkToCore()
}

func processCore(f Folder) {
	f.checkoutDevelop()
	f.pullDevelop()
	f.npmInstall()
	// Create NPM link
	f.runCommand(f.path, "npm", []string{"link"})
	fmt.Printf("NPM link successfully created for %s-%s-%s\n\n", f.base, f.platform, f.app)
}

func main() {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatalf("Failed to read files, %v", err)
	}
	absPath, err := filepath.Abs("./")
	if err != nil {
		log.Fatalf("Failed to get absolute path. %v", err)
	}
	folders := getFolders(absPath, files)
	// Process core first
	processCore(folders["core"])

	var wg sync.WaitGroup
	// How many goRoutines will be run (folders - core)
	wg.Add(len(folders) - 1)
	for _, folder := range folders {
		if folder.app != "core" {
			// Need to pass folder into go routine for async reasons
			go func(folder Folder) {
				defer wg.Done()
				processApp(folder)
			}(folder)
		}
	}
	wg.Wait()
}
