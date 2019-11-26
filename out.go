package main

import (
	"fmt"
	"strings"
)

func i(ident string) string {
	switch ident {
	case "TRUE":
		return "true"
	case "FALSE":
		return "false"
	default:
		return strings.ToUpper(ident[:1]) + ident[1:]
	}
}

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
	var res string
	res += fmt.Sprintf("{\n")
	res += fmt.Sprintf("var __arraysz uint32\n")
	res += fmt.Sprintf("xs.EncodingSetSize(&__arraysz, len(*%s));\n", valPtr)
	res += fmt.Sprintf("%s\n", typeInt{true}.goXdr("&__arraysz"))
	if t.sz != "" {
		res += fmt.Sprintf("if __arraysz > %s { xs.SetError(\"array too large\") } else {\n", t.sz)
	}
	res += fmt.Sprintf("if xs.Decoding() { *%s = make([]%s, __arraysz); }\n", valPtr, t.t.goType())
	res += fmt.Sprintf("for i := uint64(0); i < uint64(__arraysz); i++ {\n")
	res += fmt.Sprintf("%s\n", t.t.goXdr(fmt.Sprintf("&((*(%s))[i])", valPtr)))
	res += fmt.Sprintf("}\n")
	if t.sz != "" {
		res += fmt.Sprintf("}\n")
	}
	res += fmt.Sprintf("}\n")
	return res
}

type declTypeOpaqueArray struct {
	sz string
}

func (t declTypeOpaqueArray) goType() string {
	return fmt.Sprintf("[%s]byte", t.sz)
}

func (t declTypeOpaqueArray) goXdr(valPtr string) string {
	return fmt.Sprintf("xdr.XdrArray(xs, (*%s)[:]);\n", valPtr)
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
	return fmt.Sprintf("xdr.XdrVarArray(xs, int(%s), (*[]byte)(%s));\n", sz, valPtr)
}

type declTypeString struct {
	sz string
}

func (t declTypeString) goType() string {
	return "string"
}

func (t declTypeString) goXdr(valPtr string) string {
	sz := t.sz
	if sz == "" {
		sz = "-1"
	}
	return fmt.Sprintf("xdr.XdrString(xs, int(%s), (*string)(%s));\n", sz, valPtr)
}

type declTypePtr struct {
	t typespec
}

func (t declTypePtr) goType() string {
	return fmt.Sprintf("*%s", t.t.goType())
}

func (t declTypePtr) goXdr(valPtr string) string {
	var res string
	res += fmt.Sprintf("if xs.Encoding() {\n")
	res += fmt.Sprintf("opted := *(%s) != nil\n", valPtr)
	res += typeBool{}.goXdr("&opted")
	res += fmt.Sprintf("if opted {\n")
	res += t.t.goXdr(fmt.Sprintf("*(%s)", valPtr))
	res += fmt.Sprintf("}\n")
	res += fmt.Sprintf("}\n")

	res += fmt.Sprintf("if xs.Decoding() {\n")
	res += fmt.Sprintf("var opted bool\n")
	res += typeBool{}.goXdr("&opted")
	res += fmt.Sprintf("if opted {\n")
	res += fmt.Sprintf("*(%s) = new(%s)\n", valPtr, t.t.goType())
	res += t.t.goXdr(fmt.Sprintf("*(%s)", valPtr))
	res += fmt.Sprintf("}\n")
	res += fmt.Sprintf("}\n")
	return res
}

type typespec interface {
	goType() string
	goXdr(valPtr string) string
}

type typespecOpt struct {
	isVoid bool
	t      typespec
}

func (t typespecOpt) maybeGoType() string {
	if t.isVoid {
		return ""
	}

	return t.t.goType()
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
		return fmt.Sprintf("xdr.XdrU32(xs, (*uint32)(%s));\n", valPtr)
	} else {
		return fmt.Sprintf("xdr.XdrS32(xs, (*int32)(%s));\n", valPtr)
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
		return fmt.Sprintf("xdr.XdrU64(xs, (*uint64)(%s));\n", valPtr)
	} else {
		return fmt.Sprintf("xdr.XdrS64(xs, (*int64)(%s));\n", valPtr)
	}
}

type typeFloat struct{}

func (t typeFloat) goType() string             { return "float32" }
func (t typeFloat) goXdr(valPtr string) string { panic("x") }

type typeDouble struct{}

func (t typeDouble) goType() string             { return "float64" }
func (t typeDouble) goXdr(valPtr string) string { panic("x") }

type typeQuadruple struct{}

func (t typeQuadruple) goType() string             { panic("quadruple") }
func (t typeQuadruple) goXdr(valPtr string) string { panic("x") }

type typeBool struct{}

func (t typeBool) goType() string { return "bool" }
func (t typeBool) goXdr(valPtr string) string {
	return fmt.Sprintf("xdr.XdrBool(xs, (*bool)(%s));\n", valPtr)
}

type typeEnum struct {
	items []enumItem
}

func (t typeEnum) goType() string             { return "int32" }
func (t typeEnum) goXdr(valPtr string) string { panic("x") }

func declToNameGotype(d decl) string {
	switch v := d.(type) {
	case declName:
		return fmt.Sprintf("%s %s;", i(v.n), v.t.goType())
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
			res += v.t.goXdr(fmt.Sprintf("&((%s).%s)", valPtr, i(v.n)))
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
		switchName = fmt.Sprintf("(%s).%s", valPtr, i(v.n))
		res += v.t.goXdr(fmt.Sprintf("&(%s)", switchName))
	}
	res += fmt.Sprintf("switch %s {\n", switchName)
	for _, c := range t.cases.cases {
		for idx, cval := range c.cases {
			res += fmt.Sprintf("case %s:\n", i(cval))
			if idx != len(c.cases)-1 {
				res += "fallthrough\n"
			}
		}
		switch v := c.body.(type) {
		case declName:
			res += v.t.goXdr(fmt.Sprintf("&((%s).%s)", valPtr, i(v.n)))
		}
	}
	if t.cases.def != nil {
		res += "default:\n"
		switch v := t.cases.def.(type) {
		case declName:
			res += v.t.goXdr(fmt.Sprintf("&((%s).%s)", valPtr, i(v.n)))
		}
	}
	res += "}\n"
	return res
}

type typeIdent struct {
	n string
}

func (t typeIdent) goType() string { return i(t.n) }
func (t typeIdent) goXdr(valPtr string) string {
	return fmt.Sprintf("(*%s)(%s).Xdr(xs);\n", i(t.n), valPtr)
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

type progCall struct {
	name string
	arg  typespecOpt
	res  typespecOpt
	id   string
}

type progVer struct {
	name  string
	calls []progCall
	id    string
}

type progDef struct {
	name string
	vers []progVer
	id   string
}

func emitProg(d progDef) {
	fmt.Fprintf(tout, "const %s uint32 = %s\n", i(d.name), d.id)
	for _, v := range d.vers {
		fmt.Fprintf(tout, "const %s uint32 = %s\n", i(v.name), v.id)

		for _, c := range v.calls {
			fmt.Fprintf(tout, "const %s uint32 = %s\n", i(c.name), c.id)
		}

		fmt.Fprintf(out, "type %s_%s_handler interface {\n", i(d.name), i(v.name))
		for _, c := range v.calls {
			fmt.Fprintf(out, "%s(%s) %s\n", i(c.name), c.arg.maybeGoType(), c.res.maybeGoType())
		}
		fmt.Fprintf(out, "}\n")

		fmt.Fprintf(out, "type %s_%s_handler_wrapper struct {\n", i(d.name), i(v.name))
		fmt.Fprintf(out, "h %s_%s_handler\n", i(d.name), i(v.name))
		fmt.Fprintf(out, "}\n")

		for _, c := range v.calls {
			fmt.Fprintf(out, "func (w *%s_%s_handler_wrapper) %s(args *xdr.XdrState) (res xdr.Xdrable, err error) {\n",
				i(d.name), i(v.name), i(c.name))
			if !c.arg.isVoid {
				fmt.Fprintf(out, "var in %s\n", c.arg.t.goType())
				fmt.Fprintf(out, "in.Xdr(args)\n")
				fmt.Fprintf(out, "err = args.Error()\n")
				fmt.Fprintf(out, "if err != nil { return }\n")
			}

			if !c.res.isVoid {
				fmt.Fprintf(out, "var out %s\n", c.res.t.goType())
			} else {
				fmt.Fprintf(out, "var out xdr.Void\n")
			}

			if !c.res.isVoid {
				fmt.Fprintf(out, "out = ")
			}
			fmt.Fprintf(out, "w.h.%s(", i(c.name))
			if !c.arg.isVoid {
				fmt.Fprintf(out, "in")
			}
			fmt.Fprintf(out, ")\n")

			fmt.Fprintf(out, "return &out, nil")
			fmt.Fprintf(out, "}\n")
		}

		fmt.Fprintf(out, "func %s_%s_regs(h %s_%s_handler) []xdr.ProcRegistration {\n", i(d.name), i(v.name), i(d.name), i(v.name))
		fmt.Fprintf(out, "w := &%s_%s_handler_wrapper{h}\n", i(d.name), i(v.name))
		fmt.Fprintf(out, "return []xdr.ProcRegistration{\n")
		for _, c := range v.calls {
			fmt.Fprintf(out, "xdr.ProcRegistration{\n")
			fmt.Fprintf(out, "Prog: %s,\n", i(d.name))
			fmt.Fprintf(out, "Vers: %s,\n", i(v.name))
			fmt.Fprintf(out, "Proc: %s,\n", i(c.name))
			fmt.Fprintf(out, "Handler: w.%s,\n", i(c.name))
			fmt.Fprintf(out, "},\n")
		}
		fmt.Fprintf(out, "}\n")
		fmt.Fprintf(out, "}\n")
	}
}

func emitConst(ident string, val string) {
	fmt.Fprintf(tout, "const %s %s = %s\n", ident, *constTypeFlag, val)
}

func emitTypedef(val decl) {
	switch v := val.(type) {
	case declName:
		// Pointers get wrapped in an extra level of struct at
		// the top level, so that we can define a pointer receiver
		// on them.
		goType := v.t.goType()
		goRef := "v"
		switch v.t.(type) {
		case declTypePtr:
			goType = fmt.Sprintf("struct { P %s }", goType)
			goRef = "&v.P"
		}

		fmt.Fprintf(tout, "type %s %s\n", i(v.n), goType)

		fmt.Fprintf(out, "func (v *%s) Xdr(xs *xdr.XdrState) {\n", i(v.n))
		fmt.Fprintf(out, "%s", v.t.goXdr(goRef))
		fmt.Fprintf(out, "}\n")
	}
}

func emitEnum(ident string, val []enumItem) {
	t := "int32"
	if *unsignedEnumFlag {
		t = "uint32"
	}

	fmt.Fprintf(tout, "type %s %s\n", i(ident), t)

	fmt.Fprintf(out, "func (v *%s) Xdr(xs *xdr.XdrState) {\n", i(ident))
	fmt.Fprintf(out, "%s", typeInt{*unsignedEnumFlag}.goXdr("v"))
	fmt.Fprintf(out, "}\n")

	for _, v := range val {
		fmt.Fprintf(tout, "const %s %s = %s\n", i(v.name), i(ident), v.val)
	}
}

func emitStruct(ident string, val []decl) {
	fmt.Fprintf(tout, "type %s struct {\n", i(ident))
	for _, v := range val {
		switch v := v.(type) {
		case declName:
			fmt.Fprintf(tout, "  %s %s;\n", i(v.n), v.t.goType())
		}
	}
	fmt.Fprintf(tout, "}\n")

	fmt.Fprintf(out, "func (v *%s) Xdr(xs *xdr.XdrState) {\n", i(ident))
	fmt.Fprintf(out, "%s", typeStruct{val}.goXdr("v"))
	fmt.Fprintf(out, "}\n")
}

func emitUnion(ident string, val typeUnion) {
	fmt.Fprintf(tout, "type %s %s\n", i(ident), val.goType())

	fmt.Fprintf(out, "func (v *%s) Xdr(xs *xdr.XdrState) {\n", i(ident))
	fmt.Fprintf(out, "%s", val.goXdr("v"))
	fmt.Fprintf(out, "}\n")
}
