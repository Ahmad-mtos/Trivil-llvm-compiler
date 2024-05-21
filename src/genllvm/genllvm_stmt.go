package genllvm

import (
	"fmt"
	// "strings"
	"strconv"
	"trivil/ast"
	// "trivil/env"
)

var _ = fmt.Printf

func (genc *genContext) genStatementSeq(seq *ast.StatementSeq) {

	for _, s := range seq.Statements {
		genc.genStatement(s)
	}
}

func (genc *genContext) genStatement(s ast.Statement) {
	fmt.Println(s)
	switch x := s.(type) {
	case *ast.DeclStatement: // DONE
		genc.genLocalDecl(x.D)
	// case *ast.ExprStatement:
	// 	s := genc.genExpr(x.X)
	// 	genc.c("%s;", s)
	case *ast.AssignStatement: // DONE
		genc.genAssignStatement(x)
	case *ast.IncStatement: // DONE
		genc.genIncStatement(x)
	case *ast.DecStatement: // DONE
		genc.genDecStatement(x)
	case *ast.If: // DONE
		genc.genIf(x)
	case *ast.While: // DONE
		genc.genWhile(x)
	// case *ast.Cycle:
	// 	genc.genCycle(x)
	// case *ast.Guard:
	// genc.genGuard(x)
	case *ast.Select: // DONE?
		if canSelectAsSwitch(x) {
			//genc.genSelectAsSwitch(x)
		} else if x.X == nil {
			genc.genPredicateSelect(x)
		} else {
			//genc.genSelectAsIfs(x)
		}
	// case *ast.SelectType:
	// genc.genSelectType(x)
	case *ast.Return: // DONE
		genc.genReturn(x)
	case *ast.Break: // DONE
		genc.genBreak()
	// case *ast.Crash:
	// genc.genCrash(x)

	default:
		panic(fmt.Sprintf("gen statement: ni %T", s))
	}
}

func (genc *genContext) assignCast(lt, rt ast.Type) string {
	return ""
}

func (genc *genContext) genReturn(x *ast.Return) {
	for TopScope.ScopeType != FuncScope {
		popScope()
	}

	if x.X != nil {
		var xType = x.X.GetType()
		var expr = genc.genExpr(x.X)
		switch {
		case ast.IsInt64(xType), ast.IsWord64(xType):
			genc.c("ret i64 %s", expr)
		case ast.IsFloatType(xType):
			genc.c("ret double %s", expr)
		case ast.IsBoolType(xType):
			genc.c("ret i8 %s", expr)
		default:
			panic(fmt.Sprintf("genReturn: type not supported %d", 10200))
		}
	} else {
		genc.c("ret i32 0")
	}

	popScope()
}

func (genc *genContext) genBreak() {
	for TopScope.ScopeType != LoopScope {
		popScope()
	}

	var restRegister = TopScope.EndLabel
	genc.c("br label %%%s", restRegister)

	popScope()
}

func (genc *genContext) genAssignStatement(x *ast.AssignStatement) {
	data := findInScopes(x.L.(*ast.IdentExpr).Name)

	l := "%" + strconv.Itoa(data.RegisterNum)
	r := genc.genExpr(x.R)

	typ := getLLVMType(x.L.GetType())
	genc.c("store %s %s, %s* %s", typ, r, typ, l)
}

func (genc *genContext) genIncStatement(x *ast.IncStatement) {
	data := findInScopes(x.L.(*ast.IdentExpr).Name)

	l := "%" + strconv.Itoa(data.RegisterNum)
	//r := genc.genExpr(x.R)

	typ := getLLVMType(x.L.GetType())

	register1 := genc.newRegister()
	register2 := genc.newRegister()

	genc.c("%%%d = load %s, %s* %s", register1, typ, typ, l)
	switch {
	case typ == IntegerType:
		genc.c("%%%d = add nsw i64 %%%d, 1", register2, register1)
	case typ == WordType:
		genc.c("%%%d = add i64 %%%d, 1", register2, register1)
	case typ == FloatType:
		genc.c("%%%d = fadd double %%%d, 1", register2, register1)
	default:
		panic(fmt.Sprintf("Type not applicable: %s", typ))
	}

	genc.c("store %s %%%d, %s* %s", typ, register2, typ, l)
}

func (genc *genContext) genDecStatement(x *ast.DecStatement) {
	data := findInScopes(x.L.(*ast.IdentExpr).Name)

	l := "%" + strconv.Itoa(data.RegisterNum)
	//r := genc.genExpr(x.R)

	typ := getLLVMType(x.L.GetType())

	register1 := genc.newRegister()
	register2 := genc.newRegister()

	genc.c("%%%d = load %s, %s* %s", register1, typ, typ, l)
	switch {
	case typ == IntegerType:
		genc.c("%%%d = sub nsw i64 %%%d, 1", register2, register1)
	case typ == WordType:
		genc.c("%%%d = sub i64 %%%d, 1", register2, register1)
	case typ == FloatType:
		genc.c("%%%d = fsub double %%%d, 1", register2, register1)
	default:
		panic(fmt.Sprintf("Type not applicable: %s", typ))
	}

	genc.c("store %s %%%d, %s* %s", typ, register2, typ, l)
}

func (genc *genContext) genIf(x *ast.If) {
	var condExpr = genc.genExpr(x.Cond)
	var exprTrue = genc.newLabel()
	var exprFalse = genc.newLabel()
	var restRegister = genc.newLabel()

	pushScope(IfScope, "", restRegister)

	genc.c("br i1 %s, label %%%s, label %%%s", condExpr, exprTrue, exprFalse)

	genc.c("%s:", exprTrue)
	genc.genStatementSeq(x.Then)
	genc.c("br label %%%s", restRegister)

	genc.c("%s:", exprFalse)
	genc.genStatementSeq(x.Else.(*ast.StatementSeq))
	genc.c("br label %%%s", restRegister)

	genc.c("%s:", restRegister)
}

func removeExtraPars(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] == '(' && s[len(s)-1] == ')' {
		return s[1 : len(s)-1]
	}
	return s
}

func (genc *genContext) genWhile(x *ast.While) {
	var condition = genc.newLabel()
	genc.c("br label %%%s", condition)
	genc.c("%s:", condition)

	var condExpr = genc.genExpr(x.Cond)
	var exprTrue = genc.newLabel()
	var exprFalse = genc.newLabel()

	pushScope(LoopScope, condition, exprFalse)

	genc.c("br i1 %s, label %%%s, label %%%s", condExpr, exprTrue, exprFalse)

	genc.c("%s:", exprTrue)
	genc.genStatementSeq(x.Seq)
	genc.c("br label %%%s", condition)

	genc.c("%s:", exprFalse)
	popScope()
}

func (genc *genContext) genCycle(x *ast.Cycle) {

}

func (genc *genContext) genForElementSet(arrayType ast.Type, array string, index string) string {
	return ""
}

func (genc *genContext) genGuard(x *ast.Guard) {
}

func (genc *genContext) genCrash(x *ast.Crash) {

}

func genPos(pos int) string {
	return ""
}

func literal(expr ast.Expr) *ast.LiteralExpr {

	switch x := expr.(type) {
	case *ast.LiteralExpr:
		return x
	case *ast.ConversionExpr:
		if x.Done {
			return literal(x.X)
		}
	}
	return nil
}

//==== оператор выбор

func canSelectAsSwitch(x *ast.Select) bool {

	if x.X == nil {
		return false
	}

	var t = ast.UnderType(x.X.GetType())
	switch t {
	case ast.Byte, ast.Int64, ast.Word64, ast.Symbol:
	default:
		return false
	}

	for _, c := range x.Cases {
		for _, e := range c.Exprs {
			if _, ok := e.(*ast.LiteralExpr); !ok {
				return false
			}
		}
	}
	return true
}

func (genc *genContext) genSelectAsSwitch(x *ast.Select) {

}

func (genc *genContext) genSelectAsIfs(x *ast.Select) {

}

func (genc *genContext) genPredicateSelect(x *ast.Select) {
	var restRegister = genc.newLabel()
	var exprFalse = "0"
	for caseIdx, c := range x.Cases {
		var exprTrue = genc.newLabel()
		for idx, e := range c.Exprs {
			var condExpr = genc.genExpr(e)
			// to branch to either the statement sequence or next expression
			if idx != 0 {
				genc.c("%s:", exprFalse)
			}
			exprFalse = genc.newLabel()
			genc.c("br i1 %%%s, label %%%s, label %%%s", condExpr, exprTrue, exprFalse)
		}
		genc.c("%s:", exprTrue)
		genc.genStatementSeq(c.Seq)
		genc.c("br label %%%s", restRegister)

		// to branch to the next case.
		if caseIdx != len(x.Cases)-1 {
			genc.c("%s:", exprFalse)
		}
	}

	genc.c("%s:", exprFalse)
	if x.Else != nil {
		genc.genStatementSeq(x.Else)
		genc.c("br label %%%s", restRegister)
	} else {
		genc.c("br label %%%s", restRegister)
	}

	genc.c("%s:", restRegister)
}

//==== оператор выбора по типу

// if
func (genc *genContext) genSelectType(x *ast.SelectType) {

}
