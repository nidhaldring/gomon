package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	// TODO: find a better solution!
	os.Chdir(dir)

	fmt.Printf("Starting watching %s...\n", dir)
	cmd := runGoProg()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// if it's a dir event try to add it to the watch list and return
			if isDir(event.Name) && !ignoreDir(event.Name) {
				watcher.Add(event.Name)
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

func ignoreDir(dirname string) bool {
	// ignore hidden dirs + anything in gitignore
	return isHiddenDir(dirname) || isInGitIgnore(dirname)
}

func isDir(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	if stat, err := file.Stat(); err != nil {
		return stat.IsDir()
	}

	return false
}

func addFolder(w *fsnotify.Watcher, dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal("error: ", err)
			return err
		}

		if !d.IsDir() || (d.IsDir() && ignoreDir(path)) {
			fmt.Printf("this is ignored %s\n", path)
			return nil
		}

		w.Add(d.Name())

		return nil
	})
}

func main() {
	dir := flag.String("dir", ".", "--dir dirname   dirname to watch for changes")
	flag.Parse()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	go watchGoProg(watcher, *dir)

	err = addFolder(watcher, *dir)
	if err != nil {
		panic(err)
	}

	<-make(chan any)
}
