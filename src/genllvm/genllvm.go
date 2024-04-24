package genllvm

import (
	"fmt"
	"io/fs"
	"os"
	"path"
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
	registerCnt int // used to know the current register number
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
		registerCnt: 10,
		header:      make([]string, 0),
		code:        make([]string, 0),
		globals:     make([]string, 0),
		init:        make([]string, 0),
		initGlobals: make([]ast.Decl, 0),
		sysAPI:      make(map[string]bool),
	}
	pushScope(GlobalScope, "", "")
	genllvm.genModule(main)
	genllvm.finishCode()

	//genc.show()
	genllvm.save()

}


func (genc *genContext) h(format string, args ...interface{}) {
	genc.header = append(genc.header, fmt.Sprintf(format, args...))
}

func (genc *genContext) c(format string, args ...interface{}) {
	genc.code = append(genc.code, fmt.Sprintf(format, args...))
}

func (genc *genContext) g(format string, args ...interface{}) {
	genc.globals = append(genc.globals, fmt.Sprintf(format, args...))
}

func (genc *genContext) finishCode() {
	var hname = fmt.Sprintf("_%s_H", genc.outname)

	// header file
	var lines = genc.header

	genc.header = make([]string, 0)
	genc.h("#ifndef %s", hname)
	genc.h("#define %s", hname)

	// CHANGE HERE: COMMENTED OUT
	//// genc.includeSysAPI(genc.header, true)
	genc.h("")

	genc.header = append(genc.header, lines...)

	genc.h("#endif")

	// code
	lines = genc.code
	genc.code = make([]string, 0)

	genc.c("#include \"rt_api.h\"")
	genc.c("#include \"%s\"", genc.outname+".h")
	// CHANGER HERE: COMMENTED OUT
	///genc.includeSysAPI(genc.code, false)
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

func writeFileCommon(folder, name, ext string, lines []string, perm fs.FileMode) {

	var filename = path.Join(folder, name+ext)

	var out = strings.Join(lines, "\n")

	var err = os.WriteFile(filename, []byte(out), perm)

	if err != nil {
		panic("Ошибка записи файла " + filename + ": " + err.Error())
	}
}