package serve

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/t-richards/magnetico/internal/persistence"
)

// Homepage.
type homepageData struct {
	NTorrents uint
}

func rootHandler(database persistence.Database) http.HandlerFunc {
	homepageTemplate := template.Must(template.New("homepage").Funcs(templateFunctions).Parse(mustTemplate("templates/homepage.html")))

	return func(w http.ResponseWriter, r *http.Request) {
		nTorrents, err := database.GetNumberOfTorrents(r.Context())
		if err != nil {
			handlerError(errors.New("GetNumberOfTorrents "+err.Error()), w)
			return
		}

		err = homepageTemplate.Execute(w, homepageData{
			NTorrents: nTorrents,
		})
		if err != nil {
			log.Printf("while executing homepage template: %v", err)
		}
	}
}

// Torrents search page.
type torrentsData struct {
	Torrents []persistence.TorrentMetadata
	Query    string
}

func torrentsHandler(database persistence.Database) http.HandlerFunc {
	listTemplate := template.Must(template.New("torrent").Funcs(templateFunctions).Parse(mustTemplate("templates/torrents.html")))

	return func(w http.ResponseWriter, r *http.Request) {
		lastID := 0.0
		lastVal := uint64(0)
		_ = r.ParseForm()

		metadata, err := database.QueryTorrents(
			r.FormValue("query"),
			time.Now().Unix(),
			persistence.ByDiscoveredOn,
			true,
			15,
			&lastID,
			&lastVal,
		)
		if err != nil {
			log.Printf("while fetching torrents: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = listTemplate.Execute(w, torrentsData{
			Torrents: metadata,
			Query:    r.FormValue("query"),
		})
		if err != nil {
			log.Printf("while executing torrents template: %v", err)
		}
	}
}

type torrentData struct {
	Torrent persistence.TorrentMetadata
	Files   []persistence.File
	Query   string
}

func torrentsInfohashHandler(database persistence.Database) http.HandlerFunc {
	infoTemplate := template.Must(template.New("torrent").Funcs(templateFunctions).Parse(mustTemplate("templates/torrent.html")))

	return func(w http.ResponseWriter, r *http.Request) {
		infohash := chi.URLParam(r, "infohash")
		hashBytes, err := hex.DecodeString(infohash)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		metadata, err := database.GetTorrent(hashBytes)
		if err != nil {
			log.Printf("while fetching torrent: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if metadata == nil {
			http.NotFound(w, r)
			return
		}

		files, err := database.GetFiles(hashBytes)
		if err != nil {
			log.Printf("while fetching files: %v\n", err)
		}

		err = infoTemplate.Execute(w, torrentData{
			Torrent: *metadata,
			Files:   files,
			Query:   r.FormValue("query"),
		})
		if err != nil {
			log.Printf("while executing torrent template: %v", err)
		}
	}
}

func handlerError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}

func emptyFaviconHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	inputPath := strings.TrimPrefix(r.URL.Path, "/")
	file, err := static.Open(inputPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	defer file.Close()

	// Set the content type based on the file extension.
	if strings.HasSuffix(r.URL.Path, "webp") {
		w.Header().Set("Content-Type", "image/webp")
	}

	_, err = io.Copy(w, file)
	if err != nil {
		log.Printf("while serving static file: %v", err)
	}
}
