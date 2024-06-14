package genllvm

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"

	"trivil/ast"
	"trivil/env"
)

var _ = fmt.Printf

type genContext struct {
	module      *ast.Module
	outname     string
	declNames   map[ast.Decl]string
	genTypes    bool
	globalCnt 	int
	registerCnt int // used to know the current register number
	labelCnt	int // used to know the current label number
	autoNo      int // used for additional names
	header      []string
	code        []string
	globals     []string
	init        []string        // используется для типов
	initGlobals []ast.Decl      // глобалы, которые надо инициализировать во входе
	sysAPI      map[string]bool // true - exported
}

func Generate(m *ast.Module, main bool) {

	var genllvm = &genContext{
		module:      m,
		outname:     env.OutName(m.Name),
		declNames:   make(map[ast.Decl]string),
		globalCnt:   0,
		registerCnt: 0,
		labelCnt:	 0,
		header:      make([]string, 0),
		code:        make([]string, 0),
		globals:     make([]string, 0),
		init:        make([]string, 0),
		initGlobals: make([]ast.Decl, 0),
		sysAPI:      make(map[string]bool),
	}
	if TopScope == nil{
		pushScope(GlobalScope, "", "")
	}

	pushSTD()

	genllvm.genModule(main)
	genllvm.finishCode()
	if TopScope != nil{
		popScope()
	}

	//genc.show()
	genllvm.save()

}

func (genc *genContext) h(format string, args ...interface{}) {
	genc.header = append(genc.header, fmt.Sprintf(format, args...))
}

func (genc *genContext) c(format string, args ...interface{}) {
	if !isTerminated() {
		genc.code = append(genc.code, fmt.Sprintf(format, args...))
	}
}

func (genc *genContext) g(format string, args ...interface{}) {
	genc.globals = append(genc.globals, fmt.Sprintf(format, args...))
}

func (genc *genContext) newGlobal() int {
	var ret = genc.globalCnt
	genc.globalCnt++
	return ret
}

func (genc *genContext) newRegister() int {
	var ret = genc.registerCnt
	if !isTerminated() {
		genc.registerCnt++
	}
	return ret
}

func (genc *genContext) newLabel() string {
	var ret = genc.labelCnt
	genc.labelCnt++
	return "label_" + strconv.Itoa(ret)
}

func (genc *genContext) finishCode() {
	// code
	var lines = genc.code
	genc.code = make([]string, 0)

	genc.c("")

	if len(genc.globals) != 0 {
		genc.c("//--- globals")
		genc.code = append(genc.code, genc.globals...)
		genc.c("//--- end globals")
		genc.c("")
	}

	genc.code = append(genc.code, lines...)
}

func (genc *genContext) save() {
	var folder = env.PrepareOutFolder()

	writeFile(folder, genc.outname, ".ll", genc.code)
}

func writeFile(folder, name, ext string, lines []string) {
	writeFileCommon(folder, name, ext, lines, 0644)
}

func writeFileExecutable(folder, name, ext string, lines []string) {
    writeFileCommon(folder, name, ext, lines, 0755)
}

func writeFileCommon(folder, name, ext string, lines []string, perm fs.FileMode) {

	var filename = path.Join(folder, name+ext)

	var out = strings.Join(lines, "\n")

	var err = os.WriteFile(filename, []byte(out), perm)

	if err != nil {
		panic("Ошибка записи файла " + filename + ": " + err.Error())
	}
}

