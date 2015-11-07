package ast

import (
	"fmt"
	"strings"
)

func printIndent(indent int) {
	for i := 0; i < indent; i++ {
		fmt.Print("|   ")
	}
}

func Print(obj General, indent int) {
	if obj == nil {
		return
	}
	printIndent(indent)
	switch t := obj.(type) {
	case *Accessor:
		this := obj.(*Accessor)
		fmt.Println("accessor")
		Print(this.Object, indent+2)
		printIndent(indent + 1)
		fmt.Println("index")
		Print(this.Index, indent+2)
	case *AccessorRange:
		this := obj.(*AccessorRange)
		fmt.Println("range accessor")
		Print(this.Object, indent+2)
		printIndent(indent + 1)
		fmt.Println("from")
		if this.Low != nil {
			Print(this.Low, indent+2)
		} else {
			printIndent(indent + 2)
			fmt.Println("implicit 0")
		}
		printIndent(indent + 1)
		fmt.Println("to")
		if this.High != nil {
			Print(this.High, indent+2)
		} else {
			printIndent(indent + 2)
			fmt.Println("implicit length - 1")
		}
	case *Alias:
		this := obj.(*Alias)
		fmt.Printf("%v as %v\n", this.Val, this.Alias)
	case *ArrayCons:
		this := obj.(*ArrayCons)
		fmt.Println(this)
		Print(this.Size, indent+1)
	case *Array:
		fmt.Println(obj)
	case *ArrayValueList:
		fmt.Println("array val list")
		Print(obj.(*ArrayValueList).Vals, indent+1)
	case *Assign:
		this := obj.(*Assign)
		fmt.Println("assign", this.Op)
		Print(this.Left, indent+1)
		Print(this.Right, indent+1)
	case *Binary:
		this := obj.(*Binary)
		fmt.Println(this.Op)
		Print(this.Left, indent+1)
		Print(this.Right, indent+1)
	case *Blank:
		fmt.Println("_")
	case *Bool:
		fmt.Println("bool", obj.(*Bool).Val)
	case *Break:
		fmt.Println("break", obj.(*Break).Label)
	case *Char:
		fmt.Printf("char '%v'\n", obj.(*Char).Val)
	case *Class:
		this := obj.(*Class)
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
			Print(s, indent+1)
		}
	case *Constructor:
		this := obj.(*Constructor)
		fmt.Println("cons")
		Print(this.Type, indent+2)
		printIndent(indent + 1)
		fmt.Println("vals")
		for _, p := range this.params {
			Print(p, indent+2)
		}
	case *Defer:
		fmt.Println("defer")
		Print(obj.(*Defer).Expr, indent+1)
	case *Error:
		fmt.Printf("ERROR: %v\n", obj.(*Error).Val)
	case *ExprList:
		fmt.Println("expression list")
		for _, e := range obj.(*ExprList).exprs {
			Print(e, indent+1)
		}
	case *ExprStmt:
		fmt.Println("expr stmt")
		Print(obj.(*ExprStmt).Expr, indent+1)
	case *File:
		this := obj.(*File)
		fmt.Println(this.Name)
		for _, s := range this.statements {
			Print(s, indent+1)
		}
	case *For:
		this := obj.(*For)
		if len(this.Label) > 0 {
			fmt.Print(this.Label, ": ")
		}
		fmt.Println("for", strings.Join(this.vars, ", "), "in")
		Print(this.In, indent+2)
		for _, s := range this.stmts {
			Print(s, indent+1)
		}
	case *FunctionCall:
		this := obj.(*FunctionCall)
		fmt.Println("func")
		Print(this.Function, indent+2)
		if this.Params != nil {
			printIndent(indent + 1)
			fmt.Println("params")
			Print(this.Params, indent+2)
		}
	case *FunctionDef:
		this := obj.(*FunctionDef)
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
			Print(s, indent+1)
		}
	case *FunctionSig:
		fmt.Println(obj)
	case *Identifier:
		for i, id := range obj.(*Identifier).idents {
			if i > 0 {
				fmt.Print(".")
			}
			fmt.Print(id)
		}
		fmt.Println()
	case *If:
		this := obj.(*If)
		fmt.Println("if")
		Print(this.Condition, indent+1)
		if this.With != nil {
			printIndent(indent + 1)
			fmt.Println("with")
			Print(this.With, indent+2)
		}
		printIndent(indent + 1)
		fmt.Println("then")
		for _, s := range this.stmts {
			Print(s, indent+2)
		}
	case *Interface:
		this := obj.(*Interface)
		fmt.Print("interface ", this.Name)
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
		for _, f := range this.funcSigs {
			Print(f, indent+1)
		}
	case *IntfFuncSig:
		fmt.Println(obj.(*IntfFuncSig))
	case *Iota:
		fmt.Println("iota")
	case *Is:
		this := obj.(*Is)
		fmt.Println("is")
		Print(this.Condition, indent+1)
		printIndent(indent + 1)
		fmt.Println("then")
		for _, s := range this.stmts {
			Print(s, indent+2)
		}
	case *KeyVal:
		this := obj.(*KeyVal)
		fmt.Printf("%v:\n", this.Key)
		Print(this.Val, indent+1)
	case *Loop:
		this := obj.(*Loop)
		if len(this.Label) > 0 {
			fmt.Print(this.Label, ": ")
		}
		fmt.Println("loop")
		for _, s := range this.stmts {
			Print(s, indent+1)
		}
	case *Number:
		fmt.Println("number", obj.(*Number).Val)
	case *PropertySet:
		this := obj.(*PropertySet)
		fmt.Print("prop set: ")
		for i, p := range this.props {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(p)
		}
		fmt.Println()
		Print(this.Vals, indent+1)
	case *Return:
		fmt.Println("ret")
		Print(obj.(*Return).Vals, indent+1)
	case *String:
		fmt.Printf("string '%v'\n", obj.(*String).Val)
	case *Unary:
		this := obj.(*Unary)
		fmt.Println(this.Op)
		Print(this.Expr, indent+1)
	case *Use:
		this := obj.(*Use)
		fmt.Println("use")
		for _, p := range this.packages {
			printIndent(indent + 1)
			fmt.Println(p)
		}
	case *VarSet:
		fmt.Println("var set")
		for _, l := range obj.(*VarSet).lines {
			Print(l, indent+1)
		}
	case *VarSetLine:
		this := obj.(*VarSetLine)
		for i, v := range this.vars {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v)
		}
		fmt.Println()
		Print(this.Vals, indent+1)
	default:
		fmt.Println("Unknown type %T", t)
	}
}
