package genllvm

import "fmt"

const (
    IntegerType = "i64"
    BooleanType = "i8"
    FloatType = "double"
	WordType = "i64"
	SymbolType = "i32"
	StringType = "type { i64, i64, i8* }"
	NullType = "null"
	VoidType = "void"
	FunctionType = "func"
)

const (
	GlobalScope = iota
	FuncScope = iota
	LoopScope = iota
	IfScope = iota
)

type SymbolData struct{
	RegisterNum string
	Typ string
}

type Scope struct {
	Outer *Scope
	Names map[string]SymbolData 
	ScopeType int
	StartLabel string
	EndLabel string
} 

var TopScope *Scope

func pushScope(scopeType int, startLabel string, endLabel string) {
	var newScope = &Scope{
		Outer: TopScope,
		Names: make(map[string]SymbolData),
		ScopeType:  scopeType,
		StartLabel: startLabel,
		EndLabel: endLabel,
	}
	TopScope = newScope
}

func pushSTD() {
	// var trueData = SymbolData{"true", BooleanType}
	// addToScope("истина", trueData)
	// var falseData = SymbolData{"false", BooleanType}
	// addToScope("ложь", falseData)

	data := SymbolData{"(i8*, ...) @printf",SymbolType}
	addToScope("цел64", data)
	data = SymbolData{"(i8*, ...) @printf",SymbolType}
	addToScope("вещ64", data)
}

func popScope() {
	if TopScope.Outer == nil{
		panic("Stack is empty.")
	}
	TopScope = TopScope.Outer
}

func addToScope(name string, data SymbolData) {
	TopScope.Names[name] = data
}

func printScopes() {
	var cur = TopScope
	for cur != nil {
		println(cur.ScopeType,":")
		for key,name := range cur.Names{
			println(key, name.RegisterNum, name.Typ)
		}
		cur = cur.Outer
	}
}

func findScopeWithType(scopeTyp int) *Scope {
	var cur = TopScope
	for {
		if cur == nil {
			panic(fmt.Sprint("%d نوع النطاق ليس موجود", scopeTyp))
		}
		if cur.ScopeType == scopeTyp {
			return cur
		}
		cur = cur.Outer
	}
}

func findEndLabel(scopeTyp int) string {
	var cur = findScopeWithType(scopeTyp)
	return cur.EndLabel
}

func findInScopes(name string) *SymbolData {
	var cur = TopScope

	for {
		if cur == nil {
			panic(fmt.Sprintf("%s المجهول ليس في النطاق.", name))
		}

		d, ok := cur.Names[name]
		if ok {
			return &d
		}

		cur = cur.Outer
	}
}


