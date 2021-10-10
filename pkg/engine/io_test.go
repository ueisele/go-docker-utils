package engine

import (
	"path/filepath"
	"testing"
)

func TestGlob(t *testing.T) {
	filename := "../../examples/server.properties.gotpl"
	t.Log(filepath.Ext(filename))
	filenames, err := FileGlobsToFileNames("../../examples/**")
	t.Log(len(filenames))
	t.Logf("%s -> %s", filenames, err)

}
