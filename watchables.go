package assets

import (
	"go/build"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-fsnotify/fsnotify"
	// "github.com/howeyc/fsnotify"
)

// WatcherConfig to be used with watcher
type WatcherConfig struct {
	Dir      string
	Ext      []string
	Skip     []string
	MaxRetry int
	ExtraPkg []string
	Filter   func(addable, under string) bool
}

// WatchMux defines a function type for watcher change notifications
type WatchMux func(error, *fsnotify.Event, *Watcher)

// Watcher provides a directory and file watcher that scans and notifiers on a file change
type Watcher struct {
	*WatcherConfig

	assets AssetMap
	ro     sync.Mutex
	up     bool
	stop   bool
	fx     WatchMux

	// watcher *fsnotify.Watcher
}

// NewWatch returns a new watcher set over a directory for a specific extension if the ext is not a "" empty string and if not empty skips specific paths
func NewWatch(c WatcherConfig, fx WatchMux) (*Watcher, error) {

	tree, err := AssetTree(c.Dir, c.Ext, c.Skip)

	if err != nil {
		return nil, err
	}

	ws := &Watcher{
		WatcherConfig: &c,
		assets:        tree,
		fx:            fx,
	}

	if len(c.ExtraPkg) > 0 {
		for _, e := range c.ExtraPkg {
			ws.loadPkg(e)
		}
	}

	return ws, nil
}

// Stop stops the watcher process
func (w *Watcher) Stop() {
	w.ro.Lock()
	defer w.ro.Unlock()
	w.stop = true
}

func (w *Watcher) loadPkg(pkgname string) {
	if w.assets.Has(pkgname) {
		return
	}

	pkg, err := build.Import(pkgname, "", 0)

	if err != nil {
		log.Printf("Import path watch error %s", err.Error())
		return
	}

	if pkg.Goroot {
		return
	}

	// _, file := filepath.Split(pkg.Dir)

	w.assets[pkgname] = pkg.Dir

	for _, imp := range pkg.Imports {
		w.loadPkg(imp)
	}
}

// Start begins the watcher process
func (w *Watcher) Start() {

	//creation of fsnotify.Watcher retry count
	var retry = 0

	w.ro.Lock()
	if w.stop {
		w.stop = false
	}
	w.ro.Unlock()

	// fmt.Printf("Loading asset tree of length: %d", len(w.assets))

	for {

		if w.up {
			ReloadAssetMap(w.assets, w.Dir, w.Ext, w.Skip)
			// if len(w.ExtraPkg) > 0 {
			// 	for _, e := range w.ExtraPkg {
			// 		w.loadPkg(e)
			// 	}
			// }
		}

		wo, err := fsnotify.NewWatcher()

		if err != nil {

			log.Printf("Failed to created fsnotify.Watcher: %s", err)
			if retry > w.MaxRetry {
				break
			}

			retry++
			continue
		}

		for n, path := range w.assets {
			path = EnsureSlash(path)

			// if we have a filter function, use it to give more control as what gets added else use normal strategy
			if w.Filter != nil {
				if w.Filter(path, w.Dir) {
					if err := wo.Add(path); err != nil {
						log.Printf("Error adding file to watchlist: %s -> error: %s", path, err)
						delete(w.assets, n)
					}
				}
			} else {
				if !strings.Contains(path, "bin") {
					if err := wo.Add(path); err != nil {
						log.Printf("Error adding file to watchlist: %s -> error: %s", path, err)
						delete(w.assets, n)
					}
				}
			}
		}

		select {
		case ev := <-wo.Events:
			w.fx(nil, &ev, w)
		case erx := <-wo.Errors:
			w.fx(erx, nil, w)
		}

		wo.Close()
		w.up = true
	}
}

// ForceNotify forces a change notification
func (w *Watcher) ForceNotify() {
	w.fx(nil, &(fsnotify.Event{}), w)
}

// EnsureSlash ensures that backslash in the paths are made forward slashes
func EnsureSlash(path string) string {
	// if strings.Contains(runtime.GOOS,"window") || strings.Contain(runtime.GOOS,"win") {
	if strings.Contains(path, `\`) {
		return filepath.ToSlash(path)
	}
	return path
	// }
}

// ReverseSlash ensures that backslash in the paths are made remain backslashes slashes
func ReverseSlash(path string) string {
	// if strings.Contains(runtime.GOOS,"window") || strings.Contain(runtime.GOOS,"win") {
	if strings.Contains(path, `/`) {
		return filepath.ToSlash(path)
	}
	return path
	// }
}
