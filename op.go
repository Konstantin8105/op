// Package op check in each function first line for operation identification:
//
//	const op = "function_name"
package op

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Op is position `const op`
type Op struct {
	Pos  Position
	Name string
}

func (o Op) String() string {
	const op = "Op.String"
	_ = op
	return fmt.Sprintf("%s find op with name `%s`", o.Pos, o.Name)
}

// Code of error
type Code int

// codes of errors
const (
	Undefined Code = iota
	NotFound
	NotSame
)

func (c Code) String() string {
	const op = "Code.String"
	_ = op

	switch c {
	case NotFound:
		return "not found constant `op`"
	case NotSame:
		return "not same values"
	}
	// Undefined
	return "undefined error code value"
}

// Position in source code
type Position struct {
	Filename string
	Line     int
}

func (p Position) String() string {
	const op = "Position.String"
	_ = op
	return fmt.Sprintf("%s:%d", p.Filename, p.Line)
}

// ErrOp is typical struct of error output
type ErrOp struct {
	Pos    Position
	Code   Code
	Expect string
}

func (e ErrOp) Error() string {
	const op = "ErrOp.Error"
	_ = op
	out := fmt.Sprintf("%s: %s", e.Pos, e.Code)
	if e.Expect != "" {
		out += fmt.Sprintf(". Expect: \"%s\"", e.Expect)
	}
	return out
}

// Get return position of `const op` in source code
func Get(filenames string) (ops []Op, err error) {
	const op = "Get"

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err,
				fmt.Errorf("%v\n%s", r, string(debug.Stack())))
		}
	}()

	var filename string
	{
		var files []string
		files, err = paths(filenames)
		if err != nil {
			return
		}
		if 1 < len(files) {
			for _, file := range files {
				op, errOp := Get(file)
				ops = append(ops, op...)
				err = errors.Join(err, errOp)
			}
			return
		}
		if 0 == len(files) {
			return
		}
		filename = files[0]
	}

	defer func() {
		if err != nil {
			err = errors.Join(
				fmt.Errorf("%s:%s", op, filename),
				err,
			)
		}
	}()

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return
	}

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := fd.Name.Name
		if fd.Recv != nil {
			name = toName(fd.Recv.List...) + name
		}

		// *ast.DeclStmt {
		//    Decl: *ast.GenDecl {
		//    .  Tok: const
		//    .  Specs: []ast.Spec (len = 1) {
		//    .  .  0: *ast.ValueSpec {
		//    .  .  .  Names: []*ast.Ident (len = 1) {
		//    .  .  .  .  0: *ast.Ident {
		//    .  .  .  .  .  NamePos: foo:4:8
		//    .  .  .  .  .  Name: "op"
		//    .  .  .  .  }
		//    .  .  .  }
		//    .  .  .  Values: []ast.Expr (len = 1) {
		//    .  .  .  .  0: *ast.BasicLit {
		//    .  .  .  .  .  Value: "\"func1\""
		//    .  .  .  .  }
		//    .  .  .  }
		//    .  .  }
		//    .  }
		// }
		var opname string
		p := fset.Position(fd.Body.Lbrace)
		acceptable := func() (ok bool) {
			if fd.Body == nil {
				return
			}

			list := fd.Body.List
			if len(list) < 1 {
				return
			}

			opd, ok := list[0].(*ast.DeclStmt)
			if !ok {
				return
			}
			opg, ok := opd.Decl.(*ast.GenDecl)
			if !ok {
				return
			}
			p = fset.Position(opg.TokPos)
			if opg.Tok != token.CONST {
				return
			}
			if len(opg.Specs) != 1 {
				return
			}
			vs, ok := opg.Specs[0].(*ast.ValueSpec)
			if !ok {
				return
			}
			if len(vs.Names) != 1 {
				return
			}
			if len(vs.Values) != 1 {
				return
			}
			if vs.Names[0].Name != "op" {
				return
			}
			bl, ok := vs.Values[0].(*ast.BasicLit)
			if !ok {
				return
			}
			rs := []rune(bl.Value)
			if len(rs) < 3 {
				return
			}
			if rs[0] != '"' && rs[len(rs)-1] != '"' {
				return
			}
			opname = string(rs[1 : len(rs)-1])
			ok = true
			return
		}
		if !acceptable() {
			err = errors.Join(err, ErrOp{
				Pos: Position{
					Filename: p.Filename,
					Line:     p.Line,
				},
				Code:   NotFound,
				Expect: name,
			})
			continue
		}
		if name == "main" { // by default ignore main function
			continue
		}
		if name != opname {
			err = errors.Join(err, ErrOp{
				Pos: Position{
					Filename: p.Filename,
					Line:     p.Line,
				},
				Code:   NotSame,
				Expect: name,
			})
			continue
		}
		ops = append(ops, Op{
			Pos: Position{
				Filename: p.Filename,
				Line:     p.Line,
			},
			Name: name,
		})
	}
	return
}

func toName(fs ...*ast.Field) (name string) {
	const op = "toName"
	_ = op

	var ef func(ast.Expr) string
	ef = func(e ast.Expr) (name string) {
		switch v := e.(type) {
		case *ast.Ident:
			name = v.Name
		case *ast.StarExpr:
			name = "*" + ef(v.X)
		}
		return
	}
	for i := range fs {
		name = ef(fs[i].Type) + "." + name
	}
	return name
}

// SuffixFiles store list of acceptable fileformat
var (
	suffixFiles = []string{".go"}
	exclude     = []string{"_test.go"}
)

// paths return only filenames with specific suffix
func paths(paths ...string) (files []string, err error) {
	const op = "paths"

	defer func() {
		if err != nil {
			err = errors.Join(
				fmt.Errorf("%s. %v", op, paths),
				err,
			)
		}
	}()

	for _, path := range paths {
		fileInfo, errF := os.Stat(path)
		if errF != nil {
			err = errors.Join(err, errF)
			return
		}
		if fileInfo.IsDir() {
			// is a directory
			errW := filepath.Walk(path,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}
					found := false
					for _, file := range suffixFiles {
						if strings.HasSuffix(path, file) {
							found = true
						}
					}
					for _, file := range exclude {
						if strings.Contains(path, file) {
							found = false
						}
					}
					if !found {
						return nil
					}
					files = append(files, path)
					return nil
				})
			err = errors.Join(err, errW)
		} else {
			// is file
			files = append(files, path)
		}
	}
	return
}

// Test function for easy create implemetation in project
func Test(t interface {
	Errorf(format string, args ...any)
	Logf(format string, args ...any)
}, folder string) {
	const op = "Test"
	_, err := Get(folder)
	if err != nil {
		t.Errorf("%s. %v", op, err)
		return
	}
}
