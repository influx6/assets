package assets

import "sync"

// AssetTreeMap represent a map of paths and asset maps
type AssetTreeMap map[string]*BasicAssetTree

// TreeMapWriter provides a safe writer for AssetTreeMap
type TreeMapWriter struct {
	Tree AssetTreeMap
	Wo   sync.RWMutex
}

// NewTreeMapWriter returns a new tree map writer for a maptree
func NewTreeMapWriter(c AssetTreeMap) *TreeMapWriter {
	ts := TreeMapWriter{Tree: c}
	return &ts
}

// Has returns a true/false if a tree with the set string exists
func (t *TreeMapWriter) Has(c string) bool {
	t.Wo.RLock()
	defer t.Wo.RUnlock()
	_, ok := t.Tree[c]
	return ok
}

// Delete removes a tree with the set string
func (t *TreeMapWriter) Delete(c string) {
	t.Wo.Lock()
	delete(t.Tree, c)
	t.Wo.Unlock()
}

// Add adds a new tree with the set string
func (t *TreeMapWriter) Add(c string, b *BasicAssetTree) {
	if t.Has(c) {
		return
	}
	t.Wo.Lock()
	t.Tree[c] = b
	t.Wo.Unlock()
}

// Flush nils out the former lists
func (t *TreeMapWriter) Flush() {
	t.Wo.Lock()
	t.Tree = make(AssetTreeMap)
	t.Wo.Unlock()
}

// Size returns the total length of items in the map
func (t *TreeMapWriter) Size() int {
	t.Wo.RLock()
	defer t.Wo.RUnlock()
	return len(t.Tree)
}

// Get retrieves a tree with the set string
func (t *TreeMapWriter) Get(c string) *BasicAssetTree {
	t.Wo.RLock()
	defer t.Wo.RUnlock()
	return t.Tree[c]
}

// MapWriter provides a safe writer for AssetMap
type MapWriter struct {
	Tree AssetMap
	Wo   sync.RWMutex
}

// NewMapWriter returns a new tree map writer for a maptree
func NewMapWriter(c AssetMap) *MapWriter {
	ts := MapWriter{Tree: c}
	return &ts
}

// Flush nils out the former lists
func (t *MapWriter) Flush() {
	t.Wo.Lock()
	t.Tree = make(AssetMap)
	t.Wo.Unlock()
}

// Has returns a true/false if a tree with the set string exists
func (t *MapWriter) Has(c string) bool {
	t.Wo.RLock()
	defer t.Wo.RUnlock()
	_, ok := t.Tree[c]
	return ok
}

// Delete removes a tree with the set string
func (t *MapWriter) Delete(c string) {
	t.Wo.Lock()
	delete(t.Tree, c)
	t.Wo.Unlock()
}

// Add adds a new tree with the set string
func (t *MapWriter) Add(c, b string) {
	if t.Has(c) {
		return
	}
	t.Wo.Lock()
	t.Tree[c] = b
	t.Wo.Unlock()
}

// Size returns the total length of items in the map
func (t *MapWriter) Size() int {
	t.Wo.RLock()
	defer t.Wo.RUnlock()
	return len(t.Tree)
}

// Get retrieves a tree with the set string
func (t *MapWriter) Get(c string) string {
	t.Wo.RLock()
	defer t.Wo.RUnlock()
	return t.Tree[c]
}
