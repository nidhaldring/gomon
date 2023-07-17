package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func runGoProg() *exec.Cmd {
	cmd := exec.Command("go", "run", ".")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
	return cmd
}

func isGoFile(filename string) bool {
	return strings.Index(filename, ".go") == len(filename)-3
}

func shouldRestart(e fsnotify.Event) bool {
	return isGoFile(e.Name) && (e.Has(fsnotify.Create) ||
		e.Has(fsnotify.Write) ||
		e.Has(fsnotify.Remove) ||
		e.Has(fsnotify.Rename))
}

func watchGoProg(watcher *fsnotify.Watcher, dir string) {
	// HACK: go run dirname does not work for some reason
	// the quickest solution now is to chdir to the dir
	// and run the program from there
	//TODO: find a better solution!
	os.Chdir(dir)

	fmt.Printf("Starting watching %s...\n", dir)
	cmd := runGoProg()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if shouldRestart(event) {
				fmt.Println("[Info]: Detected a change: restarting...")
				cmd.Process.Kill()
				cmd = runGoProg()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			if err != nil {
				panic(err)
			}
		}
	}

}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Must specify directory name!")
		os.Exit(1)
	}

	dir := os.Args[1]

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	go watchGoProg(watcher, dir)

	err = watcher.Add(dir)
	if err != nil {
		panic(err)
	}

	<-make(chan any)
}
