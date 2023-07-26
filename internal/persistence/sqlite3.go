package persistence

import (
	"bytes"
	"database/sql"
	"embed" // Required to use embed.FS
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "modernc.org/sqlite" // Required to use sqlite3
)

//go:embed migrations/*.sql
var migrations embed.FS

type sqlite3Database struct {
	conn *sql.DB
}

func makeSqlite3Database(filename string) (Database, error) {
	db := new(sqlite3Database)

	var err error
	db.conn, err = sql.Open("sqlite", filename)
	if err != nil {
		return nil, errors.New("sql.Open " + err.Error())
	}

	// > Open may just validate its arguments without creating a connection to the database. To
	// > verify that the data source Name is valid, call Ping.
	// https://golang.org/pkg/database/sql/#Open
	if err = db.conn.Ping(); err != nil {
		return nil, errors.New("sql.DB.Ping " + err.Error())
	}

	if err := db.setupDatabase(); err != nil {
		return nil, errors.New("setupDatabase " + err.Error())
	}

	return db, nil
}

func (db *sqlite3Database) Engine() databaseEngine {
	return Sqlite3
}

func (db *sqlite3Database) DoesTorrentExist(infoHash []byte) (bool, error) {
	rows, err := db.conn.Query("SELECT 1 FROM torrents WHERE info_hash = ?;", infoHash)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// When we receive a single result row, it means the torrent is in the database.
	exists := rows.Next()
	if rows.Err() != nil {
		return false, err
	}

	return exists, nil
}

func (db *sqlite3Database) AddNewTorrent(infoHash []byte, name string, files []File) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return errors.New("conn.Begin " + err.Error())
	}
	// If everything goes as planned and no error occurs, we will commit the transaction before
	// returning from the function so the tx.Rollback() call will fail, trying to rollback a
	// committed transaction. BUT, if an error occurs, we'll get our transaction rollback'ed, which
	// is nice.
	defer tx.Rollback() // nolint:errcheck

	var totalSize uint64 = 0
	for _, file := range files {
		totalSize += uint64(file.Size)
	}

	// This is a workaround for a bug: the database will not accept total_size to be zero.
	if totalSize == 0 {
		return nil
	}

	// Although we check whether the torrent exists in the database before asking MetadataSink to
	// fetch its metadata, the torrent can also exists in the Sink before that:
	//
	// If the torrent is complete (i.e. its metadata) and if its waiting in the channel to be
	// received, a race condition arises when we query the database and seeing that it doesn't
	// exists there, add it to the sink.
	//
	// Do NOT try to be clever and attempt to use INSERT OR IGNORE INTO or INSERT OR REPLACE INTO
	// without understanding their consequences fully:
	//
	// https://www.sqlite.org/lang_conflict.html
	//
	//   INSERT OR IGNORE INTO
	//     INSERT OR IGNORE INTO will ignore:
	//       1. CHECK constraint violations
	//       2. UNIQUE or PRIMARY KEY constraint violations
	//       3. NOT NULL constraint violations
	//
	//     You would NOT want to ignore #1 and #2 as they are likely to indicate programmer errors.
	//     Instead of silently ignoring them, let the program err and investigate the causes.
	//
	//   INSERT OR REPLACE INTO
	//     INSERT OR REPLACE INTO will replace on:
	//       1. UNIQUE or PRIMARY KEY constraint violations (by "deleting pre-existing rows that are
	//          causing the constraint violation prior to inserting or updating the current row")
	//
	//     INSERT OR REPLACE INTO will abort on:
	//       2. CHECK constraint violations
	//       3. NOT NULL constraint violations (if "the column has no default value")
	//
	//     INSERT OR REPLACE INTO is definitely much closer to what you may want, but deleting
	//     pre-existing rows means that you might cause users loose data (such as seeder and leecher
	//     information, readme, and so on) at the expense of /your/ own laziness...
	if exist, err := db.DoesTorrentExist(infoHash); exist || err != nil {
		return err
	}

	res, err := tx.Exec(`
		INSERT INTO torrents (
			info_hash,
			name,
			total_size,
			discovered_on
		) VALUES (?, ?, ?, ?);
	`, infoHash, name, totalSize, time.Now().Unix())
	if err != nil {
		return errors.New("tx.Exec (INSERT OR REPLACE INTO torrents) " + err.Error())
	}

	var lastInsertId int64
	if lastInsertId, err = res.LastInsertId(); err != nil {
		return errors.New("sql.Result.LastInsertId " + err.Error())
	}

	// > last_insert_rowid()
	// >   The last_insert_rowid() function returns the ROWID of the last row insert from the
	// >   database connection which invoked the function. If no successful INSERTs into rowid
	// >   tables have ever occurred on the database connection, then last_insert_rowid() returns
	// >   zero.
	// https://www.sqlite.org/lang_corefunc.html#last_insert_rowid
	// https://www.sqlite.org/c3ref/last_insert_rowid.html
	//
	// Now, last_insert_rowid() should never return zero (or any negative values really) as we
	// insert into torrents and handle any errors accordingly right afterwards.
	if lastInsertId <= 0 {
		log.Panicf("last_insert_rowid() <= 0 (this should have never happened!). lastInsertId: %d", lastInsertId)
	}

	for _, file := range files {
		_, err = tx.Exec("INSERT INTO files (torrent_id, size, path) VALUES (?, ?, ?);",
			lastInsertId, file.Size, file.Path,
		)
		if err != nil {
			return errors.New("tx.Exec (INSERT INTO files) " + err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.New("tx.Commit " + err.Error())
	}

	return nil
}

func (db *sqlite3Database) Close() error {
	return db.conn.Close()
}

func (db *sqlite3Database) GetNumberOfTorrents() (uint, error) {
	// COUNT(1) is much more inefficient since it scans the whole table, so use MAX(ROWID).
	// Keep in mind that the value returned by GetNumberOfTorrents() might be an approximation.
	rows, err := db.conn.Query("SELECT MAX(ROWID) FROM torrents;")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, fmt.Errorf("no rows returned from `SELECT MAX(ROWID)`")
	}

	var n *uint
	if err = rows.Scan(&n); err != nil {
		return 0, err
	}

	// If the database is empty (i.e. 0 entries in 'torrents') then the query will return nil.
	if n == nil {
		return 0, nil
	} else {
		return *n, nil
	}
}

func (db *sqlite3Database) QueryTorrents(
	query string,
	epoch int64,
	orderBy OrderingCriteria,
	ascending bool,
	limit uint,
	lastOrderedValue *float64,
	lastID *uint64,
) ([]TorrentMetadata, error) {
	if query == "" && orderBy == ByRelevance {
		return nil, fmt.Errorf("torrents cannot be ordered by relevance when the query is empty")
	}
	if (lastOrderedValue == nil) != (lastID == nil) {
		return nil, fmt.Errorf("lastOrderedValue and lastID should be supplied together, if supplied")
	}

	doJoin := query != ""
	firstPage := lastID == nil

	// executeTemplate is used to prepare the SQL query, WITH PLACEHOLDERS FOR USER INPUT.
	sqlQuery := executeTemplate(`
		SELECT id 
			 , info_hash
			 , name
			 , total_size
			 , discovered_on
			 , (SELECT COUNT(*) FROM files WHERE torrents.id = files.torrent_id) AS n_files
	{{ if .DoJoin }}
			 , idx.rank
	{{ else }}
			 , 0
	{{ end }}
		FROM torrents
	{{ if .DoJoin }}
		INNER JOIN (
			SELECT rowid AS id
				 , bm25(torrents_idx) AS rank
			FROM torrents_idx
			WHERE torrents_idx MATCH ?
		) AS idx USING(id)
	{{ end }}
		WHERE     modified_on <= ?
	{{ if not .FirstPage }}
			  AND ( {{.OrderOn}}, id ) {{GTEorLTE .Ascending}} (?, ?) -- https://www.sqlite.org/rowvalue.html#row_value_comparisons
	{{ end }}
		ORDER BY {{.OrderOn}} {{AscOrDesc .Ascending}}, id {{AscOrDesc .Ascending}}
		LIMIT ?;	
	`, struct {
		DoJoin    bool
		FirstPage bool
		OrderOn   string
		Ascending bool
	}{
		DoJoin:    doJoin,
		FirstPage: firstPage,
		OrderOn:   orderOn(orderBy),
		Ascending: ascending,
	}, template.FuncMap{
		"GTEorLTE": func(ascending bool) string {
			if ascending {
				return ">"
			} else {
				return "<"
			}
		},
		"AscOrDesc": func(ascending bool) string {
			if ascending {
				return "ASC"
			} else {
				return "DESC"
			}
		},
	})

	// Prepare query
	queryArgs := make([]interface{}, 0)
	if doJoin {
		queryArgs = append(queryArgs, query)
	}
	queryArgs = append(queryArgs, epoch)
	if !firstPage {
		queryArgs = append(queryArgs, lastOrderedValue)
		queryArgs = append(queryArgs, lastID)
	}
	queryArgs = append(queryArgs, limit)

	rows, err := db.conn.Query(sqlQuery, queryArgs...)
	defer closeRows(rows)
	if err != nil {
		return nil, errors.New("query error " + err.Error())
	}

	torrents := make([]TorrentMetadata, 0)
	for rows.Next() {
		var torrent TorrentMetadata
		err = rows.Scan(
			&torrent.ID,
			&torrent.InfoHash,
			&torrent.Name,
			&torrent.Size,
			&torrent.DiscoveredOn,
			&torrent.NFiles,
			&torrent.Relevance,
		)
		if err != nil {
			return nil, err
		}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

func orderOn(orderBy OrderingCriteria) string {
	switch orderBy {
	case ByRelevance:
		return "idx.rank"

	case ByTotalSize:
		return "total_size"

	case ByDiscoveredOn:
		return "discovered_on"

	case ByNFiles:
		return "n_files"

	default:
		panic(fmt.Sprintf("unknown orderBy: %v", orderBy))
	}
}

func (db *sqlite3Database) GetTorrent(infoHash []byte) (*TorrentMetadata, error) {
	rows, err := db.conn.Query(`
		SELECT
			info_hash,
			name,
			total_size,
			discovered_on,
			(SELECT COUNT(*) FROM files WHERE torrent_id = torrents.id) AS n_files
		FROM torrents
		WHERE info_hash = ?`,
		infoHash,
	)
	defer closeRows(rows)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	var tm TorrentMetadata
	if err = rows.Scan(&tm.InfoHash, &tm.Name, &tm.Size, &tm.DiscoveredOn, &tm.NFiles); err != nil {
		return nil, err
	}

	return &tm, nil
}

func (db *sqlite3Database) GetFiles(infoHash []byte) ([]File, error) {
	rows, err := db.conn.Query(
		"SELECT size, path FROM files, torrents WHERE files.torrent_id = torrents.id AND torrents.info_hash = ?;",
		infoHash)
	defer closeRows(rows)
	if err != nil {
		return nil, err
	}

	var files []File
	for rows.Next() {
		var file File
		if err = rows.Scan(&file.Size, &file.Path); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (db *sqlite3Database) setupDatabase() error {
	// Enable Write-Ahead Logging for SQLite as "WAL provides more concurrency as readers do not
	// block writers and a writer does not block readers. Reading and writing can proceed
	// concurrently."
	// Caveats:
	//   * Might be unsupported by OSes other than Windows and UNIXes.
	//   * Does not work over a network filesystem.
	//   * Transactions that involve changes against multiple ATTACHed databases are not atomic
	//     across all databases as a set.
	// See: https://www.sqlite.org/wal.html
	//
	// Force SQLite to use disk, instead of memory, for all temporary files to reduce the memory
	// footprint.
	//
	// Enable foreign key constraints in SQLite which are crucial to prevent programmer errors on
	// our side.
	_, err := db.conn.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA temp_store=1;
		PRAGMA foreign_keys=ON;
		PRAGMA encoding='UTF-8';
	`)
	if err != nil {
		return errors.New("sql.DB.Exec (PRAGMAs) " + err.Error())
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return errors.New("sql.DB.Begin " + err.Error())
	}
	// If everything goes as planned and no error occurs, we will commit the transaction before
	// returning from the function so the tx.Rollback() call will fail, trying to rollback a
	// committed transaction. BUT, if an error occurs, we'll get our transaction rollback'ed, which
	// is nice.
	defer tx.Rollback() // nolint:errcheck

	// Get the user_version:
	rows, err := tx.Query("PRAGMA user_version;")
	if err != nil {
		return errors.New("sql.Tx.Query (user_version) " + err.Error())
	}
	defer rows.Close()

	// NOTE: The user_version starts at 0, so our first migration MUST start at 1 to be applied.
	var userVersion int
	if !rows.Next() {
		return fmt.Errorf("sql.Rows.Next (user_version): PRAGMA user_version did not return any rows")
	}
	if err = rows.Scan(&userVersion); err != nil {
		return errors.New("sql.Rows.Scan (user_version) " + err.Error())
	}

	// Given the user version, find all migrations greater than it and execute them.
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return errors.New("migrations.ReadDir " + err.Error())
	}

	for _, migration := range entries {
		migrateVerString := strings.Split(migration.Name(), "_")[0]
		migrateVersion, err := strconv.ParseInt(migrateVerString, 10, 32)
		if err != nil {
			return errors.New("strconv.ParseInt " + err.Error())
		}

		if int(migrateVersion) <= userVersion {
			continue
		}

		log.Printf("Applying migration %s", migration.Name())
		contents, err := migrations.ReadFile("migrations/" + migration.Name())
		if err != nil {
			return errors.New("fs.ReadFile " + err.Error())
		}

		_, err = tx.Exec(string(contents))
		if err != nil {
			return errors.New("sql.Tx.Exec " + err.Error())
		}

		// Update the user_version.
		_, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d;", migrateVersion))
		if err != nil {
			return errors.New("sql.Tx.Exec (PRAGMA user_version) " + err.Error())
		}
	}

	// The database likes to corrupt itself for some reason.
	// Suspect flaky migration things.
	_, err = tx.Exec("PRAGMA integrity_check;")
	if err != nil {
		return errors.New("sql.Tx.Exec (PRAGMA integrity_check) " + err.Error())
	}

	if err = tx.Commit(); err != nil {
		return errors.New("sql.Tx.Commit " + err.Error())
	}

	return nil
}

func executeTemplate(text string, data interface{}, funcs template.FuncMap) string {
	t := template.Must(template.New("anon").Funcs(funcs).Parse(text))

	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		panic(err.Error())
	}
	return buf.String()
}

func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		log.Printf("could not close row %v", err)
	}
}