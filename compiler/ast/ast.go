package ast

import (
	"fmt"
	"github.com/defiant00/char/compiler/token"
	"strings"
)

type General interface {
	Print(int)
}

type Statement interface {
	General
	isStmt()
}

type Expression interface {
	General
	isExpr()
}

func printIndent(indent int) {
	for i := 0; i < indent; i++ {
		fmt.Print("|   ")
	}
}

type Error struct {
	Val string
}

func (this *Error) isStmt() {}
func (this *Error) isExpr() {}

func (this *Error) Print(indent int) {
	printIndent(indent)
	fmt.Printf("ERROR: %v\n", this.Val)
}

type ExprStmt struct {
	Expr Expression
}

func (this *ExprStmt) isStmt() {}

func (this *ExprStmt) Print(indent int) {
	this.Expr.Print(indent)
}

type File struct {
	Name       string
	statements []Statement
}

func (this *File) Print(indent int) {
	printIndent(indent)
	fmt.Println(this.Name)
	for _, s := range this.statements {
		s.Print(indent + 1)
	}
}

func (this *File) AddStmt(stmt Statement) {
	this.statements = append(this.statements, stmt)
}

type Use struct {
	packages []usePackage
}

func (this *Use) isStmt() {}

func (this *Use) Print(indent int) {
	printIndent(indent)
	fmt.Println("use")
	for _, p := range this.packages {
		printIndent(indent + 1)
		fmt.Println(p)
	}
}

func (this *Use) AddPackage(pack, alias string) {
	this.packages = append(this.packages, usePackage{pack, alias})
}

type usePackage struct {
	pack, alias string
}

func (this usePackage) String() string {
	if this.alias != "" {
		return fmt.Sprintf("%v as %v", this.pack, this.alias)
	}
	return this.pack
}

type Class struct {
	Mixin      bool
	Name       string
	typeParams []string
	withs      []Statement
	statements []Statement
}

func (this *Class) isStmt() {}

func (this *Class) Print(indent int) {
	printIndent(indent)
	if this.Mixin {
		fmt.Print("mixin ")
	}
	fmt.Print("class ", this.Name)
	if len(this.typeParams) > 0 {
		fmt.Print("<", strings.Join(this.typeParams, ", "), ">")
	}
	if len(this.withs) > 0 {
		fmt.Print(" with ")
		for i, w := range this.withs {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(w)
		}
	}
	fmt.Println()
	for _, s := range this.statements {
		s.Print(indent + 1)
	}
}

func (this *Class) AddTypeParam(t string) {
	this.typeParams = append(this.typeParams, t)
}

func (this *Class) AddWith(s Statement) {
	this.withs = append(this.withs, s)
}

func (this *Class) AddStmt(s Statement) {
	this.statements = append(this.statements, s)
}

type TypeIdent struct {
	idents     []string
	typeParams []Statement
}

func (this *TypeIdent) isStmt() {}

func (this *TypeIdent) Print(indent int) {
	fmt.Print("TypeIdentifier ", this)
}

func (this *TypeIdent) String() string {
	ret := strings.Join(this.idents, ".")
	if len(this.typeParams) > 0 {
		ret += "<"
	}
	for i, t := range this.typeParams {
		if i > 0 {
			ret += ", "
		}
		ret += fmt.Sprint(t)
	}
	if len(this.typeParams) > 0 {
		ret += ">"
	}
	return ret
}

func (this *TypeIdent) AddIdent(ident string) {
	this.idents = append(this.idents, ident)
}

func (this *TypeIdent) AddTypeParam(s Statement) {
	this.typeParams = append(this.typeParams, s)
}

type TypeRedirect struct {
	Type Statement
	Name string
}

func (this *TypeRedirect) isStmt() {}

func (this *TypeRedirect) Print(indent int) {
	printIndent(indent)
	fmt.Printf("%v as %v\n", this.Type, this.Name)
}

type FuncSigType struct {
	params  []Statement
	returns []Statement
}

func (this *FuncSigType) isExpr() {}
func (this *FuncSigType) isStmt() {}

func (this *FuncSigType) Print(indent int) {
	printIndent(indent)
	fmt.Println(this)
}

func (this *FuncSigType) String() string {
	ret := "fn("
	for i, p := range this.params {
		if i > 0 {
			ret += ", "
		}
		ret += fmt.Sprint(p)
	}
	ret += ")"
	if len(this.returns) > 0 {
		ret += " "
		if len(this.returns) > 1 {
			ret += "("
		}
		for i, r := range this.returns {
			if i > 0 {
				ret += ", "
			}
			ret += fmt.Sprint(r)
		}
		if len(this.returns) > 1 {
			ret += ")"
		}
	}
	return ret
}

func (this *FuncSigType) AddParam(p Statement) {
	this.params = append(this.params, p)
}

func (this *FuncSigType) AddReturn(r Statement) {
	this.returns = append(this.returns, r)
}

type IotaStmt struct{}

func (this *IotaStmt) isStmt() {}

func (this *IotaStmt) Print(indent int) {
	printIndent(indent)
	fmt.Println("iota reset")
}

type PropertySet struct {
	props []property
	Vals  Expression
}

func (this *PropertySet) isStmt() {}

func (this *PropertySet) Print(indent int) {
	printIndent(indent)
	fmt.Print("prop set: ")
	for i, p := range this.props {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(p)
	}
	fmt.Println()
	if this.Vals != nil {
		this.Vals.Print(indent + 1)
	}
}

func (this *PropertySet) AddProp(static bool, name string, typ Statement) {
	this.props = append(this.props, property{static: static, name: name, typ: typ})
}

type property struct {
	static bool
	name   string
	typ    Statement
}

func (this property) String() string {
	var ret string
	if this.static {
		ret = "static "
	}
	ret += this.name
	if this.typ != nil {
		ret += fmt.Sprintf(" %v", this.typ)
	}
	return ret
}

type UnaryExpr struct {
	Expr Expression
	Op   token.Type
}

func (this *UnaryExpr) isExpr() {}

func (this *UnaryExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println(this.Op)
	this.Expr.Print(indent + 1)
}

type BinaryExpr struct {
	Left, Right Expression
	Op          token.Type
}

func (this *BinaryExpr) isExpr() {}

func (this *BinaryExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println(this.Op)
	this.Left.Print(indent + 1)
	this.Right.Print(indent + 1)
}

type AssignStmt struct {
	Left, Right Expression
	Op          token.Type
}

func (this *AssignStmt) isStmt() {}

func (this *AssignStmt) Print(indent int) {
	printIndent(indent)
	fmt.Println("assign", this.Op)
	this.Left.Print(indent + 1)
	this.Right.Print(indent + 1)
}

type StringExpr struct {
	Val string
}

func (this *StringExpr) isExpr() {}

func (this *StringExpr) Print(indent int) {
	printIndent(indent)
	fmt.Printf("string '%v'\n", this.Val)
}

type NumberExpr struct {
	Val string
}

func (this *NumberExpr) isExpr() {}

func (this *NumberExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("number", this.Val)
}

type CharExpr struct {
	Val string
}

func (this *CharExpr) isExpr() {}

func (this *CharExpr) Print(indent int) {
	printIndent(indent)
	fmt.Printf("char '%v'\n", this.Val)
}

type BoolExpr struct {
	Val bool
}

func (this *BoolExpr) isExpr() {}

func (this *BoolExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("bool", this.Val)
}

type IotaExpr struct{}

func (this *IotaExpr) isExpr() {}

func (this *IotaExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("iota")
}

type BlankExpr struct{}

func (this *BlankExpr) isExpr() {}

func (this *BlankExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("_")
}

type IdentExpr struct {
	idents []*IdentPart
}

func (this *IdentExpr) isExpr() {}

func (this *IdentExpr) Print(indent int) {
	printIndent(indent)
	for i, id := range this.idents {
		if i > 0 {
			fmt.Print(".")
		}
		fmt.Print(id)
	}
	fmt.Println()
}

func (this *IdentExpr) AddIdent(i *IdentPart) {
	this.idents = append(this.idents, i)
}

type FuncCallExpr struct {
	Function Expression
	Params   *ExprList
}

func (this *FuncCallExpr) isExpr() {}

func (this *FuncCallExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("func")
	this.Function.Print(indent + 2)
	printIndent(indent + 1)
	fmt.Println("params")
	if this.Params != nil {
		this.Params.Print(indent + 2)
	}
}

type AccessorExpr struct {
	Object Expression
	Index  Expression
}

func (this *AccessorExpr) isExpr() {}

func (this *AccessorExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("accessor")
	this.Object.Print(indent + 2)
	printIndent(indent + 1)
	fmt.Println("index")
	this.Index.Print(indent + 2)
}

type AccessorRangeExpr struct {
	Object    Expression
	Low, High Expression
}

func (this *AccessorRangeExpr) isExpr() {}

func (this *AccessorRangeExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("range accessor")
	this.Object.Print(indent + 2)
	printIndent(indent + 1)
	fmt.Println("from")
	if this.Low != nil {
		this.Low.Print(indent + 2)
	} else {
		printIndent(indent + 2)
		fmt.Println("implicit 0")
	}
	printIndent(indent + 1)
	fmt.Println("to")
	if this.High != nil {
		this.High.Print(indent + 2)
	} else {
		printIndent(indent + 2)
		fmt.Println("implicit length - 1")
	}
}

type IdentPart struct {
	Name       string
	typeParams []Statement
}

func (this *IdentPart) String() string {
	ret := this.Name
	if len(this.typeParams) > 0 {
		ret += "<"
	}
	for i, tp := range this.typeParams {
		if i > 0 {
			ret += ", "
		}
		ret += fmt.Sprint(tp)
	}
	if len(this.typeParams) > 0 {
		ret += ">"
	}
	return ret
}

func (this *IdentPart) AddTypeParam(s Statement) {
	this.typeParams = append(this.typeParams, s)
}

func (this *IdentPart) ResetTypeParams() {
	this.typeParams = make([]Statement, 0)
}

type FuncDef struct {
	Static     bool
	Name       string
	params     []param
	returns    []Statement
	statements []Statement
}

func (this *FuncDef) isExpr() {}
func (this *FuncDef) isStmt() {}

func (this *FuncDef) Print(indent int) {
	printIndent(indent)
	if this.Static {
		fmt.Print("static ")
	}
	if this.Name == "" {
		fmt.Print("fn")
	} else {
		fmt.Print(this.Name)
	}
	fmt.Print("(")
	for i, p := range this.params {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(p)
	}
	fmt.Print(")")

	if len(this.returns) > 0 {
		fmt.Print(" ")
		if len(this.returns) > 1 {
			fmt.Print("(")
		}
		for i, r := range this.returns {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(r)
		}
		if len(this.returns) > 1 {
			fmt.Print(")")
		}
	}

	fmt.Println()
	for _, s := range this.statements {
		s.Print(indent + 1)
	}
}

func (this *FuncDef) AddStmt(s Statement) {
	this.statements = append(this.statements, s)
}

func (this *FuncDef) AddParam(name string, typ Statement) {
	this.params = append(this.params, param{name: name, typ: typ})
}

func (this *FuncDef) AddReturn(r Statement) {
	this.returns = append(this.returns, r)
}

type param struct {
	name string
	typ  Statement
}

func (this param) String() string {
	ret := this.name
	if this.typ != nil {
		ret += fmt.Sprint(" ", this.typ)
	}
	return ret
}

type VarSet struct {
	lines []*VarSetLine
}

func (this *VarSet) isStmt() {}

func (this *VarSet) Print(indent int) {
	printIndent(indent)
	fmt.Println("var set")
	for _, l := range this.lines {
		l.Print(indent + 1)
	}
}

func (this *VarSet) AddLine(vsl *VarSetLine) {
	this.lines = append(this.lines, vsl)
}

type VarSetLine struct {
	vars []variable
	Vals *ExprList
}

func (this *VarSetLine) isStmt() {}

func (this *VarSetLine) Print(indent int) {
	printIndent(indent)
	for i, v := range this.vars {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(v)
	}
	fmt.Println()
	if this.Vals != nil {
		this.Vals.Print(indent + 1)
	}
}

func (this *VarSetLine) AddVar(name string, typ Statement) {
	this.vars = append(this.vars, variable{name: name, typ: typ})
}

type variable struct {
	name string
	typ  Statement
}

func (this variable) String() string {
	ret := this.name
	if this.typ != nil {
		ret += fmt.Sprint(" ", this.typ)
	}
	return ret
}

type ExprList struct {
	exprs []Expression
}

func (this *ExprList) isExpr() {}

func (this *ExprList) Print(indent int) {
	printIndent(indent)
	fmt.Println("expression list")
	for _, e := range this.exprs {
		e.Print(indent + 1)
	}
}

func (this *ExprList) AddExpr(e Expression) {
	this.exprs = append(this.exprs, e)
}

type ReturnStmt struct {
	Vals Expression
}

func (this *ReturnStmt) isStmt() {}

func (this *ReturnStmt) Print(indent int) {
	printIndent(indent)
	fmt.Println("ret")
	if this.Vals != nil {
		this.Vals.Print(indent + 1)
	}
}

type DeferStmt struct {
	Expr Expression
}

func (this *DeferStmt) isStmt() {}

func (this *DeferStmt) Print(indent int) {
	printIndent(indent)
	fmt.Println("defer")
	this.Expr.Print(indent + 1)
}

type InterfaceStmt struct {
	Name     string
	funcSigs []Statement
}

func (this *InterfaceStmt) isStmt() {}

func (this *InterfaceStmt) Print(indent int) {
	printIndent(indent)
	fmt.Println("interface", this.Name)
	for _, f := range this.funcSigs {
		f.Print(indent + 1)
	}
}

func (this *InterfaceStmt) AddFuncSig(i Statement) {
	this.funcSigs = append(this.funcSigs, i)
}

type IntfFuncSig struct {
	Name    string
	params  []Statement
	returns []Statement
}

func (this *IntfFuncSig) isStmt() {}

func (this *IntfFuncSig) Print(indent int) {
	printIndent(indent)
	fmt.Println(this)
}

func (this *IntfFuncSig) String() string {
	ret := this.Name + "("
	for i, p := range this.params {
		if i > 0 {
			ret += ", "
		}
		ret += fmt.Sprint(p)
	}
	ret += ")"
	if len(this.returns) > 0 {
		ret += " "
		if len(this.returns) > 1 {
			ret += "("
		}
		for i, r := range this.returns {
			if i > 0 {
				ret += ", "
			}
			ret += fmt.Sprint(r)
		}
		if len(this.returns) > 1 {
			ret += ")"
		}
	}
	return ret
}

func (this *IntfFuncSig) AddParam(p Statement) {
	this.params = append(this.params, p)
}

func (this *IntfFuncSig) AddReturn(r Statement) {
	this.returns = append(this.returns, r)
}

type If struct {
	Condition Expression
	With      Statement
	stmts     []Statement
}

func (this *If) isStmt() {}

func (this *If) Print(indent int) {
	printIndent(indent)
	fmt.Println("if")
	if this.Condition != nil {
		this.Condition.Print(indent + 1)
	}
	if this.With != nil {
		printIndent(indent + 1)
		fmt.Println("with")
		this.With.Print(indent + 2)
	}
	printIndent(indent + 1)
	fmt.Println("then")
	for _, s := range this.stmts {
		s.Print(indent + 2)
	}
}

func (this *If) AddStmt(s Statement) {
	this.stmts = append(this.stmts, s)
}

type Is struct {
	Condition Expression
	stmts     []Statement
}

func (this *Is) isStmt() {}

func (this *Is) Print(indent int) {
	printIndent(indent)
	fmt.Println("is")
	this.Condition.Print(indent + 1)
	printIndent(indent + 1)
	fmt.Println("then")
	for _, s := range this.stmts {
		s.Print(indent + 2)
	}
}

func (this *Is) AddStmt(s Statement) {
	this.stmts = append(this.stmts, s)
}

type For struct {
	Label string
	vars  []string
	In    Expression
	stmts []Statement
}

func (this *For) isStmt() {}

func (this *For) Print(indent int) {
	printIndent(indent)
	if len(this.Label) > 0 {
		fmt.Print(this.Label, ": ")
	}
	fmt.Println("for", strings.Join(this.vars, ", "), "in")
	this.In.Print(indent + 2)
	for _, s := range this.stmts {
		s.Print(indent + 1)
	}
}

func (this *For) AddVar(v string) {
	this.vars = append(this.vars, v)
}

func (this *For) AddStmt(s Statement) {
	this.stmts = append(this.stmts, s)
}

type Loop struct {
	Label string
	stmts []Statement
}

func (this *Loop) isStmt() {}

func (this *Loop) Print(indent int) {
	printIndent(indent)
	if len(this.Label) > 0 {
		fmt.Print(this.Label, ": ")
	}
	fmt.Println("loop")
	for _, s := range this.stmts {
		s.Print(indent + 1)
	}
}

func (this *Loop) AddStmt(s Statement) {
	this.stmts = append(this.stmts, s)
}

type Break struct {
	Label string
}

func (this *Break) isStmt() {}

func (this *Break) Print(indent int) {
	printIndent(indent)
	fmt.Println("break", this.Label)
}
