package serve

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/t-richards/magnetico/internal/persistence"
)

// Homepage.
type homepageData struct {
	NTorrents uint
}

func rootHandler(database persistence.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nTorrents, err := database.GetNumberOfTorrents()
		if err != nil {
			handlerError(errors.New("GetNumberOfTorrents "+err.Error()), w)
			return
		}

		err = templates["homepage"].Execute(w, homepageData{
			NTorrents: nTorrents,
		})
		if err != nil {
			log.Printf("while executing homepage template: %v", err)
		}
	}
}

// Torrents search page
type torrentsData struct {
	Torrents []persistence.TorrentMetadata
	Query    string
}

func torrentsHandler(database persistence.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lastId := 0.0
		lastVal := uint64(0)
		_ = r.ParseForm()

		metadata, err := database.QueryTorrents(
			r.FormValue("query"),
			time.Now().Unix(),
			persistence.ByDiscoveredOn,
			true,
			100,
			&lastId,
			&lastVal,
		)
		if err != nil {
			handlerError(errors.New("QueryTorrents "+err.Error()), w)
			return
		}

		err = templates["torrents"].Execute(w, torrentsData{
			Torrents: metadata,
			Query:    r.FormValue("query"),
		})
		if err != nil {
			log.Printf("while executing torrents template: %v", err)
		}
	}
}

func torrentsInfohashHandler(w http.ResponseWriter, r *http.Request) {
	data := mustAsset("templates/torrent.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func handlerError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}

func emptyFaviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
}
