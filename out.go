package main

import (
	"fmt"
)

type decl interface{}

type declVoid struct {
}

type declName struct {
	t declType
	n string
}

type declType interface {
	goType() string
	goXdr(valPtr string) string
}

type declTypeTypespec struct {
	t typespec
}

func (t declTypeTypespec) goType() string {
	return t.t.goType()
}

func (t declTypeTypespec) goXdr(valPtr string) string {
	return t.t.goXdr(valPtr)
}

type declTypeArray struct {
	t  typespec
	sz string
}

func (t declTypeArray) goType() string {
	return fmt.Sprintf("[%s]%s", t.sz, t.t.goType())
}

func (t declTypeArray) goXdr(valPtr string) string {
	var res string
	res += fmt.Sprintf("for i := 0; i < %s; i++ {\n", t.sz)
	res += fmt.Sprintf("%s\n", t.t.goXdr(fmt.Sprintf("&(*(%s)[i])", valPtr)))
	res += fmt.Sprintf("}\n")
	return res
}

type declTypeVarArray struct {
	t  typespec
	sz string
}

func (t declTypeVarArray) goType() string {
	return fmt.Sprintf("[]%s", t.t.goType())
}

func (t declTypeVarArray) goXdr(valPtr string) string {
	panic("x")
}

type declTypeOpaqueArray struct {
	sz string
}

func (t declTypeOpaqueArray) goType() string {
	return fmt.Sprintf("[%s]byte", t.sz)
}

func (t declTypeOpaqueArray) goXdr(valPtr string) string {
	return fmt.Sprintf("xdrArray(xs, %s, (*%s)[:]);\n", t.sz, valPtr)
}

type declTypeOpaqueVarArray struct {
	sz string
}

func (t declTypeOpaqueVarArray) goType() string {
	return "[]byte"
}

func (t declTypeOpaqueVarArray) goXdr(valPtr string) string {
	sz := t.sz
	if sz == "" {
		sz = "-1"
	}
	return fmt.Sprintf("xdrVarArray(xs, %s, %s);\n", sz, valPtr)
}

type declTypeString struct {
	sz string
}

func (t declTypeString) goType() string {
	return "string"
}

func (t declTypeString) goXdr(valPtr string) string {
	return fmt.Sprintf("xdrString(xs, %s);\n", valPtr)
}

type declTypePtr struct {
	t typespec
}

func (t declTypePtr) goType() string {
	return fmt.Sprintf("*%s", t.t.goType())
}

func (t declTypePtr) goXdr(valPtr string) string {
	return t.t.goXdr(fmt.Sprintf("*(%s)", valPtr))
}

type typespec interface {
	goType() string
	goXdr(valPtr string) string
}

type typeInt struct {
	unsig bool
}

func (t typeInt) goType() string {
	if t.unsig {
		return "uint32"
	} else {
		return "int32"
	}
}

func (t typeInt) goXdr(valPtr string) string {
	if t.unsig {
		return fmt.Sprintf("xdrU32(xs, %s);\n", valPtr)
	} else {
		return fmt.Sprintf("xdrS32(xs, %s);\n", valPtr)
	}
}

type typeHyper struct {
	unsig bool
}

func (t typeHyper) goType() string {
	if t.unsig {
		return "uint64"
	} else {
		return "int64"
	}
}

func (t typeHyper) goXdr(valPtr string) string {
	if t.unsig {
		return fmt.Sprintf("xdrU64(xs, %s);\n", valPtr)
	} else {
		return fmt.Sprintf("xdrS64(xs, %s);\n", valPtr)
	}
}

type typeFloat struct{}

func (t typeFloat) goType() string { return "float32" }
func (t typeFloat) goXdr(valPtr string) string { panic("x") }

type typeDouble struct{}

func (t typeDouble) goType() string { return "float64" }
func (t typeDouble) goXdr(valPtr string) string { panic("x") }

type typeQuadruple struct{}

func (t typeQuadruple) goType() string { panic("quadruple") }
func (t typeQuadruple) goXdr(valPtr string) string { panic("x") }

type typeBool struct{}

func (t typeBool) goType() string { return "bool" }
func (t typeBool) goXdr(valPtr string) string {
	return fmt.Sprintf("xdrBool(xs, %s);\n", valPtr)
}

type typeEnum struct {
	items []enumItem
}

func (t typeEnum) goType() string { return "int32" }
func (t typeEnum) goXdr(valPtr string) string { panic("x") }

func declToNameGotype(d decl) string {
	switch v := d.(type) {
	case declName:
		return fmt.Sprintf("%s %s;", v.n, v.t.goType())
	}

	return ""
}

type typeStruct struct {
	items []decl
}

func (t typeStruct) goType() string {
	res := "struct { "
	for _, i := range t.items {
		res += declToNameGotype(i)
	}
	res += "}"
	return res
}

func (t typeStruct) goXdr(valPtr string) string {
	var res string
	for _, v := range t.items {
		switch v := v.(type) {
		case declName:
			res += v.t.goXdr(fmt.Sprintf("&(%s.%s)", valPtr, v.n))
		}
	}
	return res
}

type typeUnion struct {
	switchDecl decl
	cases      unionCasesDef
}

func (t typeUnion) goType() string {
	res := "struct { "
	res += declToNameGotype(t.switchDecl)
	for _, i := range t.cases.cases {
		res += declToNameGotype(i.body)
	}
	if t.cases.def != nil {
		res += declToNameGotype(t.cases.def)
	}
	res += "}"
	return res
}

func (t typeUnion) goXdr(valPtr string) string {
	var res string
	var switchName string
	switch v := t.switchDecl.(type) {
	case declVoid:
		panic("void union switch")
	case declName:
		switchName = fmt.Sprintf("%s.%s", valPtr, v.n)
		res += v.t.goXdr(fmt.Sprintf("&(%s)", switchName))
	}
	res += fmt.Sprintf("switch %s {\n", switchName)
	for _, c := range t.cases.cases {
		for i, cval := range c.cases {
			res += fmt.Sprintf("case %s:\n", cval)
			if i != len(c.cases)-1 {
				res += "fallthrough\n"
			}
		}
		switch v := c.body.(type) {
		case declName:
			res += v.t.goXdr(fmt.Sprintf("&(%s.%s)", valPtr, v.n))
		}
	}
	if t.cases.def != nil {
		res += "default:\n"
		switch v := t.cases.def.(type) {
		case declName:
			res += v.t.goXdr(fmt.Sprintf("&(%s.%s)", valPtr, v.n))
		}
	}
	res += "}\n"
	return res
}

type typeIdent struct {
	n string
}

func (t typeIdent) goType() string { return t.n }
func (t typeIdent) goXdr(valPtr string) string {
	return fmt.Sprintf("((%s)*%s).xdr(xs);\n", t.n, valPtr)
}

type enumItem struct {
	name string
	val  string
}

type unionCasesDef struct {
	cases []unionCaseDecl
	def   decl
}

type unionCaseDecl struct {
	cases []string
	body  decl
}

func emitConst(ident string, val string) {
	fmt.Printf("const %s = %s\n", ident, val)
}

func emitTypedef(val decl) {
	switch v := val.(type) {
	case declName:
		fmt.Printf("type %s %s\n", v.n, v.t.goType())

		fmt.Printf("func (v *%s) xdr(xs *xdrState) {\n", v.n)
		fmt.Printf("%s", v.t.goXdr("v"))
		fmt.Printf("}\n")
	}
}

func emitEnum(ident string, val []enumItem) {
	fmt.Printf("type %s int\n", ident)

	fmt.Printf("func (v *%s) xdr(xs *xdrState) {\n", ident)
	fmt.Printf("%s", typeInt{false}.goXdr("v"))
	fmt.Printf("}\n")

	for _, v := range val {
		fmt.Printf("const %s = %s\n", v.name, v.val)
	}
}

func emitStruct(ident string, val []decl) {
	fmt.Printf("type %s struct {\n", ident)
	for _, v := range val {
		switch v := v.(type) {
		case declName:
			fmt.Printf("  %s %s;\n", v.n, v.t.goType())
		}
	}
	fmt.Printf("}\n")

	fmt.Printf("func (v *%s) xdr(xs *xdrState) {\n", ident)
	fmt.Printf("%s", typeStruct{val}.goXdr("v"))
	fmt.Printf("}\n")
}

func emitUnion(ident string, val typeUnion) {
	fmt.Printf("type %s struct {\n", ident)

	switch v := val.switchDecl.(type) {
	case declName:
		fmt.Printf("  %s %s;\n", v.n, v.t.goType())
	}

	for _, c := range val.cases.cases {
		switch v := c.body.(type) {
		case declName:
			fmt.Printf("  %s %s;\n", v.n, v.t.goType())
		}
	}

	if val.cases.def != nil {
		switch v := val.cases.def.(type) {
		case declName:
			fmt.Printf("  %s %s;\n", v.n, v.t.goType())
		}
	}

	fmt.Printf("}\n")

	fmt.Printf("func (v *%s) xdr(xs *xdrState) {\n", ident)
	fmt.Printf("%s", val.goXdr("v"))
	fmt.Printf("}\n")
}
