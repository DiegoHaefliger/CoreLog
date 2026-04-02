package ingest

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// StartTailing monitors a specific log file and feeds new lines to the parser.
func StartTailing(filePath string, parser *Parser) {
	go func() {
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			log.Println("Warning: could not create log directory:", err)
		}
		
		var file *os.File
		var err error

		// Wait until file is created the first time
		for {
			file, err = os.Open(filePath)
			if err == nil {
				break
			}
			time.Sleep(1 * time.Second)
		}
		defer file.Close()

		// Seek to the end (Don't read historical stuff for Live Tailing in MVP)
		file.Seek(0, io.SeekEnd)

		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// wait for new data without burning CPU
					time.Sleep(500 * time.Millisecond)
					// trigger parser flush if something is pending over time
					parser.Flush()
					continue
				}
				log.Println("Error reading file:", err)
				break
			}

			// Clean the newline char
			parser.ParseLine(line)
		}
	}()
}
