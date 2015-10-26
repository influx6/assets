package assets

import "go/build"

//SanitizeDuplicates cleans out all duplicates
func SanitizeDuplicates(b []string) {
	sz := len(b) - 1
	for i := 0; i < sz; i++ {
		for j := i + 1; j <= sz; j++ {
			if (b)[i] == ((b)[j]) {
				(b)[j] = (b)[sz]
				(b) = (b)[0:sz]
				sz--
				j--
			}
		}
	}
}

func getPackageLists(pkgname string, paths []string) ([]string, error) {
	pkg, err := build.Import(pkgname, "", 0)

	// log.Printf("package found: %s", paths)
	if err != nil {
		return nil, err
	}

	if pkg.Goroot {
		return paths, nil
	}

	paths = append(paths, pkg.Dir)

	for _, imp := range pkg.Imports {
		if p, err := getPackageLists(imp, paths); err == nil {
			paths = p
		} else {
			return nil, err
		}
	}

	return paths, nil
}

// GetPackageLists retrieves a packages  directory and those of its dependencies
func GetPackageLists(pkgname string) ([]string, error) {
	var paths []string
	var err error

	if paths, err = getPackageLists(pkgname, paths); err != nil {
		return nil, err
	}

	SanitizeDuplicates(paths)

	return paths, nil
}

// GetAllPackageLists retrieves a set of packages directory and those of its dependencies
func GetAllPackageLists(pkgnames []string) ([]string, error) {
	var packages []string
	var err error

	for _, pkg := range pkgnames {
		if packages, err = getPackageLists(pkg, packages); err != nil {
			return nil, err
		}
	}

	// log.Printf("Packages: %s", packages)
	SanitizeDuplicates(packages)
	return packages, nil
}
