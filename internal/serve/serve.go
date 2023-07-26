package serve

import (
	"embed"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"

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
	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler(database))
	router.HandleFunc("/favicon.ico", emptyFaviconHandler)
	router.HandleFunc("/torrents", torrentsHandler(database))
	router.HandleFunc("/torrents/{infohash:[a-f0-9]{40}}", torrentsInfohashHandler)

	templateFunctions := template.FuncMap{
		"add": func(augend int, addends int) int {
			return augend + addends
		},

		"subtract": func(minuend int, subtrahend int) int {
			return minuend - subtrahend
		},

		"bytesToHex": hex.EncodeToString,

		"unixTimeToYearMonthDay": func(s int64) string {
			tm := time.Unix(s, 0)
			// > Format and Parse use example-based layouts. Usually you’ll use a constant from time
			// > for these layouts, but you can also supply custom layouts. Layouts must use the
			// > reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to
			// > format/parse a given time/string. The example time must be exactly as shown: the
			// > year 2006, 15 for the hour, Monday for the day of the week, etc.
			// https://gobyexample.com/time-formatting-parsing
			// Why you gotta be so weird Go?
			return tm.Format("02/01/2006")
		},

		"humanizeSize": humanize.IBytes,

		"humanizeSizeF": func(s int64) string {
			if s < 0 {
				return ""
			}
			return humanize.IBytes(uint64(s))
		},

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
