package assets

import (
	"testing"

	"github.com/influx6/flux"
)

func TestDebugBindFS(t *testing.T) {
	bf, err := NewBindFS(BindFSConfig{
		InDir:   "./",
		Dir:     "./tests/debug",
		Package: "debug",
		File:    "debug",
		Gzipped: false,
	})

	if err != nil {
		flux.FatalFailed(t, "Unable to create bindfs for: %s", err)
	}

	err = bf.Record()

	if err != nil {
		flux.FatalFailed(t, "Bindfs finished with err: %s", err)
	}
}

func TestProductionBindFS(t *testing.T) {
	bf, err := NewBindFS(BindFSConfig{
		InDir:      "./",
		Dir:        "./tests/prod",
		Package:    "prod",
		File:       "prod",
		Gzipped:    true,
		Production: true,
	})

	if err != nil {
		flux.FatalFailed(t, "Unable to create bindfs for: %s", err)
	}

	err = bf.Record()

	if err != nil {
		flux.FatalFailed(t, "Bindfs finished with err: %s", err)
	}
}
