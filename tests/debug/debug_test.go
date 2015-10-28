package debug

import (
	"testing"

	"github.com/influx6/flux"
)

func TestVirtualDir(t *testing.T) {
	if _, err := RootDirectory.GetDir("/"); err != nil {
		flux.FatalFailed(t, "Unable to located asset dir in dirCollection")
	}

	if _, err := RootDirectory.GetDir("/fixtures"); err != nil {
		flux.FatalFailed(t, "Unable to located asset/tests dir in dirCollection")
	}

	to, err := RootDirectory.GetDir("/fixtures/base")

	if err != nil {
		flux.FatalFailed(t, "Unable to located asset/tests directory: %s", err)
	}

	al, err := to.GetFile("basic.tmpl")

	if err != nil {
		flux.FatalFailed(t, "Unable to located /fixtures/base/basic.tmpl file: %s", err)
	}

	if data, err := al.Data(); err != nil {
		flux.FatalFailed(t, "failed to load /fixtures/base/basic.tmpl contents: %s", err)
	} else if len(data) != 364 {
		flux.FatalFailed(t, "incorrect assets/tests/lock.md content, expected length %d got %d", 364, len(data))
	}

}
