// magnetico is the universal build with both the DHT crawler and the web interface.
package main

import (
	"log"

	"github.com/t-richards/magnetico/internal/crawler"
	"github.com/t-richards/magnetico/internal/persistence"
	"github.com/t-richards/magnetico/internal/serve"
)

const (
	DatabasePath = "data/magnetico.db"
)

func main() {
	// open the database
	database, err := persistence.NewSqlite3Database(DatabasePath)
	if err != nil {
		log.Fatalf("Could not open the database %s. %v", DatabasePath, err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Could not close database! %v", err)
		}
	}()

	// launch the web service in the background
	go serve.Run(database)

	// run the crawler with primary interrupt handling logic
	crawler.Run(database)
}
