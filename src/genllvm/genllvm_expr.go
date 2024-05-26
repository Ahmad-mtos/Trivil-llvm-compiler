package genllvm

import (
	"fmt"
	"strings"

	//"unicode"

	"trivil/ast"
	"trivil/lexer"
)

var _ = fmt.Printf

func (genc *genContext) genExpr(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.IdentExpr: // NEED WORK DO
		return genc.genIdent(x)
	case *ast.LiteralExpr: // DONE
		return genc.genLiteral(x)
	case *ast.UnaryExpr: // DONE
		return genc.genUnaryOp(x)
	case *ast.BinaryExpr: // DONE
		return genc.genBinaryExpr(x)
	// case *ast.OfTypeExpr:
	// 	return genc.genOfTypeExpr(x)
	case *ast.SelectorExpr:
		return genc.genSelectorExpr(x)
	case *ast.CallExpr:
		return genc.genCall(x)
	// case *ast.ConversionExpr:
	// 	if x.Caution {
	// 		return genc.genCautionCast(x)
	// 	} else {
	// 		return genc.genConversion(x)
	// 	}
	// case *ast.NotNilExpr:
	// 	return genc.genNotNil(x)

	// case *ast.GeneralBracketExpr:
	// 	return genc.genBracketExpr(x)

	// case *ast.ClassCompositeExpr:
	// 	return genc.genClassComposite(x)

	default:
		panic(fmt.Sprintf("gen expression: la %T", expr))
	}
}

func (genc *genContext) genSelectorExpr(s *ast.SelectorExpr) string {
	name := s.Name
	findInScopes(name)
	return name
}

func (genc *genContext) genIdent(id *ast.IdentExpr) string {
	// In case of ident is var
	var identName = id.Name
	
	var data = findInScopes(identName)
	ret := ""
	
	if getLLVMType(id.Typ) == FunctionType {
		ret = data.RegisterNum
	} else {
		var register = genc.newRegister()
		genc.c("%%%d = load %s, %s* %s", register, data.Typ, data.Typ, data.RegisterNum) // TODO: add align here
		ret = fmt.Sprintf("%%%d", register)
	}
	return ret
}

func (genc *genContext) genLiteral(li *ast.LiteralExpr) string {
	switch li.Kind {
	case ast.Lit_Int:
		return fmt.Sprintf("%d", li.IntVal)
	case ast.Lit_Word:
		return fmt.Sprintf("u0x%X", li.WordVal)
	case ast.Lit_Float:
		return li.FloatStr
	case ast.Lit_Symbol:
		return fmt.Sprintf("u0x%X", li.WordVal)
	// case ast.Lit_String:
	// 	return genc.genStringLiteral(li)
	default:
		panic("ni")
	}
}

func (genc *genContext) genUnaryOp(x *ast.UnaryExpr) string {

	var X = genc.genExpr(x.X)
	var result = genc.newRegister()
	var typ = x.X.GetType()

	switch x.Op {
	case lexer.NOT:
		switch{
		case ast.IsBoolType(typ):
			var temp1 = genc.newRegister()
			var temp2 = genc.newRegister()
			genc.c("%%%d = trunc i8 %s to i1", temp1, X)
			genc.c("%%%d = xor i1 %%%d, true", temp2, temp1)
			genc.c("%%%d = zext i1 %%%d to i8", result, temp2)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}
	case lexer.BITNOT:
		switch{
		case ast.IsInt64(typ):
			genc.c("%%%d = xor i64 %s, -1", result, X)
		case ast.IsWord64(typ):
			genc.c("%%%d = xor i64 %s, -1", result, X)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}
	case lexer.SUB:
		switch{
		case ast.IsInt64(typ):
			genc.c("%%%d = sub nsw i64 0, %s", result, X)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fneg double %s", result, X)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}
	}

	return fmt.Sprintf("%%%d", result)
}

func (genc *genContext) genBinaryExpr(x *ast.BinaryExpr) string {
	// In case x, y are from basic constructs. Need str, type

	var X = genc.genExpr(x.X)
	var Y = genc.genExpr(x.Y)
	var result = genc.newRegister()
	var typ = x.X.GetType()

	switch x.Op {
	case lexer.ADD:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = add nsw i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = add i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fadd double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.SUB:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = sub nsw i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = sub i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fsub double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.MUL:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = mul nsw i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = mul i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fmul double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.QUO:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = sdiv i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = udiv i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fdiv double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.REM:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = srem i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = urem i64 %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.BITAND:
		genc.c("%%%d = and i64 %s, %s", result, X, Y)

	case lexer.BITOR:
		genc.c("%%%d = or i64 %s, %s", result, X, Y)

	case lexer.BITXOR:
		genc.c("%%%d = xor i64 %s, %s", result, X, Y)

	case lexer.SHL:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = shl nsw i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = shl nuw i64 %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.SHR:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = ashr i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = lshr i64 %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.EQ:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = icmp eq i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fcmp eq double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.LSS:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = icmp slt i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = icmp ult i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fcmp ult double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.GTR:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = icmp sgt i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = icmp ugt i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fcmp ugt double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.NEQ:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = icmp ne i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fcmp ne double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.LEQ:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = icmp sle i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = icmp ule i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fcmp ule double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	case lexer.GEQ:
		switch {
		case ast.IsInt64(typ):
			genc.c("%%%d = icmp sge i64 %s, %s", result, X, Y)
		case ast.IsWord64(typ):
			genc.c("%%%d = icmp uge i64 %s, %s", result, X, Y)
		case ast.IsFloatType(typ):
			genc.c("%%%d = fcmp uge double %s, %s", result, X, Y)
		default:
			panic(fmt.Sprintf("Type not applicable: %s", typ))
		}

	default:
		panic(fmt.Sprintf("gen expression: la %d", x.Op))
	}

	return fmt.Sprintf("%%%d",result)
}

func encodeLiteralString(runes []rune) string {
	return ""
}

func (genc *genContext) genCall(call *ast.CallExpr) string {

	// if call.StdFunc != nil {
	// 	return ""
	// 	//return genc.genStdFuncCall(call)
	// }

	var ft = call.X.GetType().(*ast.FuncType)
	var left = genc.genExpr(call.X)
	var cargs = genc.genArgs(call)
	var typ = getLLVMType(ft.ReturnTyp)
	var isSTD = false

	switch left {
	case "цел64":
		left = findInScopes(left).RegisterNum
		reg := genc.newRegister()
		genc.c("%%%d = getelementptr inbounds [4 x i8], [4 x i8]*  @.int, i32 0, i32 0",reg)
		cargs = fmt.Sprintf("i8* %%%d, %s", reg, cargs)
		isSTD = true
	case "вещ64":
		left = findInScopes(left).RegisterNum
		reg := genc.newRegister()
		genc.c("%%%d = getelementptr inbounds [5 x i8], [5 x i8]*  @.double, i32 0, i32 0",reg)
		cargs = fmt.Sprintf("i8* %%%d, %s", reg, cargs)
		isSTD = true
	}

	var register = genc.newRegister()
	var _call = fmt.Sprintf("call %s %s(%s)", typ, left, cargs)
	
	if typ == "void" || isSTD{
		return _call
	}

	genc.c("%%%d = %s", register, _call)


	return fmt.Sprintf("%%%d", register)
}

func (genc *genContext) genArgs(call *ast.CallExpr) string {

	var ft = call.X.GetType().(*ast.FuncType)

	var cargs = make([]string, len(ft.Params))
	var normLen = len(ft.Params)

	// var vPar = ast.VariadicParam(ft)
	// if vPar != nil {
	// 	normLen--
	// }

	// не вариативные параметры
	for i := 0; i < normLen; i++ {
		var p = ft.Params[i]
		var arg = call.Args[i]
		var expr = genc.genExpr(arg)
		var typ = getLLVMType(p.Typ)

		cargs[i] = fmt.Sprintf("%s %s", typ, expr)
	}

	// if vPar != nil {
	// 	var vTyp = vPar.Typ.(*ast.VariadicType)

	// 	if ast.IsTagPairType(vTyp.ElementTyp) {
	// 		cargs[normLen] = genc.genVariadicTaggedArgs(call, vPar, normLen)
	// 	} else {
	// 		cargs[normLen] = genc.genVariadicArgs(call, vPar, vTyp, normLen)
	// 	}
	// }
	return strings.Join(cargs, ", ")
}