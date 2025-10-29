package main

import "fmt"

var emit = fmt.Println

type Environment struct {
	locals          map[string]int
	nextLocalOffset int
}

func NewEnvironment() *Environment {
	return &Environment{
		locals:          make(map[string]int),
		nextLocalOffset: 0,
	}
}

// AST Interface and Implementations
type AST interface {
	Emit(env *Environment)
	Equals(other AST) bool
}

type Number struct {
	value int
}

func (n Number) Emit(env *Environment) {
	emit(fmt.Sprintf("  ldr r0, =%d", n.value))
}

func (n Number) Equals(other AST) bool {
	if otherNum, ok := other.(*Number); ok {
		return n.value == otherNum.value
	}
	return false
}

type Id struct {
	value string
}

func (i Id) Emit(env *Environment) {
	if offset, exists := env.locals[i.value]; exists {
		emit(fmt.Sprintf("  ldr r0, [fp, #%d]", offset))
	} else {
		panic(fmt.Sprintf("Undefined variable: %s", i.value))
	}
}

func (i Id) Equals(other AST) bool {
	if otherId, ok := other.(*Id); ok {
		return i.value == otherId.value
	}
	return false
}

type Not struct {
	term AST
}

func (n Not) Emit(env *Environment) {
	n.term.Emit(env)
	emit("  cmp r0, #0")
	emit("  moveq r0, #1")
	emit("  movne r0, #0")
}

func (n Not) Equals(other AST) bool {
	if otherNot, ok := other.(*Not); ok {
		return n.term.Equals(otherNot.term)
	}
	return false
}

type Equal struct {
	left, right AST
}

func (e Equal) Emit(env *Environment) {
	e.left.Emit(env)
	emit("  push {r0, ip}")
	e.right.Emit(env)
	emit("  pop {r1, ip}")
	emit("  cmp r0, r1")
	emit("  moveq r0, #1")
	emit("  movne r0, #0")
}

func (e Equal) Equals(other AST) bool {
	if otherEqual, ok := other.(*Equal); ok {
		return e.left.Equals(otherEqual.left) && e.right.Equals(otherEqual.right)
	}
	return false
}

type NotEqual struct {
	left, right AST
}

func (ne NotEqual) Emit(env *Environment) {
	ne.left.Emit(env)
	emit("  push {r0, ip}")
	ne.right.Emit(env)
	emit("  pop {r1, ip}")
	emit("  cmp r0, r1")
	emit("  movne r0, #1")
	emit("  moveq r0, #0")
}

func (ne NotEqual) Equals(other AST) bool {
	if otherNotEqual, ok := other.(*NotEqual); ok {
		return ne.left.Equals(otherNotEqual.left) && ne.right.Equals(otherNotEqual.right)
	}
	return false
}

type Add struct {
	left, right AST
}

func (a Add) Emit(env *Environment) {
	a.left.Emit(env)
	emit("  push {r0, ip}")
	a.right.Emit(env)
	emit("  pop {r1, ip}")
	emit("  add r0, r1, r0")
}

func (a Add) Equals(other AST) bool {
	if otherAdd, ok := other.(*Add); ok {
		return a.left.Equals(otherAdd.left) && a.right.Equals(otherAdd.right)
	}
	return false
}

type Subtract struct {
	left, right AST
}

func (s Subtract) Emit(env *Environment) {
	s.left.Emit(env)
	emit("  push {r0, ip}")
	s.right.Emit(env)
	emit("  pop {r1, ip}")
	emit("  sub r0, r1, r0")
}

func (s Subtract) Equals(other AST) bool {
	if otherSub, ok := other.(*Subtract); ok {
		return s.left.Equals(otherSub.left) && s.right.Equals(otherSub.right)
	}
	return false
}

type Multiply struct {
	left, right AST
}

func (m Multiply) Emit(env *Environment) {
	m.left.Emit(env)
	emit("  push {r0, ip}")
	m.right.Emit(env)
	emit("  pop {r1, ip}")
	emit("  mul r0, r1, r0")
}

func (m Multiply) Equals(other AST) bool {
	if otherMul, ok := other.(*Multiply); ok {
		return m.left.Equals(otherMul.left) && m.right.Equals(otherMul.right)
	}
	return false
}

type Divide struct {
	left, right AST
}

func (d Divide) Emit(env *Environment) {
	d.left.Emit(env)
	emit("  push {r0, ip}")
	d.right.Emit(env)
	emit("  pop {r1, ip}")
	emit("  udiv r0, r1, r0")
}

func (d Divide) Equals(other AST) bool {
	if otherDiv, ok := other.(*Divide); ok {
		return d.left.Equals(otherDiv.left) && d.right.Equals(otherDiv.right)
	}
	return false
}

type Call struct {
	callee string
	args   []AST
}

func (c Call) Emit(env *Environment) {
	count := len(c.args)
	if count == 0 {
		emit(fmt.Sprintf("  bl %s", c.callee))
	} else if count == 1 {
		c.args[0].Emit(env)
		emit(fmt.Sprintf("  bl %s", c.callee))
	} else if count >= 2 && count <= 4 {
		emit("  sub sp, sp, #16")
		for i, arg := range c.args {
			arg.Emit(env)
			emit(fmt.Sprintf("  str r0, [sp, #%d]", 4*i))
		}
		emit("  pop {r0, r1, r2, r3}")
		emit(fmt.Sprintf("  bl %s", c.callee))
	} else {
		panic("More than 4 arguments are not supported")
	}
}

func (c Call) Equals(other AST) bool {
	if otherCall, ok := other.(*Call); ok {
		if c.callee != otherCall.callee || len(c.args) != len(otherCall.args) {
			return false
		}
		for i, arg := range c.args {
			if !arg.Equals(otherCall.args[i]) {
				return false
			}
		}
		return true
	}
	return false
}

type Return struct {
	term AST
}

func (r Return) Emit(env *Environment) {
	r.term.Emit(env)
	emit("  mov sp, fp")
	emit("  pop {fp, pc}")
}

func (r Return) Equals(other AST) bool {
	if otherReturn, ok := other.(*Return); ok {
		return r.term.Equals(otherReturn.term)
	}
	return false
}

type Block struct {
	statements []AST
}

func (b Block) Emit(env *Environment) {
	for _, statement := range b.statements {
		statement.Emit(env)
	}
}

func (b Block) Equals(other AST) bool {
	if otherBlock, ok := other.(*Block); ok {
		if len(b.statements) != len(otherBlock.statements) {
			return false
		}
		for i, stmt := range b.statements {
			if !stmt.Equals(otherBlock.statements[i]) {
				return false
			}
		}
		return true
	}
	return false
}

type If struct {
	conditional, consequence, alternative AST
}

func (i If) Emit(env *Environment) {
	ifFalseLabel := NewLabel()
	endIfLabel := NewLabel()

	i.conditional.Emit(env)
	emit("  cmp r0, #0")
	emit(fmt.Sprintf("  beq %s", ifFalseLabel))
	i.consequence.Emit(env)
	emit(fmt.Sprintf("  b %s", endIfLabel))
	emit(fmt.Sprintf("%s:", ifFalseLabel))
	i.alternative.Emit(env)
	emit(fmt.Sprintf("%s:", endIfLabel))
}

func (i If) Equals(other AST) bool {
	if otherIf, ok := other.(*If); ok {
		return i.conditional.Equals(otherIf.conditional) &&
			i.consequence.Equals(otherIf.consequence) &&
			i.alternative.Equals(otherIf.alternative)
	}
	return false
}

type While struct {
	conditional, body AST
}

func (w While) Emit(env *Environment) {
	loopStart := NewLabel()
	loopEnd := NewLabel()

	emit(fmt.Sprintf("%s:", loopStart))
	w.conditional.Emit(env)
	emit("  cmp r0, #0")
	emit(fmt.Sprintf("  beq %s", loopEnd))
	w.body.Emit(env)
	emit(fmt.Sprintf("  b %s", loopStart))
	emit(fmt.Sprintf("%s:", loopEnd))
}

func (w While) Equals(other AST) bool {
	if otherWhile, ok := other.(*While); ok {
		return w.conditional.Equals(otherWhile.conditional) && w.body.Equals(otherWhile.body)
	}
	return false
}

type Assign struct {
	name  string
	value AST
}

func (a Assign) Emit(env *Environment) {
	a.value.Emit(env)
	if offset, exists := env.locals[a.name]; exists {
		emit(fmt.Sprintf("  str r0, [fp, #%d]", offset))
	} else {
		panic(fmt.Sprintf("Undefined variable: %s", a.name))
	}
}

func (a Assign) Equals(other AST) bool {
	if otherAssign, ok := other.(*Assign); ok {
		return a.name == otherAssign.name && a.value.Equals(otherAssign.value)
	}
	return false
}

type Var struct {
	name  string
	value AST
}

func (v Var) Emit(env *Environment) {
	v.value.Emit(env)
	emit("  push {r0, ip}")
	env.locals[v.name] = env.nextLocalOffset - 4
	env.nextLocalOffset -= 8
}

func (v Var) Equals(other AST) bool {
	if otherVar, ok := other.(*Var); ok {
		return v.name == otherVar.name && v.value.Equals(otherVar.value)
	}
	return false
}

type Function struct {
	name       string
	parameters []string
	body       AST
}

func (f Function) Emit(env *Environment) {
	if len(f.parameters) > 4 {
		panic("More than 4 params is not supported")
	}

	emit("")
	emit(fmt.Sprintf(".global %s", f.name))
	emit(fmt.Sprintf("%s:", f.name))

	f.emitPrologue()
	funcEnv := f.setUpEnvironment()
	f.body.Emit(funcEnv)
	f.emitEpilogue()
}

func (f Function) emitPrologue() {
	emit("  push {fp, lr}")
	emit("  mov fp, sp")
	emit("  push {r0, r1, r2, r3}")
}

func (f Function) setUpEnvironment() *Environment {
	env := NewEnvironment()
	for i, param := range f.parameters {
		env.locals[param] = 4*i - 16
	}
	env.nextLocalOffset = -20
	return env
}

func (f Function) emitEpilogue() {
	emit("  mov sp, fp")
	emit("  mov r0, #0")
	emit("  pop {fp, pc}")
}

func (f Function) Equals(other AST) bool {
	if otherFunc, ok := other.(*Function); ok {
		if f.name != otherFunc.name || len(f.parameters) != len(otherFunc.parameters) {
			return false
		}
		for i, param := range f.parameters {
			if param != otherFunc.parameters[i] {
				return false
			}
		}
		return f.body.Equals(otherFunc.body)
	}
	return false
}

type Main struct {
	statements []AST
}

func (m Main) Emit(env *Environment) {
	emit(".global main")
	emit("main:")
	emit("  push {fp, lr}")
	for _, statement := range m.statements {
		statement.Emit(env)
	}
	emit("  mov r0, #0")
	emit("  pop {fp, pc}")
}

func (m Main) Equals(other AST) bool {
	if otherMain, ok := other.(*Main); ok {
		if len(m.statements) != len(otherMain.statements) {
			return false
		}
		for i, stmt := range m.statements {
			if !stmt.Equals(otherMain.statements[i]) {
				return false
			}
		}
		return true
	}
	return false
}

type Assert struct {
	condition AST
}

func (a Assert) Emit(env *Environment) {
	a.condition.Emit(env)
	emit("  cmp r0, #1")
	emit("  moveq r0, #'.'")
	emit("  movne r0, #'F'")
	emit("  bl putchar")
}

func (a Assert) Equals(other AST) bool {
	if otherAssert, ok := other.(*Assert); ok {
		return a.condition.Equals(otherAssert.condition)
	}
	return false
}

// Label implementation
type Label struct {
	value int
}

var labelCounter = 0

func NewLabel() *Label {
	label := &Label{value: labelCounter}
	labelCounter++
	return label
}

func (l Label) String() string {
	return fmt.Sprintf(".L%d", l.value)
}
