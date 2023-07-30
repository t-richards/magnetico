package serve

import (
	"strings"

	"github.com/t-richards/magnetico/internal/persistence"
)

type Directory struct {
	Name           string
	Files          map[string]persistence.File
	Subdirectories map[string]Directory
}

func makeTree(flatFiles []persistence.File) Directory {
	root := Directory{
		Name:           "/",
		Files:          make(map[string]persistence.File),
		Subdirectories: make(map[string]Directory),
	}

	for _, file := range flatFiles {
		parts := strings.Split(file.Path, "/")
		currentDirectory := root

		for i, part := range parts {
			if i == len(parts)-1 { // This is a file.
				currentDirectory.Files[part] = file
			} else { // This is a directory.
				if _, ok := currentDirectory.Subdirectories[part]; !ok {
					currentDirectory.Subdirectories[part] = Directory{
						Name:           part,
						Files:          make(map[string]persistence.File),
						Subdirectories: make(map[string]Directory),
					}
				}
				currentDirectory = currentDirectory.Subdirectories[part]
			}
		}
	}

	return root
}
