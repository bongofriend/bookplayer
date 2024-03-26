package metadataextractor_test

import (
	"testing"

	"github.com/bongofriend/bookplayer/backend/lib/processing/metadataextractor"
)

func TestNewMetadataExtractor(t *testing.T) {
	if _, err := metadataextractor.NewMetadataExtractor(); err != nil {
		t.Fatal(err)
	}
}
