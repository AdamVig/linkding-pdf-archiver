package linkding

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

var extensionMap = map[string]string{
	".pdf": "application/pdf",
}

var mimeTypes = slices.Compact(slices.Collect(maps.Values(extensionMap)))

func GetMimeType(fileName string) (mimeType string, err error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	mimeType, ok := extensionMap[ext]

	if !ok {
		err = fmt.Errorf("unknown MIME type for %s", fileName)
	}

	return
}

func IsKnownMimeType(mimeType string) bool {
	return slices.Contains(mimeTypes, strings.ToLower(mimeType))
}
