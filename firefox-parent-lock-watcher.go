package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"regexp"
	"os"
	"gopkg.in/ini.v1"
	"fmt"
	"strings"
	"path/filepath"
	"path"
	"sync"
)

var (
	ProfilesList = make(map[string]*ProfileEntry, 0)
	pattern, _ = regexp.Compile("\\.parentlock$")
)

type ProfileEntry struct {
	IniName string
	Name string
	IsRelative bool
	Path string
}

func FindProfile(name string) *ProfileEntry {
	mutex := sync.RWMutex{}
	mutex.RLock()
	defer mutex.RUnlock()
	for _, profile := range ProfilesList {
		profileDir, _ := filepath.Abs(profile.Path)
		searchDir, _ := filepath.Abs(filepath.Dir(name))
		if profileDir == searchDir {
			return profile
		}
	}
	return nil
}

func InstallWatchers() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {

		for {
			select {
			case event := <-watcher.Events:
				if pattern.Match([]byte(event.Name)) {
					_, err := os.Stat(event.Name)
					if !os.IsNotExist(err) {
						err = os.Remove(event.Name)
						if err == nil {
							if profileName := FindProfile(event.Name); profileName != nil {
								log.Printf("Removed '.parentlock' from profile '%s'\n", profileName.Name)
							} else {
								log.Printf("Removed '.parentlock' from profile path '%s'\n", event.Name)
							}
						} else {
							if profileName := FindProfile(event.Name); profileName != nil {
								log.Printf("Can't remove '.parentlock' from profile '%s'. Error: %s\n",
									profileName.Name, err.Error())
							} else {
								log.Printf("Match found but can't remove file '%s', error: %s\n",
									event.Name, err.Error())
							}
						}
					}
				}

			case err := <-watcher.Errors:
				log.Printf("Watcher error: %s\n", err.Error())
			}
		}
	}()

	mutex := sync.Mutex{}
	mutex.Lock()
	for _, profile := range ProfilesList {
		err = watcher.Add(profile.Path)
		if err != nil {
			log.Printf("Failed to add watcher for Firefox Profile '%s' (path: '%s'). Error: %s\n",
				profile.Name, profile.Path, err.Error())
		} else {
			log.Printf("Installing watcher to Firefox Profile '%s' (path: '%s')\n",
				profile.Name, profile.Path)
		}
	}
	mutex.Unlock()
	<-done
}

func main()  {
	if len(os.Args) < 2 {
		return
	}

	configFile := os.Args[1]
	dirFile, _ := filepath.Abs(filepath.Dir(configFile))

	cfg, err := ini.Load(configFile)
	if err != nil {
		fmt.Printf("Fail to read config file '%s': %s\n", configFile, err.Error())
		os.Exit(-1)
	}

	for _, profile := range cfg.Sections() {
		IniName := profile.Name()
		Name := profile.Key("Name").String()
		if strings.TrimSpace(Name) == "" {
			continue
		}
		IsRelative, _ := profile.Key("IsRelative").Bool()
		Path := profile.Key("Path").String()

		if IsRelative {
			Path, _ = filepath.Abs(path.Join(dirFile, Path))
		}

		ProfilesList[IniName] = &ProfileEntry{
			IniName: IniName,
			Name: Name,
			IsRelative: IsRelative,
			Path: Path,
		}
	}

	InstallWatchers()
}
