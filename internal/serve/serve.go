package serve

import (
	"embed"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"

	"github.com/t-richards/magnetico/internal/persistence"
)

//go:embed templates/*
var fs embed.FS

var templates map[string]*template.Template

const (
	BindAddress = ":8080"
)

func Run(database persistence.Database) {
	// Main application routes
	router := chi.NewRouter()
	router.Use(noIndex)
	router.Get("/", rootHandler(database))
	router.Get("/favicon.ico", emptyFaviconHandler)
	router.Get("/torrents", torrentsHandler(database))
	router.Get("/torrents/{infohash:[a-f0-9]{40}}", torrentsInfohashHandler)

	templateFunctions := template.FuncMap{
		"hex": hex.EncodeToString,

		"unixTimeToString": func(s int64) string {
			tm := time.Unix(s, 0)
			// > Format and Parse use a reference time for specifying the format.
			// https://gobyexample.com/time-formatting-parsing
			return tm.Format("2006-01-02 15:04:05")
		},

		"humanizeTime": func(s int64) string {
			return humanize.Time(time.Unix(s, 0))
		},

		"humanizeSize": humanize.IBytes,

		"comma": func(s uint) string {
			return humanize.Comma(int64(s))
		},
	}

	templates = make(map[string]*template.Template)
	templates["homepage"] = template.Must(template.New("homepage").Funcs(templateFunctions).Parse(string(mustAsset("templates/homepage.html"))))
	templates["torrent"] = template.Must(template.New("torrent").Funcs(templateFunctions).Parse(string(mustAsset("templates/torrent.html"))))
	templates["torrents"] = template.Must(template.New("torrents").Funcs(templateFunctions).Parse(string(mustAsset("templates/torrents.html"))))

	log.Printf("magnetico is ready to serve on %s!", BindAddress)
	err := http.ListenAndServe(BindAddress, router)
	if err != nil {
		log.Printf("ListenAndServe error %v", err)
	}
}

func mustAsset(name string) []byte {
	data, err := fs.ReadFile(name)
	if err != nil {
		return nil
	}
	return data
}

func noIndex(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}
