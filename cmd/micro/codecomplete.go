package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	pkgIndex = make(map[string][]pkg)
)

type pkg struct {
	importpath string
	dir        string
	exports    []string
}

func GetCodeComplete(substring string) []string {
	if len(pkgIndex) == 0 {
		ReindexCodeComplete()
	}
	split := strings.Split(substring, ".")
	ret := []string{}
	if len(split) == 2 {
		for _, value := range pkgIndex[split[0]] {
			for _, value := range value.exports {
				if strings.Index(value, split[1]) > -1 {
					ret = append(ret, value)
				}
			}
		}
	} else if len(split) == 1 {
		for key, value := range pkgIndex {
			for _, value := range value {
				for _, value := range value.exports {
					if strings.Index(value, substring) > -1 {
						ret = append(ret, value[0:strings.Index(value, ",,")+2]+key+"."+value[strings.Index(value, ",,")+2:])
					}
				}
			}
		}
	} else {
		for key, value := range pkgIndex {
			if strings.Index(key, substring) > -1 {
				for _, value := range value {
					for _, value := range value.exports {
						ret = append(ret, value)
					}
				}
			}
		}
	}

	return ret
}

func ReindexCodeComplete() {
	ctx := build.Default
	for _, p := range ctx.SrcDirs() {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		children, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			continue
		}
		for _, child := range children {
			if child.IsDir() {
				loadPkg(p, child.Name())
			}
		}
	}
	for psi, ps := range pkgIndex {
		for pi, p := range ps {
			e := loadExports(p.dir)
			if e != nil {
				pkgIndex[psi][pi].exports = e
			}
		}
	}
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
		return nil
	}
	for _, file := range buildPkg.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(dir, file), nil, 0)
		if err != nil {
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
					name = fmt.Sprint("func,," + name + "(" + strings.Join(paramNames, ",") + ")" + ",,")
				} else if object.Kind == ast.Var {
					name = fmt.Sprint("var,," + name + ",,")
				} else if object.Kind == ast.Typ {
					name = fmt.Sprint("type,," + name + ",,")
				} else if object.Kind == ast.Con {
					name = fmt.Sprint("const,," + name + ",,")
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
