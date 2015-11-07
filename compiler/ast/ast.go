package ast

import (
	"fmt"
	"github.com/defiant00/char/compiler/token"
	"strings"
)

type General interface{}

type Expression interface {
	isExpr()
}

type Statement interface {
	isStmt()
}

// Expressions
func (this *Accessor) isExpr()       {}
func (this *AccessorRange) isExpr()  {}
func (this *ArrayCons) isExpr()      {}
func (this *ArrayValueList) isExpr() {}
func (this *Binary) isExpr()         {}
func (this *Blank) isExpr()          {}
func (this *Bool) isExpr()           {}
func (this *Char) isExpr()           {}
func (this *Constructor) isExpr()    {}
func (this *Error) isExpr()          {}
func (this *ExprList) isExpr()       {}
func (this *FunctionCall) isExpr()   {}
func (this *FunctionDef) isExpr()    {}
func (this *FunctionSig) isExpr()    {}
func (this *Identifier) isExpr()     {}
func (this *Iota) isExpr()           {}
func (this *Number) isExpr()         {}
func (this *String) isExpr()         {}
func (this *Unary) isExpr()          {}

// Statements
func (this *Alias) isStmt()       {}
func (this *Array) isStmt()       {}
func (this *Assign) isStmt()      {}
func (this *Break) isStmt()       {}
func (this *Class) isStmt()       {}
func (this *Defer) isStmt()       {}
func (this *Error) isStmt()       {}
func (this *ExprStmt) isStmt()    {}
func (this *For) isStmt()         {}
func (this *FunctionDef) isStmt() {}
func (this *FunctionSig) isStmt() {}
func (this *If) isStmt()          {}
func (this *Interface) isStmt()   {}
func (this *IntfFuncSig) isStmt() {}
func (this *Iota) isStmt()        {}
func (this *Is) isStmt()          {}
func (this *KeyVal) isStmt()      {}
func (this *Loop) isStmt()        {}
func (this *PropertySet) isStmt() {}
func (this *Return) isStmt()      {}
func (this *TypeIdent) isStmt()   {}
func (this *Use) isStmt()         {}
func (this *VarSet) isStmt()      {}
func (this *VarSetLine) isStmt()  {}

type Accessor struct {
	Object Expression
	Index  Expression
}

type AccessorRange struct {
	Object    Expression
	Low, High Expression
}

type Alias struct {
	Val   Statement
	Alias string
}

type ArrayCons struct {
	Type Statement
	Size Expression
}

func (this *ArrayCons) String() string {
	return fmt.Sprintf("cons []%v", this.Type)
}

type Array struct {
	Type Statement
}

func (this *Array) String() string {
	return fmt.Sprintf("[]%v", this.Type)
}

type ArrayValueList struct {
	Vals Expression
}

type Assign struct {
	Left, Right Expression
	Op          token.Type
}

type Binary struct {
	Left, Right Expression
	Op          token.Type
}

type Blank struct{}

type Bool struct {
	Val bool
}

type Break struct {
	Label string
}

type Char struct {
	Val string
}

type Class struct {
	Mixin      bool
	Name       string
	typeParams []string
	withs      []Statement
	statements []Statement
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

type Constructor struct {
	Type   Expression
	params []Statement
}

func (this *Constructor) AddParam(p Statement) {
	this.params = append(this.params, p)
}

type Defer struct {
	Expr Expression
}

type Error struct {
	Val string
}

type ExprList struct {
	exprs []Expression
}

func (this *ExprList) AddExpr(e Expression) {
	this.exprs = append(this.exprs, e)
}

type ExprStmt struct {
	Expr Expression
}

type File struct {
	Name       string
	statements []Statement
}

func (this *File) AddStmt(stmt Statement) {
	this.statements = append(this.statements, stmt)
}

type For struct {
	Label string
	vars  []string
	In    Expression
	stmts []Statement
}

func (this *For) AddVar(v string) {
	this.vars = append(this.vars, v)
}

func (this *For) AddStmt(s Statement) {
	this.stmts = append(this.stmts, s)
}

type FunctionCall struct {
	Function Expression
	Params   Expression
}

type FunctionDef struct {
	Static     bool
	Name       string
	params     []param
	returns    []Statement
	statements []Statement
}

func (this *FunctionDef) AddStmt(s Statement) {
	this.statements = append(this.statements, s)
}

func (this *FunctionDef) AddParam(name string, typ Statement) {
	this.params = append(this.params, param{name: name, typ: typ})
}

func (this *FunctionDef) AddReturn(r Statement) {
	this.returns = append(this.returns, r)
}

type FunctionSig struct {
	params  []Statement
	returns []Statement
}

func (this *FunctionSig) String() string {
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

func (this *FunctionSig) AddParam(p Statement) {
	this.params = append(this.params, p)
}

func (this *FunctionSig) AddReturn(r Statement) {
	this.returns = append(this.returns, r)
}

type Identifier struct {
	idents []*IdentPart
}

func (this *Identifier) AddIdent(i *IdentPart) {
	this.idents = append(this.idents, i)
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

type If struct {
	Condition Expression
	With      Statement
	stmts     []Statement
}

func (this *If) AddStmt(s Statement) {
	this.stmts = append(this.stmts, s)
}

type Interface struct {
	Name     string
	withs    []Statement
	funcSigs []Statement
}

func (this *Interface) AddWith(w Statement) {
	this.withs = append(this.withs, w)
}

func (this *Interface) AddFuncSig(i Statement) {
	this.funcSigs = append(this.funcSigs, i)
}

type IntfFuncSig struct {
	Name    string
	params  []Statement
	returns []Statement
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

type Iota struct{}

type Is struct {
	Condition Expression
	stmts     []Statement
}

func (this *Is) AddStmt(s Statement) {
	this.stmts = append(this.stmts, s)
}

type KeyVal struct {
	Key string
	Val Expression
}

type Loop struct {
	Label string
	stmts []Statement
}

func (this *Loop) AddStmt(s Statement) {
	this.stmts = append(this.stmts, s)
}

type Number struct {
	Val string
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

type PropertySet struct {
	props []property
	Vals  Expression
}

func (this *PropertySet) AddProp(static bool, name string, typ Statement) {
	this.props = append(this.props, property{static: static, name: name, typ: typ})
}

type Return struct {
	Vals Expression
}

type String struct {
	Val string
}

type TypeIdent struct {
	idents     []string
	typeParams []Statement
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

type Unary struct {
	Expr Expression
	Op   token.Type
}

type Use struct {
	packages []usePackage
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

type VarSet struct {
	lines []*VarSetLine
}

func (this *VarSet) AddLine(vsl *VarSetLine) {
	this.lines = append(this.lines, vsl)
}

type VarSetLine struct {
	vars []variable
	Vals Expression
}

func (this *VarSetLine) AddVar(name string, typ Statement) {
	this.vars = append(this.vars, variable{name: name, typ: typ})
}
