package genllvm

import (
	"fmt"
)

var printScope = false

const (
    IntegerType = "i64"
    BooleanType = "i1"
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
	Terminated bool 
	ScopeType int
	StartLabel string
	EndLabel string
} 

var TopScope *Scope

func pushScope(scopeType int, startLabel string, endLabel string) {
	var newScope = &Scope{
		Outer: TopScope,
		Names: make(map[string]SymbolData),
		Terminated: false,
		ScopeType:  scopeType,
		StartLabel: startLabel,
		EndLabel: endLabel,
	}
	TopScope = newScope
	if printScope {
		printScopeTree("\\", " " + numToScope(scopeType) + ":")
	}
}

func terminateScope() {
	TopScope.Terminated = true
}

func isTerminated() bool {
	var cur = TopScope
	for cur != nil {
		if cur.Terminated {
			return true
		}
		cur = cur.Outer
	}
	return false
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

	data = SymbolData{"@true",BooleanType}
	addToScope("истина", data)
	data = SymbolData{"@false",BooleanType}
	addToScope("ложь", data)
}


func popScope() {
	if printScope {
		printScopeTree("/", "")
	}
	if TopScope == nil{
		panic("Stack is empty.")
	}

	TopScope = TopScope.Outer

}

func addToScope(name string, data SymbolData) {
	TopScope.Names[name] = data
	if printScope {
		suffix := " " + name + " " + data.Typ + " " + data.RegisterNum
		printScopeTree("-", suffix)
	}
}

func numToScope(scope int) string {
	switch scope{
	case GlobalScope:
		return "Global Scope"
	case FuncScope:
		return "Function Scope"
	case LoopScope:
		return "Loop Scope"
	case IfScope:
		return "If Scope"
	default:
		return "Undefined Scope"
	}
}

func printScopeTree(char string, suffix string) {
	var prefix = ""
	var cur = TopScope
	var depth = 0
	if char == "\\"{
		depth--
	}
	for (cur != nil){
		depth++
		cur = cur.Outer
	}
	for depth > 0 {
		prefix += "│   "
		depth--
	}
	if char == "/"{
		prefix += "└──────────────────"
	} else {
		prefix += "├──"
	}
	if printScope{
		println(prefix + suffix)
	}
}

func printScopes() {
	var cur = TopScope
	for cur != nil {
		println(numToScope(cur.ScopeType) + ":")
		for key,name := range cur.Names{
			println(key, name.RegisterNum, name.Typ)
		}
		println("-------------------\n")
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


