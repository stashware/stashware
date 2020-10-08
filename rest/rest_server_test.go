package rest

import (
	"mime"
	"testing"
)

func TestMine(t *testing.T) {
	content_type := mime.TypeByExtension(".zip")
	t.Logf("%s", content_type)
}
