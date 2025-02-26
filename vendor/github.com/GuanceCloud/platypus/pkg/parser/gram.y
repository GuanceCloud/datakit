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
	DIV GTE GT NOT
	LT LTE MOD MUL
	NEQ EQ EQEQ SUB
	ADD_EQ SUB_EQ DIV_EQ
	MUL_EQ MOD_EQ
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
/* %type <item>
	unary_op */


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
	assignment_stmt
	for_in_stmt
	for_stmt
	continue_stmt
	break_stmt
	ifelse_stmt
	call_expr
	named_arg

%type <node>
	identifier
	unary_expr
	binary_expr
	conditional_expr
	arithmeticExpr
	paren_expr
	index_expr
	attr_expr
	slice_expr
	slice_expr_start
	in_expr
	expr
	map_literal
	map_literal_start
	list_literal
	list_literal_start
	basic_literal
	for_stmt_elem
	value_stmt
	//columnref

%start start

// operator listed with increasing precedence
%right EQ SUB_EQ ADD_EQ MUL_EQ DIV_EQ MOD_EQ
%left OR
%left AND
%left IN
%left GTE GT NEQ EQEQ LTE LT
%left ADD SUB
%left MUL DIV MOD
%right NOT UMINUS
%left LEFT_BRACKET RIGHT_BRACKET LEFT_PAREN RIGHT_PAREN DOT
%%


sep: SEMICOLON
| EOL
| sep SEMICOLON
| sep EOL
;

sem: SEMICOLON
| sem EOL
| sem SEMICOLON
;

start: START_STMTS stmts
{
	yylex.(*parser).parseResult = $2
}
| START_STMTS EOLS
{
	yylex.(*parser).parseResult = ast.Stmts{}
}
| START_STMTS EOLS stmts
{
	yylex.(*parser).parseResult = $3
}
| start EOF
| error
{
	yylex.(*parser).unexpected("", "")
}
;


stmts: stmts_list stmt
{
	s := $1
	s = append(s, $2)
	$$ = s
}
| stmts_list
| stmt
{
	$$ = ast.Stmts{$1}
}
;

stmts_list: stmt sep
{
	$$ = ast.Stmts{$1}
}
| sem
{
	$$ = ast.Stmts{}
}
| stmts_list stmt sep
{
	s := $1
	s = append(s, $2)
	$$ = s
}
;

stmt: ifelse_stmt
| for_in_stmt
| for_stmt
| continue_stmt
| break_stmt
| value_stmt
| assignment_stmt
;

/* expression */
expr: basic_literal 
| list_literal 
| map_literal 
| paren_expr
| call_expr
| unary_expr
| binary_expr
| attr_expr
| slice_expr
| index_expr
| in_expr
| identifier
; // arithmeticExpr

value_stmt: expr
;

EOLS: EOL
| EOLS EOL
;

SPACE_EOLS: EOLS
|
;

assignment_stmt: expr EQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newAssignmentStmt($1, $4, $2)
}
| expr ADD_EQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newAssignmentStmt($1, $4, $2)
}
| expr SUB_EQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newAssignmentStmt($1, $4, $2)
}
| expr MUL_EQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newAssignmentStmt($1, $4, $2)
}
| expr DIV_EQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newAssignmentStmt($1, $4, $2)
}
| expr MOD_EQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newAssignmentStmt($1, $4, $2)
}
;


in_expr: expr IN SPACE_EOLS expr
{
	$$ = yylex.(*parser).newInExpr($1, $4, $2)
}

break_stmt: BREAK
{
	$$ = yylex.(*parser).newBreakStmt($1.Pos)
}
;

continue_stmt: CONTINUE
{
	$$ = yylex.(*parser).newContinueStmt($1.Pos)
}
;

/*
	for identifier IN identifier
	for identifier IN list_init
	for identifier IN string
*/
for_in_stmt : FOR in_expr stmt_block
{
	$$ = yylex.(*parser).newForInStmt($2, $3, $1)
}
;

/*
	for init ; cond expr; loop { stmts }
	for init ; cond expr; 	   { stmts }
	for 	 ; cond expr; loop { stmts }
	for 	 ; cond expr;      { stmts }
*/
for_stmt : FOR for_stmt_elem SEMICOLON expr SEMICOLON for_stmt_elem stmt_block
{
	$$ = yylex.(*parser).newForStmt($2, $4, $6, $7)
}
| FOR for_stmt_elem SEMICOLON expr SEMICOLON stmt_block
{
	$$ = yylex.(*parser).newForStmt($2, $4, nil, $6)
}
| FOR SEMICOLON expr SEMICOLON for_stmt_elem stmt_block
{
	$$ = yylex.(*parser).newForStmt(nil, $3, $5, $6)
}
| FOR SEMICOLON expr SEMICOLON stmt_block
{
	$$ = yylex.(*parser).newForStmt(nil, $3, nil, $5)
}
| FOR for_stmt_elem SEMICOLON SEMICOLON for_stmt_elem stmt_block
{
	$$ = yylex.(*parser).newForStmt($2, nil, $5, $6)
}
| FOR for_stmt_elem SEMICOLON SEMICOLON stmt_block
{
	$$ = yylex.(*parser).newForStmt($2, nil, nil, $5)
}
| FOR SEMICOLON SEMICOLON for_stmt_elem stmt_block
{
	$$ = yylex.(*parser).newForStmt(nil, nil, $4, $5)
}
| FOR SEMICOLON SEMICOLON stmt_block
{
	$$ = yylex.(*parser).newForStmt(nil, nil, nil, $4)
}
;

for_stmt_elem: expr | assignment_stmt
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
{
	$$ = yylex.(*parser).newIfElem($1, $2, $3)
} 
;

if_elif_list: if_elem
{
	$$ = []*ast.IfStmtElem{ $1 }
}
| if_elif_list elif_elem
{
	$$ = append($1, $2)
}
;

elif_elem: ELIF expr stmt_block
{
	$$ = yylex.(*parser).newIfElem($1, $2, $3)
}
;


stmt_block: empty_block
| LEFT_BRACE SPACE_EOLS stmts RIGHT_BRACE
{
	$$ = yylex.(*parser).newBlockStmt($1, $3, $4)
}
;

empty_block : LEFT_BRACE SPACE_EOLS RIGHT_BRACE
{
	$$ = yylex.(*parser).newBlockStmt($1, ast.Stmts{} , $3)
}
;

call_expr: identifier LEFT_PAREN SPACE_EOLS  function_args SPACE_EOLS RIGHT_PAREN
{
	$$ = yylex.(*parser).newCallExpr($1, $4, $2, $6)
}
| identifier LEFT_PAREN SPACE_EOLS function_args COMMA SPACE_EOLS RIGHT_PAREN
{
	$$ = yylex.(*parser).newCallExpr($1, $4, $2, $7)
}
| identifier LEFT_PAREN SPACE_EOLS RIGHT_PAREN
{
	$$ = yylex.(*parser).newCallExpr($1, nil, $2, $4)
}
;


function_args: function_args COMMA SPACE_EOLS expr
{
	$$ = append($$, $4)
}
| function_args COMMA SPACE_EOLS named_arg
{
	$$ = append($$, $4)
}
| named_arg
{
	$$ = []*ast.Node{$1}
}
| expr
{
	$$ = []*ast.Node{$1}
}
;

named_arg: identifier EQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newAssignmentStmt($1, $4, $2)
}
;

unary_expr: ADD expr %prec UMINUS 
{
	$$ = yylex.(*parser).newUnaryExpr($1, $2)
}
| SUB expr %prec UMINUS 
{
	$$ = yylex.(*parser).newUnaryExpr($1, $2)
}
| NOT expr
{
	$$ = yylex.(*parser).newUnaryExpr($1, $2)
}
;

binary_expr: conditional_expr | arithmeticExpr ;

conditional_expr: expr GTE SPACE_EOLS expr
{
	$$ = yylex.(*parser).newConditionalExpr($1, $4, $2)
}
| expr GT SPACE_EOLS expr
{
	$$ = yylex.(*parser).newConditionalExpr($1, $4, $2)
}
| expr OR SPACE_EOLS expr
{
	$$ = yylex.(*parser).newConditionalExpr($1, $4, $2)
}
| expr AND SPACE_EOLS expr
{
	$$ = yylex.(*parser).newConditionalExpr($1, $4, $2)
}
| expr LT SPACE_EOLS expr
{
	$$ = yylex.(*parser).newConditionalExpr($1, $4, $2)
}
| expr LTE SPACE_EOLS expr
{
	$$ = yylex.(*parser).newConditionalExpr($1, $4, $2)
}
| expr NEQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newConditionalExpr($1, $4, $2)
}
| expr EQEQ SPACE_EOLS expr
{
	$$ = yylex.(*parser).newConditionalExpr($1, $4, $2)
}
;


arithmeticExpr: expr ADD SPACE_EOLS expr
{
	$$ = yylex.(*parser).newArithmeticExpr($1, $4, $2)
}
| expr SUB SPACE_EOLS expr
{
	$$ = yylex.(*parser).newArithmeticExpr($1, $4, $2)
}
| expr MUL SPACE_EOLS expr
{
	$$ = yylex.(*parser).newArithmeticExpr($1, $4, $2)
}
| expr DIV SPACE_EOLS expr
{
	$$ = yylex.(*parser).newArithmeticExpr($1, $4, $2)
}
| expr MOD SPACE_EOLS expr
{
	$$ = yylex.(*parser).newArithmeticExpr($1, $4, $2)
}
;

// TODO: 支持多个表达式构成的括号表达式
paren_expr: LEFT_PAREN SPACE_EOLS expr SPACE_EOLS RIGHT_PAREN
{
	$$ = yylex.(*parser).newParenExpr($1, $3, $5)
}
;


index_expr: identifier LEFT_BRACKET SPACE_EOLS expr SPACE_EOLS RIGHT_BRACKET
{
	$$ = yylex.(*parser).newIndexExpr($1, $2 ,$4, $6)
}
| DOT LEFT_BRACKET SPACE_EOLS expr SPACE_EOLS RIGHT_BRACKET	
// 兼容原有语法，仅作为 json 函数的第二个参数
{
	$$ = yylex.(*parser).newIndexExpr(nil, $2, $4, $6)
}
| index_expr LEFT_BRACKET SPACE_EOLS expr SPACE_EOLS RIGHT_BRACKET
{
	$$ = yylex.(*parser).newIndexExpr($1, $2, $4, $6)
}
;


// TODO 实现结构体或类，当前不进行取值操作
// 仅用于 json 函数
attr_expr: identifier DOT index_expr
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

slice_expr_start
: basic_literal 
{
    $$ = $1
}
| list_literal
{
    $$ = $1
}
| slice_expr
{
    $$ = $1
}
;

slice_expr: identifier LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS expr COLON SPACE_EOLS expr RIGHT_BRACKET   //a[1:3:2]
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, $7, $10, true, $2, $11)
}
| identifier LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS expr COLON SPACE_EOLS RIGHT_BRACKET					//a[1:3:]
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, $7, nil, true, $2, $10)
}
| identifier LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS COLON SPACE_EOLS expr RIGHT_BRACKET					//a[1::2]
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, nil, $9, true, $2, $10)
}
| identifier LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS COLON SPACE_EOLS RIGHT_BRACKET						//a[1::]
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, nil, nil, true, $2, $9)
}
| identifier LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS expr COLON SPACE_EOLS expr RIGHT_BRACKET					//a[:3:2]
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, $6, $9, true, $2, $10)
}
| identifier LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS expr COLON SPACE_EOLS RIGHT_BRACKET						//a[:3:]
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, $6, nil, true, $2, $9)
}
| identifier LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS COLON SPACE_EOLS expr RIGHT_BRACKET						//a[::2]
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, nil, $8, true, $2, $9)
}
| identifier LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS COLON SPACE_EOLS RIGHT_BRACKET							//a[::]
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, nil, nil, true, $2, $8)
}
| identifier LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS expr RIGHT_BRACKET									//a[1:3]
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, $7, nil, false, $2, $8)
}
| identifier LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS expr RIGHT_BRACKET										//a[:3]
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, $6, nil, false, $2, $7)
}
| identifier LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS RIGHT_BRACKET										//a[1:]
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, nil, nil, false, $2, $7)
}
| identifier LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS RIGHT_BRACKET												//a[:]
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, nil, nil, false, $2, $6)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS expr COLON SPACE_EOLS expr RIGHT_BRACKET   
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, $7, $10, true, $2, $11)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS expr COLON SPACE_EOLS RIGHT_BRACKET					
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, $7, nil, true, $2, $10)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS COLON SPACE_EOLS expr RIGHT_BRACKET					
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, nil, $9, true, $2, $10)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS COLON SPACE_EOLS RIGHT_BRACKET					
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, nil, nil, true, $2, $9)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS expr COLON SPACE_EOLS expr RIGHT_BRACKET					
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, $6, $9, true, $2, $10)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS expr COLON SPACE_EOLS RIGHT_BRACKET						
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, $6, nil, true, $2, $9)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS COLON SPACE_EOLS expr RIGHT_BRACKET						
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, nil, $8, true, $2, $9)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS COLON SPACE_EOLS RIGHT_BRACKET							
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, nil, nil, true, $2, $8)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS expr RIGHT_BRACKET									
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, $7, nil, false, $2, $8)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS expr RIGHT_BRACKET										
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, $6, nil, false, $2, $7)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS expr COLON SPACE_EOLS RIGHT_BRACKET										
{
	$$ = yylex.(*parser).newSliceExpr($1, $4, nil, nil, false, $2, $7)
}
| slice_expr_start LEFT_BRACKET SPACE_EOLS COLON SPACE_EOLS RIGHT_BRACKET											
{
	$$ = yylex.(*parser).newSliceExpr($1, nil, nil, nil, false, $2, $6)
}
;

list_literal : list_literal_start RIGHT_BRACKET
{
	$$ = yylex.(*parser).newListLiteralEnd($$, $2.Pos)
}
| list_literal_start COMMA SPACE_EOLS RIGHT_BRACKET
{
	$$ = yylex.(*parser).newListLiteralEnd($$, $4.Pos)
}
| LEFT_BRACKET SPACE_EOLS RIGHT_BRACKET
{ 
	$$ = yylex.(*parser).newListLiteralStart($1.Pos)
	$$ = yylex.(*parser).newListLiteralEnd($$, $3.Pos)
}
;

list_literal_start : LEFT_BRACKET SPACE_EOLS expr
{
	$$ = yylex.(*parser).newListLiteralStart($1.Pos)
	$$ = yylex.(*parser).newListLiteralAppendExpr($$, $3)
}
| list_literal_start COMMA SPACE_EOLS expr
{
	$$ = yylex.(*parser).newListLiteralAppendExpr($$, $4)
}
| list_literal_start EOL
;


map_literal : map_literal_start SPACE_EOLS RIGHT_BRACE
{
	$$ = yylex.(*parser).newMapLiteralEnd($$, $3.Pos)
}
|  map_literal_start COMMA SPACE_EOLS RIGHT_BRACE
{
	$$ = yylex.(*parser).newMapLiteralEnd($$, $4.Pos)
}
| empty_block
{ 
	$$ = yylex.(*parser).newMapLiteralStart($1.LBracePos.Pos)
	$$ = yylex.(*parser).newMapLiteralEnd($$, $1.RBracePos.Pos)
}
;

map_literal_start: LEFT_BRACE SPACE_EOLS expr COLON SPACE_EOLS expr
{ 
	$$ = yylex.(*parser).newMapLiteralStart($1.Pos)
	$$ = yylex.(*parser).newMapLiteralAppendExpr($$, $3, $6)
}
| map_literal_start COMMA SPACE_EOLS expr COLON SPACE_EOLS expr
{
	$$ = yylex.(*parser).newMapLiteralAppendExpr($1, $4, $7)
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

basic_literal: NUMBER
{
	$$ =  yylex.(*parser).newNumberLiteral($1)
}
| TRUE
{
	$$ = yylex.(*parser).newBoolLiteral($1.Pos, true)
}
| FALSE
{
	$$ =  yylex.(*parser).newBoolLiteral($1.Pos, false)
}
| STRING
{ 
	$1.Val = yylex.(*parser).unquoteString($1.Val)
	$$ = yylex.(*parser).newStringLiteral($1) 
}
| MULTILINE_STRING
{
	$1.Val = yylex.(*parser).unquoteMultilineString($1.Val)
	$$ = yylex.(*parser).newStringLiteral($1)
}
| NIL
{ 
	$$ = yylex.(*parser).newNilLiteral($1.Pos)
}
| NULL
{
	$$ = yylex.(*parser).newNilLiteral($1.Pos)
}
;

%%
