// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

%{
package parser

import (
	ast "github.com/GuanceCloud/platypus/pkg/ast"
)

%}

%union {
	aststmts   ast.Stmts
	astblock   *ast.BlockStmt

	ifitem     *ast.IfStmtElem
	iflist	   []*ast.IfStmtElem
	node       *ast.Node
	nodes      []*ast.Node
	item       Item

}

%token <item> SEMICOLON COMMA COMMENT DOT EOF ERROR ID NUMBER 
	LEFT_PAREN LEFT_BRACKET LEFT_BRACE RIGHT_BRACE
	RIGHT_PAREN RIGHT_BRACKET SPACE STRING QUOTED_STRING MULTILINE_STRING
	FOR IN WHILE BREAK CONTINUE	RETURN EOL COLON
	STR INT FLOAT BOOL LIST MAP

// operator
%token operatorsStart
%token <item> ADD
	DIV GTE GT
	LT LTE MOD MUL
	NEQ EQ EQEQ SUB
%token operatorsEnd

// keywords
%token keywordsStart
%token <item>
TRUE FALSE IDENTIFIER AND OR 
NIL NULL IF ELIF ELSE
%token keywordsEnd

// start symbols for parser
%token startSymbolsStart
%token START_STMTS
%token startSymbolsEnd

////////////////////////////////////////////////////
// grammar rules
////////////////////////////////////////////////////
%type <item>
	unary_op


%type<astblock>
	stmt_block
	empty_block


%type<aststmts>
	stmts
	stmts_list

%type<ifitem>
	if_elem
	elif_elem

%type<iflist>
	if_elif_list

%type<nodes>
	function_args

%type <node>
	stmt
	assignment_expr
	for_in_stmt
	for_stmt
	continue_stmt
	break_stmt
	ifelse_stmt
	call_expr

%type <node>
	identifier
	binary_expr
	conditional_expr
	arithmeticExpr
	paren_expr
	index_expr
	attr_expr
	/* in_expr */
	expr
	map_init
	map_init_start
	list_init
	list_init_start
	array_elem
	bool_literal
	string_literal
	nil_literal
	number_literal
	value_stmt
	//columnref

%start start

// operator listed with increasing precedence
%right EQ
%left OR
%left AND
%left GTE GT NEQ EQEQ LTE LT
%left ADD SUB
%left MUL DIV MOD

%%


sep : SEMICOLON
	| EOL
	| sep SEMICOLON
	| sep EOL

start	: START_STMTS stmts
		{ yylex.(*parser).parseResult = $2 }
	| start EOF
	| error
		{ yylex.(*parser).unexpected("", "") }
	;


stmts: stmts_list stmt
		{
		s := $1
		s = append(s, $2)
		$$ = s
		}
	| stmts_list
	| stmt
		{ $$ = ast.Stmts{$1} }	
	;

stmts_list	: stmt sep
		{ $$ = ast.Stmts{$1} }
	| sep
		{ $$ = ast.Stmts{} }
	| stmts_list stmt sep
		{
		s := $1
		s = append(s, $2)
		$$ = s
		}
	;

stmt	: ifelse_stmt
	| for_in_stmt
	| for_stmt
	| continue_stmt
	| break_stmt
	| value_stmt
	;


value_stmt: expr
	;

/* expression */
expr	: array_elem | list_init | map_init | paren_expr | call_expr | binary_expr | attr_expr | index_expr ; // arithmeticExpr


break_stmt: BREAK
			{ $$ = yylex.(*parser).newBreakStmt($1.Pos) }
		;

continue_stmt: CONTINUE
			{ $$ = yylex.(*parser).newContinueStmt($1.Pos) }
		;

/*
	for identifier IN identifier
	for identifier IN list_init
	for identifier IN string
*/
for_in_stmt : FOR identifier IN expr stmt_block
			{ $$ = yylex.(*parser).newForInStmt($2, $4, $5, $1, $3) }
		;


/*
	for init expr; cond expr; loop expr  block_smt
	for init expr; cond expr; 			 block_stmt
	for 		 ; cond expr; loop expr  block_stmt
	for 		 ; cond expr; 		     block_stmt
*/
for_stmt : FOR expr SEMICOLON expr SEMICOLON expr stmt_block
		{ $$ = yylex.(*parser).newForStmt($2, $4, $6, $7) }
	| FOR expr SEMICOLON expr SEMICOLON stmt_block
		{ $$ = yylex.(*parser).newForStmt($2, $4, nil, $6) }
	| FOR SEMICOLON expr SEMICOLON expr stmt_block
		{ $$ = yylex.(*parser).newForStmt(nil, $3, $5, $6) }
	| FOR SEMICOLON expr SEMICOLON stmt_block
		{ $$ = yylex.(*parser).newForStmt(nil, $3, nil, $5) }

	| FOR expr SEMICOLON SEMICOLON expr stmt_block
		{ $$ = yylex.(*parser).newForStmt($2, nil, $5, $6) }
	| FOR expr SEMICOLON SEMICOLON stmt_block
		{ $$ = yylex.(*parser).newForStmt($2, nil, nil, $5) }
	| FOR SEMICOLON SEMICOLON expr stmt_block
		{ $$ = yylex.(*parser).newForStmt(nil, nil, $4, $5) }
	| FOR SEMICOLON SEMICOLON stmt_block
		{ $$ = yylex.(*parser).newForStmt(nil, nil, nil, $4) }
	;

ifelse_stmt: if_elif_list
		{
			$$ = yylex.(*parser).newIfElifStmt($1)
		}
	| if_elif_list ELSE stmt_block
		{
			$$ = yylex.(*parser).newIfElifelseStmt($1, $2, $3)
		}
	;

if_elem: IF expr stmt_block
	{ $$ = yylex.(*parser).newIfElem($1, $2, $3) } 
	;

if_elif_list: if_elem
		{ $$ = []*ast.IfStmtElem{ $1 } }
	| if_elif_list elif_elem
		{ $$ = append($1, $2) }
	;

elif_elem: ELIF expr stmt_block
		{ $$ = yylex.(*parser).newIfElem($1, $2, $3) }
	;


stmt_block	: empty_block
	| LEFT_BRACE stmts RIGHT_BRACE
		{ $$ = yylex.(*parser).newBlockStmt($1, $2, $3) }
	;

empty_block : LEFT_BRACE RIGHT_BRACE
		{ $$ = yylex.(*parser).newBlockStmt($1, ast.Stmts{} , $2) }
	;


call_expr : identifier LEFT_PAREN function_args RIGHT_PAREN
		{
			$$ = yylex.(*parser).newCallExpr($1, $3, $2, $4)
		}
	| identifier LEFT_PAREN RIGHT_PAREN
		{
			$$ = yylex.(*parser).newCallExpr($1, nil, $2, $3)
		}
	| identifier LEFT_PAREN function_args EOLS RIGHT_PAREN
		{
			$$ = yylex.(*parser).newCallExpr($1, $3, $2, $5)
		}
	| identifier LEFT_PAREN EOLS RIGHT_PAREN
		{
			$$ = yylex.(*parser).newCallExpr($1, nil, $2, $4)
		}
	;


function_args	: function_args COMMA expr
			{
			$$ = append($$, $3)
			}
		| function_args COMMA
		| expr
			{ $$ = []*ast.Node{$1} }
		;


binary_expr: conditional_expr | assignment_expr | arithmeticExpr ;

assignment_expr	: expr EQ expr
           		{ $$ = yylex.(*parser).newAssignmentExpr($1, $3, $2) }
		;

conditional_expr	: expr GTE expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr GT expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr OR expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr AND expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr LT expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr LTE expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr NEQ expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr EQEQ expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			;


arithmeticExpr	: expr ADD expr
				{ $$ = yylex.(*parser).newArithmeticExpr($1, $3, $2) }
			| expr SUB expr
				{ $$ = yylex.(*parser).newArithmeticExpr($1, $3, $2) }
			| expr MUL expr
				{ $$ = yylex.(*parser).newArithmeticExpr($1, $3, $2) }
			| expr DIV expr
				{ $$ = yylex.(*parser).newArithmeticExpr($1, $3, $2) }
			| expr MOD expr
				{ $$ = yylex.(*parser).newArithmeticExpr($1, $3, $2) }
			;

// TODO: 支持多个表达式构成的括号表达式
paren_expr: LEFT_PAREN expr RIGHT_PAREN
			{ $$ = yylex.(*parser).newParenExpr($1, $2, $3) }
		| LEFT_PAREN expr EOLS RIGHT_PAREN
			{ $$ = yylex.(*parser).newParenExpr($1, $2, $4) }
		;


EOLS: EOL
	| EOLS EOL
	;

index_expr	: identifier LEFT_BRACKET expr RIGHT_BRACKET
			{ $$ = yylex.(*parser).newIndexExpr($1, $2 ,$3, $4) }
		| DOT LEFT_BRACKET expr RIGHT_BRACKET	
			// 兼容原有语法，仅作为 json 函数的第二个参数
			{ $$ = yylex.(*parser).newIndexExpr(nil, $2, $3, $4) }
		| index_expr LEFT_BRACKET expr RIGHT_BRACKET
			{ $$ = yylex.(*parser).newIndexExpr($1, $2, $3, $4) }
		;


// TODO 实现结构体或类，当前不进行取值操作
// 仅用于 json 函数
attr_expr	: identifier DOT index_expr
			{ 
				$$ =  yylex.(*parser).newAttrExpr($1, $3)
			}
		| identifier DOT identifier
			{ 
				$$ =  yylex.(*parser).newAttrExpr($1, $3)
			}
		| index_expr DOT index_expr
			{ 
				$$ = yylex.(*parser).newAttrExpr($1, $3)
			}
	  	| index_expr DOT identifier
			{ 
				$$ =  yylex.(*parser).newAttrExpr($1, $3)
			}
		| attr_expr DOT index_expr
			{ 
				$$ = yylex.(*parser).newAttrExpr($1, $3)
			}
		| attr_expr DOT identifier
			{ 
				$$ =  yylex.(*parser).newAttrExpr($1, $3)
			}
		;


list_init : list_init_start RIGHT_BRACKET
			{
				$$ = yylex.(*parser).newListInitEndExpr($$, $2.Pos)
			}
		| list_init_start COMMA RIGHT_BRACKET
			{
				$$ = yylex.(*parser).newListInitEndExpr($$, $2.Pos)
			}
		| LEFT_BRACKET RIGHT_BRACKET
			{ 
				$$ = yylex.(*parser).newListInitStartExpr($1.Pos)
				$$ = yylex.(*parser).newListInitEndExpr($$, $2.Pos)

			}
		;

list_init_start : LEFT_BRACKET expr
			{ 
				$$ = yylex.(*parser).newListInitStartExpr($1.Pos)
				$$ = yylex.(*parser).newListInitAppendExpr($$, $2)
			}
		| list_init_start COMMA expr
				{				
					$$ = yylex.(*parser).newListInitAppendExpr($$, $3)
				}
		| list_init_start EOL
	;


map_init : map_init_start RIGHT_BRACE
			{
				$$ = yylex.(*parser).newMapInitEndExpr($$, $2.Pos)
			}
		|  map_init_start COMMA RIGHT_BRACE
			{
				$$ = yylex.(*parser).newMapInitEndExpr($$, $3.Pos)
			}
		| empty_block
			{ 
				$$ = yylex.(*parser).newMapInitStartExpr($1.LBracePos.Pos)
				$$ = yylex.(*parser).newMapInitEndExpr($$, $1.RBracePos.Pos)
			}
		;

map_init_start: LEFT_BRACE expr COLON expr
		{ 
			$$ = yylex.(*parser).newMapInitStartExpr($1.Pos)
			$$ = yylex.(*parser).newMapInitAppendExpr($$, $2, $4)
		}
	| map_init_start COMMA expr COLON expr
		{
			$$ = yylex.(*parser).newMapInitAppendExpr($1, $3, $5)
		}
	| map_init_start EOL
	;



array_elem	: bool_literal
		| string_literal
		| nil_literal
		| number_literal
		| identifier
		;

/*
	literal:
		bool
		number (int float)
		nil
*/
bool_literal	: TRUE
			{ $$ = yylex.(*parser).newBoolLiteral($1.Pos, true) }
		| FALSE
			{ $$ =  yylex.(*parser).newBoolLiteral($1.Pos, false) }
		;


string_literal	: STRING
			{ 
				$1.Val = yylex.(*parser).unquoteString($1.Val)
				$$ = yylex.(*parser).newStringLiteral($1) 
			}
		| MULTILINE_STRING
			{
				$1.Val = yylex.(*parser).unquoteMultilineString($1.Val)
				$$ = yylex.(*parser).newStringLiteral($1)
			}
		;


nil_literal	: NIL
			{ $$ = yylex.(*parser).newNilLiteral($1.Pos) }
		| NULL
			{ $$ = yylex.(*parser).newNilLiteral($1.Pos) }
		;



number_literal	: NUMBER
			{ $$ =  yylex.(*parser).newNumberLiteral($1) }
		| unary_op NUMBER
			{
			num :=  yylex.(*parser).newNumberLiteral($2) 
			switch $1.Typ {
			case ADD: // pass
			case SUB:
				if num.NodeType == ast.TypeFloatLiteral {
					num.FloatLiteral.Val = -num.FloatLiteral.Val
					num.FloatLiteral.Start = yylex.(*parser).posCache.LnCol($1.Pos)
				} else {
					num.IntegerLiteral.Val = -num.IntegerLiteral.Val
					num.IntegerLiteral.Start = yylex.(*parser).posCache.LnCol($1.Pos)

				}
			}
			$$ = num
			}
		;

identifier: ID
			{
				$$ = yylex.(*parser).newIdentifierLiteral($1)
			}
		| QUOTED_STRING
			{
				$1.Val = yylex.(*parser).unquoteString($1.Val) 
				$$ = yylex.(*parser).newIdentifierLiteral($1)
			}
		;


unary_op	: ADD | SUB ;


%%

