package persistence

type OrderingCriteria uint8

const (
	ByRelevance OrderingCriteria = iota
	ByName
	ByTotalSize
	ByDiscovered
	ByNFiles
	ByUpdatedOn
)

type File struct {
	Size int64  `json:"size"`
	Path string `json:"path"`
}

type TorrentMetadata struct {
	ID        uint64  `json:"id"`
	InfoHash  []byte  `json:"infoHash"` // marshalled differently
	Name      string  `json:"name"`
	Size      uint64  `json:"size"`
	CreatedAt int64   `json:"createdAt"`
	UpdatedAt int64   `json:"updatedAt"`
	NFiles    uint    `json:"nFiles"`
	Relevance float64 `json:"relevance"`
}
