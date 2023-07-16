package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func runGoProg() *exec.Cmd {
	cmd := exec.Command("go", "run", ".")
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

func watchGoProg(watcher *fsnotify.Watcher) {
	fmt.Println("Starting watching...")
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
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	go watchGoProg(watcher)

	err = watcher.Add(".")
	if err != nil {
		panic(err)
	}

	<-make(chan any)
}
