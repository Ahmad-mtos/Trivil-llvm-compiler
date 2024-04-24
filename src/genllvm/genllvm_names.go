package genllvm

import (
	"fmt"
	"trivil/ast"
	"trivil/env"
)

var _ = fmt.Printf

const typeNamePrefix = "T"

const (
	rt_prefix = "tri_"

	rt_init = rt_prefix + "init"

	rt_equalStrings     = rt_prefix + "equalStrings"

	rt_crash = rt_prefix + "crash"
	
)

const (
	nm_VT_field    = "vtable"

	nm_class_info_ptr_suffix = "_class_info_ptr"

	nm_variadic_len_suffic = "_len"
)

func (genc *genContext) localName(prefix string) string {
	return ""
}

//====

func (genc *genContext) declName(d ast.Decl) string {

	f, is_fn := d.(*ast.Function)

	if is_fn && f.External {
		name, ok := f.Mod.Attrs["имя"]
		if !ok {
			name = env.OutName(f.Name)
		}
		genc.declNames[d] = name

		return name
	}

	var out = ""
	var host = d.GetHost()
	if host != nil {
		out = genc.declName(host) + "__"
	}

	var prefix = ""
	if _, ok := d.(*ast.TypeDecl); ok {
		prefix = typeNamePrefix
	}

	out += prefix + env.OutName(d.GetName())

	genc.declNames[d] = out

	return out
}

func (genc *genContext) outName(name string) string {
	return ""
}

func (genc *genContext) functionName(f *ast.Function) string {
	return ""
}
