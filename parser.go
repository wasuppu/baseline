package main

import "strconv"

// Parser Combinators for Expressions and Statements
var (
	whitespace = Regexp(`[ \n\r\t]+`)
	comments   = Or(
		Regexp(`//.*`),
		Regexp(`(?s)/\*.*?\*/`),
	)
	ignored = Many(Or(whitespace, comments))
)

func token(pattern string) Parser[string] {
	return Bind(Regexp(pattern), func(value string) Parser[string] {
		return And(ignored, Constant(value))
	})
}

var (
	FUNCTION = token(`function\b`)
	IF       = token(`if\b`)
	WHILE    = token(`while\b`)
	ELSE     = token(`else\b`)
	RETURN   = token(`return\b`)
	VAR      = token(`var\b`)

	COMMA       = token(`,`)
	SEMICOLON   = token(`;`)
	LEFT_PAREN  = token(`\(`)
	RIGHT_PAREN = token(`\)`)
	LEFT_BRACE  = token(`\{`)
	RIGHT_BRACE = token(`\}`)

	NUMBER = Map(token(`[0-9]+`), func(digits string) AST {
		val, _ := strconv.Atoi(digits)
		return Number{value: val}
	})

	ID = token(`[a-zA-Z_][a-zA-Z0-9_]*`)

	idParser = Map(ID, func(x string) AST {
		return Id{value: x}
	})
)

// Operators
var (
	NOT   = Map(token(`!`), func(_ string) AST { return Not{} })
	EQUAL = Map(token(`==`), func(_ string) func(AST, AST) AST {
		return func(l, r AST) AST { return Equal{left: l, right: r} }
	})
	NOT_EQUAL = Map(token(`!=`), func(_ string) func(AST, AST) AST {
		return func(l, r AST) AST { return NotEqual{left: l, right: r} }
	})
	PLUS = Map(token(`\+`), func(_ string) func(AST, AST) AST {
		return func(l, r AST) AST { return Add{left: l, right: r} }
	})
	MINUS = Map(token(`-`), func(_ string) func(AST, AST) AST {
		return func(l, r AST) AST { return Subtract{left: l, right: r} }
	})
	STAR = Map(token(`\*`), func(_ string) func(AST, AST) AST {
		return func(l, r AST) AST { return Multiply{left: l, right: r} }
	})
	SLASH = Map(token(`/`), func(_ string) func(AST, AST) AST {
		return func(l, r AST) AST { return Divide{left: l, right: r} }
	})
	ASSIGN_OP = Map(token(`=`), func(_ string) func(string, AST) AST {
		return func(name string, value AST) AST { return Assign{name: name, value: value} }
	})
)

var (
	expression Parser[AST]
	statement  Parser[AST]
	parser     Parser[AST]
)

func init() {
	// use function to delay initialization in order to avoid cycle initialization
	expression = Parser[AST]{func(source *Source) *ParseResult[AST] {
		return getComparisonParser().Parse(source)
	}}

	statement = Parser[AST]{func(source *Source) *ParseResult[AST] {
		return getStatementParser().Parse(source)
	}}

	parser = Map(And(ignored, Many(statement)),
		func(statements []AST) AST {
			return Block{statements: statements}
		})
}

func getComparisonParser() Parser[AST] {
	// args <- (expression (COMMA expression)*)?
	args := Or(
		Bind(expression, func(arg AST) Parser[[]AST] {
			return Bind(Many(And(COMMA, expression)), func(args []AST) Parser[[]AST] {
				allArgs := append([]AST{arg}, args...)
				return Constant(allArgs)
			})
		}),
		Constant([]AST{}),
	)

	// call <- ID LEFT_PAREN args RIGHT_PAREN
	call := Bind(ID, func(callee string) Parser[AST] {
		return And(LEFT_PAREN, Bind(args, func(args []AST) Parser[AST] {
			if callee == "__assert" {
				return And(RIGHT_PAREN, Constant[AST](Assert{condition: args[0]}))
			} else {
				return And(RIGHT_PAREN, Constant[AST](Call{callee: callee, args: args}))
			}
		}))
	})

	// atom <- call / ID / NUMBER / LEFT_PAREN expression RIGHT_PAREN
	atom := Or(call, idParser, NUMBER,
		Bind(And(LEFT_PAREN, expression), func(e AST) Parser[AST] {
			return And(RIGHT_PAREN, Constant(e))
		}))

	// unary <- NOT? atom
	unary := Bind(Maybe(NOT), func(not *AST) Parser[AST] {
		return Map(atom, func(term AST) AST {
			if not != nil {
				return Not{term: term}
			} else {
				return term
			}
		})
	})

	// product <- unary ((STAR / SLASH) unary)*
	product := infix(Or(STAR, SLASH), unary)

	// sum <- product ((PLUS / MINUS) product)*
	sum := infix(Or(PLUS, MINUS), product)

	// comparison <- sum ((EQUAL / NOT_EQUAL) sum)*
	return infix(Or(EQUAL, NOT_EQUAL), sum)
}

func infix(operatOr Parser[func(AST, AST) AST], termParser Parser[AST]) Parser[AST] {
	return Bind(termParser, func(left AST) Parser[AST] {
		return Bind(Many(
			Bind(operatOr, func(op func(AST, AST) AST) Parser[func(AST) AST] {
				return Bind(termParser, func(right AST) Parser[func(AST) AST] {
					return Constant(func(current AST) AST {
						return op(current, right)
					})
				})
			}),
		), func(ops []func(AST) AST) Parser[AST] {
			result := left
			for _, op := range ops {
				result = op(result)
			}
			return Constant(result)
		})
	})
}

func getStatementParser() Parser[AST] {
	// returnStatement <- RETURN expression SEMICOLON
	returnStatement := Bind(And(RETURN, expression),
		func(term AST) Parser[AST] {
			return And(SEMICOLON, Constant[AST](Return{term: term}))
		})

	// expressionStatement <- expression SEMICOLON
	expressionStatement := Bind(expression, func(term AST) Parser[AST] {
		return And(SEMICOLON, Constant(term))
	})

	// ifStatement <- IF LEFT_PAREN expression RIGHT_PAREN statement ELSE statement
	ifStatement := Bind(And(And(IF, LEFT_PAREN), expression),
		func(conditional AST) Parser[AST] {
			return Bind(And(RIGHT_PAREN, statement), func(consequence AST) Parser[AST] {
				return Bind(And(ELSE, statement), func(alternative AST) Parser[AST] {
					return Constant[AST](If{
						conditional: conditional,
						consequence: consequence,
						alternative: alternative,
					})
				})
			})
		})

	// whileStatement <- WHILE LEFT_PAREN expression RIGHT_PAREN statement
	whileStatement := Bind(And(And(WHILE, LEFT_PAREN), expression),
		func(conditional AST) Parser[AST] {
			return Bind(And(RIGHT_PAREN, statement), func(body AST) Parser[AST] {
				return Constant[AST](While{
					conditional: conditional,
					body:        body,
				})
			})
		})

	// varStatement <- VAR ID ASSIGN expression SEMICOLON
	varStatement := Bind(And(VAR, ID),
		func(name string) Parser[AST] {
			return Bind(And(ASSIGN_OP, expression), func(value AST) Parser[AST] {
				return And(SEMICOLON, Constant[AST](Var{name: name, value: value}))
			})
		})

	// assignmentStatement <- ID ASSIGN expression SEMICOLON
	assignmentStatement := Bind(ID, func(name string) Parser[AST] {
		return Bind(And(ASSIGN_OP, expression), func(value AST) Parser[AST] {
			return And(SEMICOLON, Constant[AST](Assign{name: name, value: value}))
		})
	})

	// blockStatement <- LEFT_BRACE statement* RIGHT_BRACE
	blockStatement := Bind(And(LEFT_BRACE, Many(statement)),
		func(statements []AST) Parser[AST] {
			return And(RIGHT_BRACE, Constant[AST](Block{statements: statements}))
		})

	// functionStatement <- FUNCTION ID LEFT_PAREN parameters RIGHT_PAREN blockStatement
	functionStatement := Bind(And(FUNCTION, ID), func(name string) Parser[AST] {
		return Bind(And(LEFT_PAREN, parameters), func(parameters []string) Parser[AST] {
			return Bind(And(RIGHT_PAREN, blockStatement), func(block AST) Parser[AST] {
				if name == "__main" {
					if blockStmt, ok := block.(Block); ok {
						return Constant[AST](Main(blockStmt))
					}
				}
				return Constant[AST](Function{
					name:       name,
					parameters: parameters,
					body:       block,
				})
			})
		})
	})

	return Or(
		returnStatement,
		functionStatement,
		ifStatement,
		whileStatement,
		varStatement,
		assignmentStatement,
		blockStatement,
		expressionStatement,
	)
}

// parameters <- (ID (COMMA ID)*)?
var parameters = Or(
	Bind(ID, func(param string) Parser[[]string] {
		return Bind(Many(And(COMMA, ID)),
			func(params []string) Parser[[]string] {
				allParams := append([]string{param}, params...)
				return Constant(allParams)
			})
	}),
	Constant([]string{}),
)
