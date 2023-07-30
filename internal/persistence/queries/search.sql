SELECT id 
    , info_hash
    , name
    , total_size
    , created_at
    , updated_at
    , (SELECT COUNT(*) FROM files WHERE torrents.id = files.torrent_id) AS n_files
    , idx.rank

FROM torrents

INNER JOIN (
    SELECT rowid AS id
        , bm25(torrents_idx) AS rank
    FROM torrents_idx
    WHERE torrents_idx MATCH ?
) AS idx USING(id)

{{ if not .FirstPage }}
WHERE ( {{.OrderOn}}, id ) {{GTEorLTE .Ascending}} (?, ?) {{/* https://www.sqlite.org/rowvalue.html#row_value_comparisons */}}
{{ end }}
ORDER BY {{.OrderOn}} {{AscOrDesc .Ascending}}, id {{AscOrDesc .Ascending}}

LIMIT ?;
