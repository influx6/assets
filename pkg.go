package assets

import "go/build"

func getPackageLists(pkgname string, paths []string) error {
	pkg, err := build.Import(pkgname, "", 0)

	if err != nil {
		return err
	}

	if pkg.Goroot {
		return nil
	}

	paths = append(paths, pkg.Dir)

	for _, imp := range pkg.Imports {
		getPackageLists(imp, paths)
	}

	return nil
}

// GetPackageLists retrieves a packages  directory and those of its dependencies
func GetPackageLists(pkgname string) ([]string, error) {
	var paths []string

	if err := getPackageLists(pkgname, paths); err != nil {
		return nil, err
	}

	return paths, nil
}
