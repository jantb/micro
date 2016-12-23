package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	pkgIndex = make(map[string][]pkg)
	exports  = make(map[string][]string)
)

type pkg struct {
	importpath string // full pkg import path, e.g. "net/http"
	dir        string // absolute file path to pkg directory e.g. "/usr/lib/go/src/fmt"
	exports    []string
}

func getCodeComplete() {
	ctx := build.Default
	for _, path := range ctx.SrcDirs() {
		f, err := os.Open(path)
		if err != nil {
			log.Print(err)
			continue
		}
		children, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			log.Print(err)
			continue
		}
		for _, child := range children {
			if child.IsDir() {
				loadPkg(path, child.Name())
			}
		}
	}
	// Populate exports global.
	for psi, ps := range pkgIndex {
		for pi, p := range ps {
			e := loadExports(p.dir)
			if e != nil {
				pkgIndex[psi][pi].exports = e
			}
		}
	}

	// Construct source file.
	fmt.Println(pkgIndex["fmt"][0].dir)
	fmt.Println(pkgIndex["fmt"][0].exports)
}

var fset = token.NewFileSet()

func loadPkg(root, importpath string) {
	shortName := path.Base(importpath)
	if shortName == "testdata" {
		return
	}

	dir := filepath.Join(root, importpath)
	pkgIndex[shortName] = append(pkgIndex[shortName], pkg{
		importpath: importpath,
		dir:        dir,
	})

	pkgDir, err := os.Open(dir)
	if err != nil {
		return
	}
	children, err := pkgDir.Readdir(-1)
	pkgDir.Close()
	if err != nil {
		return
	}
	for _, child := range children {
		name := child.Name()
		if name == "" {
			continue
		}
		if c := name[0]; c == '.' || ('0' <= c && c <= '9') {
			continue
		}
		if child.IsDir() {
			loadPkg(root, filepath.Join(importpath, name))
		}
	}
}

func loadExports(dir string) []string {
	exports := make(map[string]bool)
	buildPkg, err := build.ImportDir(dir, 0)
	if err != nil {
		if strings.Contains(err.Error(), "no buildable Go source files in") {
			return nil
		}
		log.Printf("could not import %q: %v", dir, err)
		return nil
	}
	for _, file := range buildPkg.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(dir, file), nil, 0)
		if err != nil {
			log.Printf("could not parse %q: %v", file, err)
			continue
		}
		for name, object := range f.Scope.Objects {
			if ast.IsExported(name) {

				if object.Kind == ast.Fun {
					f := object.Decl.(*ast.FuncDecl)
					paramNames := []string{}
					for x, value := range f.Type.Params.List {
						for i, value := range value.Names {
							paramNames = append(paramNames, fmt.Sprintf("$%d_%s$", x+i, value.Name))
						}
					}
					name = fmt.Sprint(name + "(" + strings.Join(paramNames, ",") + ")")
				}
				exports[name] = true
			}
		}
	}
	keys := make([]string, 0, len(exports))
	for k := range exports {
		keys = append(keys, k)
	}
	return keys
}
