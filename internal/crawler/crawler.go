package crawler

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/t-richards/magnetico/internal/dht"
	"github.com/t-richards/magnetico/internal/metadata"
	"github.com/t-richards/magnetico/internal/persistence"
)

type crawlerOpts struct {
	IndexerAddrs        []string
	IndexerInterval     time.Duration
	IndexerMaxNeighbors uint

	LeechMaxN int
}

func Run(database persistence.Database) {
	// Hardcoded options for now.
	opts := crawlerOpts{
		IndexerAddrs:        []string{"0.0.0.0:0"},
		IndexerInterval:     1 * time.Second,
		IndexerMaxNeighbors: 1000,
		LeechMaxN:           50,
	}

	// Handle Ctrl-C gracefully.
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	trawlingManager := dht.NewManager(opts.IndexerAddrs, opts.IndexerInterval, opts.IndexerMaxNeighbors)
	metadataSink := metadata.NewSink(5*time.Second, opts.LeechMaxN)

	// The "event loop".
	for {
		select {
		case <-interruptChan:
			trawlingManager.Terminate()
			return

		case result := <-trawlingManager.Output():
			infoHash := result.InfoHash()

			exists, err := database.DoesTorrentExist(infoHash[:])
			if err != nil {
				log.Fatalf("Could not check whether torrent exists! %V", err)
			} else if !exists {
				metadataSink.Sink(result)
			}

		case md := <-metadataSink.Drain():
			if err := database.AddNewTorrent(md.InfoHash, md.Name, md.Files); err != nil {
				log.Fatalf("Could not add new torrent to the database. %v", err)
			}
		}
	}
}
