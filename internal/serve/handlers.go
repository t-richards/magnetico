package serve

import (
	"errors"
	"net/http"

	"github.com/t-richards/magnetico/internal/persistence"
)

// Homepage.
func rootHandler(database persistence.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nTorrents, err := database.GetNumberOfTorrents()
		if err != nil {
			handlerError(errors.New("GetNumberOfTorrents "+err.Error()), w)
			return
		}

		_ = templates["homepage"].Execute(w, struct {
			NTorrents uint
		}{
			NTorrents: nTorrents,
		})
	}
}

func torrentsHandler(database persistence.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lastId := 0.0
		lastVal := uint64(0)
		_ = r.ParseForm()

		metadata, err := database.QueryTorrents(
			r.Form.Get("query"),
			0,
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

		_ = templates["homepage"].Execute(w, struct {
			Torrents []persistence.TorrentMetadata
		}{
			Torrents: metadata,
		})
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