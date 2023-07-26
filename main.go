// magnetico is the universal build with both the DHT crawler and the web interface.
package main

import (
	"log"

	"github.com/t-richards/magnetico/internal/persistence"
	"github.com/t-richards/magnetico/internal/serve"
)

const (
	DatabasePath = "data/magnetico.db"
)

func main() {
	// open the database
	database, err := persistence.MakeDatabase(DatabasePath)
	if err != nil {
		log.Fatalf("Could not open the database %s. %v", DatabasePath, err)
	}
	defer database.Close()

	// run the crawler in the background
	// TODO(tom): Fix this
	// crawler.Run(database)

	// launch the web service
	serve.Run(database)
}
