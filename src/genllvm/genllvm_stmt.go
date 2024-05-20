package genllvm

import (
	"fmt"
	"strings"
	"strconv"
	"trivil/ast"
	"trivil/env"
)

var _ = fmt.Printf

func (genc *genContext) genStatementSeq(seq *ast.StatementSeq) {

	for _, s := range seq.Statements {
		genc.genStatement(s)
	}
}

func (genc *genContext) genStatement(s ast.Statement) {
	switch x := s.(type) {
	case *ast.DeclStatement:
		s := genc.genLocalDecl(x.D)
		genc.c("%s", s)
	case *ast.ExprStatement:
		s := genc.genExpr(x.X)
		genc.c("%s;", s)
	case *ast.AssignStatement:
		// CHANGE HERE: why tf is there x.L????
		l := genc.genExpr(x.L)
		r := genc.genExpr(x.R)

		var cast = genc.assignCast(x.L.GetType(), x.R.GetType())
		genc.c("%s = %s%s;", l, cast, r)
	case *ast.IncStatement:
		l := genc.genExpr(x.L)
		genc.c("%s++;", l)
	case *ast.DecStatement:
		l := genc.genExpr(x.L)
		genc.c("%s--;", l)
	case *ast.If:  // DONE
		genc.genIf(x)
	case *ast.While: // DONE
		genc.genWhile(x)
	// case *ast.Cycle:
	// 	genc.genCycle(x)
	// case *ast.Guard:
		// genc.genGuard(x)
	case *ast.Select:
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

	if ast.UnderType(lt) != ast.UnderType(rt) {
		return "(" + genc.typeRef(lt) + ")"
	}
	return ""
}

func (genc *genContext) genReturn(x *ast.Return) {
	for (TopScope.ScopeType != FuncScope) {
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

func (genc *genContext) genBreak () {
	for (TopScope.ScopeType != LoopScope) {
		popScope()
	}

	var restRegister = TopScope.EndLabel
	genc.c("br label %%%s", restRegister)

	popScope()
}

func (genc *genContext) genIf(x *ast.If) {
	var condExpr = genc.genExpr(x.Cond)
	var exprTrue = genc.newRegister()
	var exprFalse = genc.newRegister()
	var restRegister = genc.newRegister()

	pushScope(IfScope, "", strconv.Itoa(restRegister))

    genc.c("br i1 %s, label %%%d, label %%%d", condExpr, exprTrue, exprFalse)
	

	genc.c("%d:", exprTrue)
	genc.genStatementSeq(x.Then)
	genc.c("br label %%%d", restRegister)

	genc.c("%d:", exprFalse)
	genc.genStatement(x.Else)
	genc.c("br label %%%d", restRegister)

	genc.c("%d:", restRegister)
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
	var condition = genc.newRegister()
	genc.c("br label %%%d", condition)
	genc.c("%d:", condition)

	
	var condExpr = genc.genExpr(x.Cond)
	var exprTrue = genc.newRegister()
	var exprFalse = genc.newRegister()

	pushScope(LoopScope, strconv.Itoa(condition), strconv.Itoa(exprFalse))

	genc.c("br i1 %s, label %%%d, label %%%d", condExpr, exprTrue, exprFalse)

	genc.c("%d:", exprTrue)
	genc.genStatementSeq(x.Seq)
	genc.c("br label %%%d", condition)

	genc.c("%d:", exprFalse)
	popScope()
}

func (genc *genContext) genCycle(x *ast.Cycle) {

	var index = ""
	if x.IndexVar != nil {
		index = genc.declName(x.IndexVar)
	} else {
		index = genc.localName("i")
	}

	if ast.IsVariadicType(x.Expr.GetType()) {
		// нельзя использовать временную переменную
		panic("не реализовано для вариативных")
	}

	var loc = genc.localName("")
	genc.c("%s %s = %s;", genc.typeRef(x.Expr.GetType()), loc, genc.genExpr(x.Expr))

	genc.c("for (%s %s = 0;%s < %s;%s++) {",
		predefinedTypeName(ast.Int64.Name),
		index,
		index,
		genc.genLen(loc, x.Expr.GetType()),
		index)

	if x.ElementVar != nil {
		genc.c("%s %s = %s;",
			genc.typeRef(x.ElementVar.Typ),
			genc.declName(x.ElementVar),
			genc.genForElementSet(x.Expr.GetType(), loc, index))
	}
	genc.genStatementSeq(x.Seq)
	genc.c("}")
}

func (genc *genContext) genForElementSet(arrayType ast.Type, array string, index string) string {
	switch xt := ast.UnderType(arrayType).(type) {
	case *ast.VectorType:
		return fmt.Sprintf("%s->body[%s]", array, index)
	case *ast.VariadicType:
		panic("ni")
	default:
		if xt == ast.String8 {
			return fmt.Sprintf("%s->body[%s]", array, index)
		}
		panic("assert")
	}

}

func (genc *genContext) genGuard(x *ast.Guard) {
	genc.c("if (!(%s)) {", removeExtraPars(genc.genExpr(x.Cond)))
	seq, ok := x.Else.(*ast.StatementSeq)
	if ok {
		genc.genStatementSeq(seq)
	} else {
		genc.genStatement(x.Else)
	}
	genc.c("}")
}

func (genc *genContext) genCrash(x *ast.Crash) {

	var expr string
	var li = literal(x.X)
	if li != nil {
		expr = "\"" + encodeLiteralString(li.StrVal) + "\""
	} else {
		expr = genc.genExpr(x.X) + "->body"
	}

	genc.c("%s((char *)%s,%s);", rt_crash, expr, genPos(x.Pos))
}

func genPos(pos int) string {
	src, line, col := env.SourcePos(pos)
	return fmt.Sprintf("\"%s/%s:%d:%d\"", src.OriginPath, src.FileName, line, col)
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
	genc.c("switch (%s) {", genc.genExpr(x.X))

	for _, c := range x.Cases {
		for _, e := range c.Exprs {
			genc.c("case %s: ", genc.genExpr(e))
		}
		genc.c("{") // for clang 15.0.7 on linux
		genc.genStatementSeq(c.Seq)
		genc.c("}")
		genc.c("break;")
	}

	if x.Else != nil {
		genc.c("default:{")
		genc.genStatementSeq(x.Else)
		genc.c("}");
	}

	genc.c("}")
}

func (genc *genContext) genSelectAsIfs(x *ast.Select) {

	var strCompare = ast.IsStringType(x.X.GetType())

	var loc = genc.localName("")
	genc.c("%s %s = %s;", genc.typeRef(x.X.GetType()), loc, genc.genExpr(x.X))

	var els = ""
	for _, c := range x.Cases {

		var conds = make([]string, 0)
		for _, e := range c.Exprs {
			if strCompare {
				conds = append(conds, fmt.Sprintf("%s(%s, %s)", rt_equalStrings, loc, genc.genExpr(e)))
			} else {
				conds = append(conds, fmt.Sprintf("%s == %s", loc, genc.genExpr(e)))
			}
		}
		genc.c("%sif (%s) {", els, strings.Join(conds, " || "))
		els = "else "
		genc.genStatementSeq(c.Seq)
		genc.c("}")
	}

	if x.Else != nil {
		genc.c("else {")
		genc.genStatementSeq(x.Else)
		genc.c("}")
	}
}

func (genc *genContext) genPredicateSelect(x *ast.Select) {
	var restRegister = genc.newRegister()
	var exprFalse = 0
	for caseIdx, c := range x.Cases {
		var exprTrue = genc.newRegister()
		for idx, e := range c.Exprs {
			var condExpr = genc.genExpr(e)
			// to branch to either the statement sequence or next expression
			if (idx != 0) {
				genc.c("%d:", exprFalse)
			}
			exprFalse = genc.newRegister()
			genc.c("br i1 %%%s, label %%%d, label %%%d", condExpr, exprTrue, exprFalse)
		}
		genc.c("%d:", exprTrue)
		genc.genStatementSeq(c.Seq)
		genc.c("br label %%%d", restRegister)
		
		// to branch to the next case.
		if caseIdx != len(x.Cases) - 1 {
			genc.c("%d:", exprFalse)
		}
	}

	genc.c("%d:", exprFalse)
	if x.Else != nil {
		genc.genStatementSeq(x.Else)
		genc.c("br label %%%d", restRegister)
	} else {
		genc.c("br label %%%d", restRegister)
	}

	genc.c("%d:", restRegister)
}

//==== оператор выбора по типу

// if
func (genc *genContext) genSelectType(x *ast.SelectType) {

	var loc = genc.localName("")
	genc.c("%s %s = %s;", genc.typeRef(x.X.GetType()), loc, genc.genExpr(x.X))

	var tag = genc.localName("tag")
	genc.c("void* %s = %s->%s;", tag, loc, nm_VT_field)

	var els = ""
	for _, c := range x.Cases {

		var conds = make([]string, 0)
		for _, t := range c.Types {
			var tname = genc.typeRef(t)
			conds = append(conds, fmt.Sprintf("%s == %s", tag, tname+nm_class_info_ptr_suffix))
		}
		genc.c("%sif (%s) {", els, strings.Join(conds, " || "))
		els = "else "

		if c.Var != nil {
			var v = c.Var

			genc.c("%s %s = %s%s;",
				genc.typeRef(v.Typ),
				genc.declName(v),
				genc.assignCast(v.Typ, x.X.GetType()),
				loc)
		}
		genc.genStatementSeq(c.Seq)
		genc.c("}")
	}

	if x.Else != nil {
		genc.c("else {")
		genc.genStatementSeq(x.Else)
		genc.c("}")
	}

}
