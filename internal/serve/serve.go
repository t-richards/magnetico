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

//go:embed static/*
var static embed.FS

//go:embed templates/*
var fs embed.FS

// Shared template functions across all templates.
var templateFunctions = template.FuncMap{
	"comma": func(s any) string {
		switch v := s.(type) {
		case uint64:
			return humanize.Comma(int64(v))
		case int64:
			return humanize.Comma(v)
		case uint:
			return humanize.Comma(int64(v))
		case int:
			return humanize.Comma(int64(v))
		default:
			return "unknown"
		}
	},

	"hex": hex.EncodeToString,

	"humanizeTime": func(s int64) string {
		return humanize.Time(time.Unix(s, 0))
	},

	"humanizeSize": func(i any) string {
		switch v := i.(type) {
		case uint64:
			return humanize.IBytes(v)
		case int64:
			return humanize.IBytes(uint64(v))
		case uint:
			return humanize.IBytes(uint64(v))
		case int:
			return humanize.IBytes(uint64(v))
		default:
			return "unknown"
		}
	},

	"unixTimeToString": func(s int64) string {
		tm := time.Unix(s, 0)
		// > Format and Parse use a reference time for specifying the format.
		// https://gobyexample.com/time-formatting-parsing
		return tm.Format("2006-01-02 15:04:05 MST")
	},
}

const (
	BindAddress = ":8080"
)

func Run(database *persistence.Database) {
	// Main application routes
	router := chi.NewRouter()
	router.Use(securityHeaders)
	router.Get("/", rootHandler(database))
	router.Get("/static/*", staticHandler)
	router.Get("/favicon.ico", emptyFaviconHandler)
	router.Get("/torrents", torrentsHandler(database))
	router.Get("/torrents/{infohash:[a-f0-9]{40}}", torrentsInfohashHandler(database))

	log.Printf("magnetico is ready to serve on %s!", BindAddress)
	err := http.ListenAndServe(BindAddress, router)
	if err != nil {
		log.Printf("ListenAndServe error %v", err)
	}
}

func mustTemplate(name string) string {
	data, err := fs.ReadFile(name)
	if err != nil {
		log.Panic(err)
	}
	return string(data)
}
