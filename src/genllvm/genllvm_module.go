package genllvm

import (
	"fmt"
	"strings"

	"trivil/ast"
)

var _ = fmt.Printf

func (genc *genContext) genModule(main bool) {

	//=== import
	// CHANGER HERE: SHOULD BE HANDELED IN BUILD COMMAND IG?
	/*
		for _, i := range genc.module.Imports {
			genc.h("#include \"%s\"", genc.declName(i.Mod)+".h")
		}
		genc.h("")
	*/
	//=== gen types

	// Prints to header, might delete
	// for _, d := range genc.module.Decls {
	// 	d, ok := d.(*ast.TypeDecl)
	// 	if ok {
	// 		genc.genTypeDecl(d)
	// 	}
	// }
	genc.genTypes = false

	//=== gen vars, consts
	for _, d := range genc.module.Decls {
		switch x := d.(type) {
		case *ast.ConstDecl:
			genc.genGlobalConst(x)
		case *ast.VarDecl:
			genc.genGlobalVar(x)
		}
	}

	//=== gen functions
	for _, d := range genc.module.Decls {
		f, ok := d.(*ast.Function)
		if ok {
			genc.genFunction(f)
		}
	}

	genc.genEntry(genc.module.Entry, main)
}

func (genc *genContext) genFunction(f *ast.Function) {

}

func (genc *genContext) returnType(ft *ast.FuncType) string {
	return ""
}

func (genc *genContext) params(ft *ast.FuncType) string {
	return ""
}

//=== глобальные константы и переменные

func (genc *genContext) genGlobalConst(x *ast.ConstDecl) {
	// TODO: Check function and solve expression evaluation

	var val = genc.genExpr(x.Value)
	var ptr = genc.newRegister()
	var typ = getLLVMType(x.Typ)

	genc.c("@%d = constant %s %s", ptr, typ, val)
	var data = SymbolData{ptr, typ}
	addToScope(x.GetName(), data)
}

/**

scope[ruski] = register, typ
register is a pointer to the memory
int x = 5
=>
scope[x] = 3
alloca %3
store %3, 5
______________________________

const int x = 5
=>
scope[x] = 3
alloca
*/
// Обработка кода: конст к = функ
func (genc *genContext) typeOfFunction(x ast.Expr) string {

	checkFunctionRef(x)

	var ft = ast.UnderType(x.GetType()).(*ast.FuncType)

	var name = genc.localName("FT")

	var ps = make([]string, len(ft.Params))

	for i, p := range ft.Params {
		if ast.IsVariadicType(p.Typ) {
			ps[i] = "TInt64"
			ps = append(ps, "void*")
		} else {
			ps[i] = genc.typeRef(p.Typ)
		}
	}

	genc.g("typedef %s (*%s)(%s);",
		genc.returnType(ft),
		name,
		strings.Join(ps, ", "))

	return name
}

func checkFunctionRef(expr ast.Expr) {

	switch x := expr.(type) {
	case *ast.IdentExpr:
		if _, ok := x.Obj.(*ast.Function); ok {
			return
		}
	case *ast.SelectorExpr:
		if _, ok := x.Obj.(*ast.Function); ok {
			return
		}
	}

	panic("assert - должна быть ссылка на функцию")
}

func initializeInPlace(t ast.Type) bool {

	t = ast.UnderType(t)
	switch t {
	case ast.Byte, ast.Int64, ast.Float64, ast.Bool, ast.Symbol:
		return true
	}
	return false
}

func (genc *genContext) genGlobalVar(x *ast.VarDecl) {
	// TODO: Solve evaluation of exprestion globally problem
	if x.Exported {
		panic("экспортированные глобалы - запретить или сделать")
	}
	if x.Later {
		panic("ni - 'позже' для глобалов")
	}

	var val = genc.genExpr(x.Init)
	var ptr = genc.newRegister()
	var typ = getLLVMType(x.Typ)

	genc.c("@%d = global %s %s", ptr, typ, val)
	var data = SymbolData{ptr, typ}
	addToScope(x.GetName(), data)
}

func getLLVMType(typ ast.Type) string {
	switch {
	case ast.IsBoolType(typ):
		return BooleanType
	case ast.IsInt64(typ):
		return IntegerType
	case ast.IsFloatType(typ):
		return FloatType
	case ast.IsWord64(typ):
		return WordType
	case ast.IsSymbol(typ):
		return SymbolType
	case ast.IsStringType(typ):
		return StringType
	default:
		panic("Error: LLVM type not implemented")
	}
}

func (genc *genContext) genLocalDecl(d ast.Decl) {
	// switch x := d.(type) {
	// case *ast.VarDecl:

	// 	return fmt.Sprintf("%s %s = %s%s;",
	// 		genc.typeRef(x.Typ),
	// 		genc.declName(x),
	// 		genc.assignCast(x.Typ, x.Init.GetType()),
	// 		genc.genExpr(x.Init))
	// default:
	// 	panic(fmt.Sprintf("genDecl: ni %T", d))
	// }
	var x = d.(*ast.VarDecl)
	var val = genc.genExpr(x.Init)
	var ptr = genc.newRegister()
	var typ = getLLVMType(x.Typ)

	genc.c("%%%d = alloca %s", ptr, typ)
	genc.c("store %s %s, %s* %%%d", typ, val, typ, ptr)

	var data = SymbolData{ptr, typ}

	addToScope(x.GetName(), data)
}

//==== вход - инициализация или головной

const (
	init_fn  = "init"
	init_var = "init_done"
)

func (genc *genContext) genEntry(entry *ast.EntryFn, main bool) {

	if main {
		genc.c("define dso_local i32 @main() #0 {")
	} else {
		// var init_header = fmt.Sprintf("void %s__%s()", genc.outname, init_fn)

		// genc.h("%s;", init_header)

		// genc.c("static TBool %s = false;", init_var)
		// genc.c("%s {", init_header)
		// genc.c("if (%s) return;", init_var)
		// genc.c("%s = true;", init_var)
	}

	// for _, i := range genc.module.Imports {
	// 	genc.c("%s__%s();", genc.declName(i.Mod), init_fn)
	// }

	// genc.code = append(genc.code, genc.init...)

	// genc.genInitGlobals()

	if entry != nil {
		genc.genStatementSeq(entry.Seq)
	}

	if main {
		genc.c("ret i32 0")
		genc.c("}")
	}
}
