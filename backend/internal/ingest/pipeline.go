package ingest

import (
	"log"
	"time"

	"corelog/backend/internal/api"
	"corelog/backend/internal/db"
)

func StartPipeline(logPath string, useDocker bool) {
	// Channel connects Parser -> DB Batcher
	outChan := make(chan api.LogEntry, 5000)

	// Create Parser
	parser := NewParser(outChan)

	if useDocker {
		StartDockerCollector(parser)
	} else {
		StartTailing(logPath, parser)
	}

	// Start Batch Inserter Goroutine
	go func() {
		var batch []api.LogEntry
		ticker := time.NewTicker(2 * time.Second)

		for {
			select {
			case entry := <-outChan:
				// 1. Send straight to WebSockets for live UI update
				api.Hub.Broadcast(api.LogEntry{
					Timestamp: entry.Timestamp,
					Level:     entry.Level,
					SourceApp: entry.SourceApp,
					Message:   entry.Message,
				})

				// 2. Buffer for massive Database Inserts
				batch = append(batch, entry)
				if len(batch) >= 1000 {
					flushBatch(&batch)
				}
			case <-ticker.C:
				if len(batch) > 0 {
					flushBatch(&batch)
				}
			}
		}
	}()
}

func flushBatch(batch *[]api.LogEntry) {
	if len(*batch) == 0 {
		return
	}

	// Begin Transaction for speed
	tx, err := db.DB.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO logs (timestamp, level, source_app, message) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Println("Error preparing statement:", err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for _, entry := range *batch {
		_, err := stmt.Exec(entry.Timestamp, entry.Level, entry.SourceApp, entry.Message)
		if err != nil {
			log.Println("Error inserting log:", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println("Error committing batch:", err)
	}

	// Reset batch slice (reuse capacity)
	*batch = (*batch)[:0]
}
