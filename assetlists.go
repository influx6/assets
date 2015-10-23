package assets

import (
	"os"
	"path/filepath"
	"sync"
)

// PathValidators provides a type for validating a path and info set
type PathValidators func(string, os.FileInfo) bool

//PathMux provides a function type for mixing a new file path
type PathMux func(string, os.FileInfo) string

func defaultValidator(_ string, _ os.FileInfo) bool {
	return true
}

func defaultMux(n string, _ os.FileInfo) string {
	return n
}

// BasicAssetTree represent a directory structure and its corresponding assets
type BasicAssetTree struct {
	Dir      string
	AbsDir   string
	Info     os.FileInfo
	Tree     *MapWriter
	Ml       sync.RWMutex
	Children []*BasicAssetTree
}

// Add adds a BasicAssetTree into the children lists
func (b *BasicAssetTree) Add(bs *BasicAssetTree) {
	b.Ml.Lock()
	defer b.Ml.Lock()
	b.Children = append(b.Children, bs)
}

// EmptyAssetTree returns a new AssetTree based of the given path
func EmptyAssetTree(path string, info os.FileInfo, abs string) *BasicAssetTree {
	as := BasicAssetTree{
		Dir:      path,
		AbsDir:   abs,
		Info:     info,
		Tree:     NewMapWriter(make(AssetMap)),
		Children: make([]*BasicAssetTree, 0),
	}

	return &as
}

// BuildAssetPath reloads the files into the map skipping the already found ones
func BuildAssetPath(ws *sync.WaitGroup, base string, files []os.FileInfo, dirs *TreeMapWriter, pathtree *BasicAssetTree, validator PathValidators, mux PathMux) error {
	ws.Add(1)
	defer ws.Done()

	for _, pitem := range files {

		//get the file path using the provided base
		dir := filepath.Join(base, pitem.Name())

		if !validator(dir, pitem) {
			continue
		}

		// create a BasicAssetTree for this path

		if !pitem.IsDir() {
			pathtree.Tree.Add(mux(dir, pitem), dir)
			// pathtree.Tree[dir] = mux(dir,pitem)
			continue
		}

		var target *BasicAssetTree
		var muxdir = mux(dir, pitem)

		if dirs.Has(muxdir) {
			target = dirs.Get(muxdir)
		} else {
			tabsDir, _ := filepath.Abs(dir)
			target := EmptyAssetTree(dir, pitem, tabsDir)

			//add into the global dir listings
			dirs.Add(muxdir, target)

			//add to parenttree as a root dir
			pathtree.Add(target)
		}

		// var directories []os.FileInfo

		//open up the filepath since its a directory, read and sort
		dfiles, err := getDirListings(dir)

		if err != nil {
			//TODO: should we continue or fatal return here?
			return err
			// continue
		}

		//do another build but send into go-routine
		go BuildAssetPath(ws, dir, dfiles, dirs, target, validator, mux)
	}

	return nil
}

// LoadTree loads the path into the assettree updating any found trees along
func LoadTree(dir string, tree *TreeMapWriter, fx PathValidators, fxm PathMux) (*TreeMapWriter, error) {
	var st os.FileInfo
	var err error

	if st, err = os.Stat(dir); err != nil {
		return nil, err
	}

	// is this validate according to the current function validator
	if !fx(dir, st) {
		return tree, nil
	}

	//create the assettree for this path parent
	cdir, _ := filepath.Split(dir)

	//get roots stat
	cstat, err := os.Stat(cdir)

	if err != nil {
		return tree, err
	}

	abs, _ := filepath.Abs(cdir)

	var parentTree *BasicAssetTree

	if tree.Has(cdir) {
		parentTree = tree.Get(cdir)
	} else {
		parentTree = EmptyAssetTree(cdir, cstat, abs)

		//register root into tree
		tree.Add(cdir, parentTree)
		// tree[cdir] = parentTree
	}

	if !st.IsDir() {

		//its a file so we just register it into the parent's tree
		parentTree.Tree.Add(fxm(dir, st), dir)
		// tree[fxm(cdir, cstat)] = cur

	} else {

		files, err := getDirListings(dir)

		//unable to retrieve directory list, return tree and error
		if err != nil {
			return tree, err
		}

		//get the absolute path for the path
		absdir, _ := filepath.Abs(dir)

		var cur *BasicAssetTree
		var muxcur = fxm(dir, st)

		if tree.Has(muxcur) {
			cur = tree.Get(dir)
		} else {
			//create the assettree for this path
			cur = EmptyAssetTree(dir, st, absdir)

			//register and mux the path as requried to the super tree as a directory tree
			tree.Add(muxcur, cur)

			// tree[fxm(dir, st)] = cur
			//register into the parent tree as  child tree
			parentTree.Add(cur)
		}

		//since we will be using some go-routines to improve performance,lets create a waitgroup
		var wsg = new(sync.WaitGroup)

		//ok lets build the path children
		if err := BuildAssetPath(wsg, dir, files, tree, cur, fx, fxm); err != nil {
			return nil, err
		}

		wsg.Wait()
	}

	return tree, nil
}

// Assets returns a tree map listing of a specific path and if an error occured before the map was created it will return a nil and an error but if the path was valid but its children were faced with an invalid state then it returns the map and an error
func Assets(dir string, fx PathValidators, fxm PathMux) (*TreeMapWriter, error) {
	// use defaultValidator if non is set
	if fx == nil {
		fx = defaultValidator
	}

	// use defaultMux if non is set
	if fxm == nil {
		fxm = defaultMux
	}

	var tmap = make(AssetTreeMap)
	var tree = NewTreeMapWriter(tmap)

	return LoadTree(dir, tree, fx, fxm)
}
