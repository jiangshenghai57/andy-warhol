package log

import (
	"log"
	"os"
)

func LogToFile(log_path, log_file string) {

	path_file := log_path + log_file

	// Open a file for writing logs. Create it if it doesn't exist, append if it does.
	file, err := os.OpenFile(path_file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Set the output destination for the default logger
	log.SetOutput(file)

	// Now, all logs will go into app.log
	log.Println("This is a log message.")
	log.Printf("Another log entry: %d", 42)
}
