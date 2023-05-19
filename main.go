package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var (
	dbConfig         string
	tableForChanges  string = "change_in_repository"
	tableForCommands string = "commands"
)

func main() {
	// Read path from config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Fatal error config file:", err)
		os.Exit(1)
	}

	// Open the file to record logs
	file, err := os.OpenFile(viper.GetString("log_file"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Setting up the logger to write to the file
	log.SetOutput(file)

	dbConfig = fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("db.HOST"), viper.GetInt("db.PORT"), viper.GetString("db.USER"),
		viper.GetString("db.PASSWORD"), viper.GetString("db.DBNAME"), viper.GetString("db.SSLMODE"),
	)

	if err := createTable(tableForChanges); err != nil {
		panic(err)
	}

	if err := createTable(tableForCommands); err != nil {
		panic(err)
	}

	// Initialize fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error initializing fsnotify watcher:", err)
		os.Exit(1)
	}
	defer watcher.Close()

	// Add path to watch
	path := viper.GetString("path")
	err = watcher.Add(path)
	if err != nil {
		fmt.Println("Error adding path to watch:", err)
		os.Exit(1)
	}

	fmt.Println("start!")

	// Handle file system events
	var wg sync.WaitGroup
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}


				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("Create:", event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("Delete:", event.Name)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("Rename:", event.Name)
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Modify:", event.Name)
				}

				includeRegexp := viper.GetStringSlice("include_regexp")
				excludeRegexp := viper.GetStringSlice("exclude_regexp")

				if matchesInclude(includeRegexp, event.Name) && !matchesInclude(excludeRegexp, event.Name) {
					//save in db
					if err := collectData(tableForChanges, event.Name, event.Op.String()); err != nil {
						fmt.Println("error saving data about changes in the repository:", err)
					}
					
					//running commands
					err := executeCommands()
					if err != nil {
						fmt.Println("Error executing commands:", err)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	// Handle interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nExiting program...")
		done <- true
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-time.After(1 * time.Minute):
				err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						fmt.Println("Error walking path:", err)
						return err
					}
					if info.Mode().IsRegular() {
						err = watcher.Add(path)
						if err != nil {
							fmt.Println("Error adding path to watch:", err)
							return err
						}
					}
					return nil
				})
				if err != nil {
					fmt.Println("Error walking path:", err)
				}
			case <-done:
				return
			}
		}
	}()
	wg.Wait()
}

// Execute commands
func executeCommands() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	commands := viper.GetStringSlice("commands")
	for _, command := range commands {
		cmd := exec.Command( "sh", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}

		log.Println(command)

		if err := collectData(tableForCommands, "command", command); err != nil {
			fmt.Println("error saving data about commands:", err)
		}
	}

	return nil
}

func collectData(tableName string, file_or_rep_name string, event string) error {

	db, err := sql.Open("postgres", dbConfig)
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf(`INSERT INTO %v (name, event) VALUES ($1, $2)`, tableName)

	_, err = db.Exec(query, file_or_rep_name, event)
	if err != nil {
		log.Println("error inserting data:", err)
		return err
	}

	return nil
}

func createTable(tableName string) error {
	db, err := sql.Open("postgres", dbConfig)
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		time_of_change timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
		name TEXT,
		event TEXT
	)`, tableName)

	_, err = db.Exec(query)
	if err != nil {
		log.Println("error creating table:", err)
		return err
	}

	return nil
}

func matchesInclude(pattern []string, event string) bool {
	//ext := filepath.Ext(event)
	for _, v := range pattern {
		matched, _ := regexp.MatchString(v,event)
		if matched {
			return true
		} 
	}
	return false
}
