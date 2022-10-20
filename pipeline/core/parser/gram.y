%{
package parser

import (
	"fmt"
	ast "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
)

%}

%union {
	aststmts   ast.Stmts
	ifitem     *ast.IfStmtElem
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
	identifier


%type<aststmts>
	stmt_block
	stmts
	stmts_list
	empty_block

%type<ifitem>
	if_expr_cond_block

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
	ifelse_stmt_start
	call_expr
	call_expr_start

%type <node>
	function_arg
	binary_expr
	conditional_expr
	arithmeticExpr
	paren_expr
	paren_expr_start
	index_expr
	attr_expr
	expr
	map_init
	map_init_start
	array_list
	array_list_start
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
expr	: array_elem | array_list | map_init | paren_expr | call_expr | binary_expr | attr_expr | index_expr ; // arithmeticExpr


break_stmt: BREAK
			{ $$ = yylex.(*parser).newBreakStmt() }
		;

continue_stmt: CONTINUE
			{ $$ = yylex.(*parser).newContinueStmt() }
		;

/*
	for identifier IN identifier
	for identifier IN array_list
	for identifier IN string
*/
for_in_stmt : FOR identifier IN expr stmt_block
			{ $$ = yylex.(*parser).newForInStmt($2.Val, $4, $5) }
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

ifelse_stmt: ifelse_stmt_start 
		| ifelse_stmt_start ELSE stmt_block
			{ $$ = yylex.(*parser).newIfelseStmt($1 , nil, nil, $3) }
		;


ifelse_stmt_start: IF if_expr_cond_block
				{ $$ = yylex.(*parser).newIfelseStmt(nil ,$2, nil, nil) }
		| ifelse_stmt_start ELIF if_expr_cond_block
				{ $$ = yylex.(*parser).newIfelseStmt($1 , nil, $3, nil) }
		;

if_expr_cond_block	: expr stmt_block
		{ $$ = yylex.(*parser).newIfExpr($1, $2) }
	| stmt_block
		{
		yylex.(*parser).addParseErr(nil, fmt.Errorf("if/elif expr not found condition"))
		$$ = nil
		}
	;



stmt_block	: empty_block
			| LEFT_BRACE stmts RIGHT_BRACE
				{ $$ = $2 }
			;

empty_block : LEFT_BRACE RIGHT_BRACE
				{ $$ = nil }
			;

call_expr : call_expr_start RIGHT_PAREN ;

call_expr_start	: identifier LEFT_PAREN function_args
		{
			f, err := yylex.(*parser).newCallExpr($1.Val, $3)
			if err != nil {
				yylex.(*parser).addParseErr(nil, err)
				$$ = nil
			} else {
				$$ = f
			}
		}
	| identifier LEFT_PAREN
		{
			f, err := yylex.(*parser).newCallExpr($1.Val, nil)
			if err != nil {
				yylex.(*parser).addParseErr(nil, err)
				$$ = nil
			} else {
				$$ = f
			}
		}
	| call_expr_start EOL
	;


function_args	: function_args COMMA function_arg
			{
			$$ = append($$, $3)
			}
		| function_args COMMA
		| function_arg
			{ $$ = []*ast.Node{$1} }
		;

function_arg: expr
	;

// function_arg	: assignment_expr
// 			{ $$ = $1 }
// 		| array_elem
// 			{ $$ = $1 }
// 		| paren_expr
// 			{ $$ = $1 }
// 		| arithmeticExpr
// 			{ $$ = $1 }
// 		| attr_expr
// 			{ $$ = $1 }
// 		| array_list
// 			{ $$ = $1 }
// 		;

binary_expr: conditional_expr | assignment_expr | arithmeticExpr ;

assignment_expr	: expr EQ expr
           		{ $$ = yylex.(*parser).newAssignmentExpr($1, $3) }
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



paren_expr: 	paren_expr_start RIGHT_PAREN

paren_expr_start	: LEFT_PAREN expr 
			{ $$ = ast.WrapParenExpr(&ast.ParenExpr{Param: $2}) }
		| paren_expr_start EOL
		;


index_expr	: identifier LEFT_BRACKET expr RIGHT_BRACKET
			{ $$ = yylex.(*parser).newIndexExpr(
				ast.WrapIdentifier(&ast.Identifier{Name: $1.Val}), $3) }
		| DOT LEFT_BRACKET expr RIGHT_BRACKET	
			// 兼容原有语法，仅作为 json 函数的第二个参数
			{ $$ = yylex.(*parser).newIndexExpr(nil, $3) }
		| index_expr LEFT_BRACKET expr RIGHT_BRACKET
			{ $$ = yylex.(*parser).newIndexExpr($1, $3) }
		;


// TODO 实现结构体或类，当前不进行取值操作
// 仅用于 json 函数
attr_expr	: identifier DOT index_expr
			{ 
				$$ =  yylex.(*parser).newAttrExpr(
					ast.WrapIdentifier(&ast.Identifier{Name: $1.Val}), $3)
			}
		| identifier DOT identifier
			{ 
				$$ =  yylex.(*parser).newAttrExpr(
					ast.WrapIdentifier(&ast.Identifier{Name: $1.Val}), 
					ast.WrapIdentifier(&ast.Identifier{Name: $3.Val}),
					)
			}
		| index_expr DOT index_expr
			{ 
				$$ = yylex.(*parser).newAttrExpr($1, $3)
			}
	  	| index_expr DOT identifier
			{ 
				$$ =  yylex.(*parser).newAttrExpr( $1,
					ast.WrapIdentifier(&ast.Identifier{Name: $3.Val}))
			}
		| attr_expr DOT index_expr
			{ 
				$$ = yylex.(*parser).newAttrExpr($1, $3)
			}
		| attr_expr DOT identifier
			{ 
				$$ =  yylex.(*parser).newAttrExpr( $1,
					ast.WrapIdentifier(&ast.Identifier{Name: $3.Val}))
			}
		;


array_list : array_list_start RIGHT_BRACKET
		| array_list_start COMMA RIGHT_BRACKET
		| LEFT_BRACKET RIGHT_BRACKET
			{ $$ = ast.WrapListInitExpr(&ast.ListInitExpr{}) }
		;

array_list_start : LEFT_BRACKET array_elem
				{ $$ = ast.WrapListInitExpr(&ast.ListInitExpr{List: []*ast.Node{$2}}) }
			| array_list_start COMMA array_elem
					{ $$.ListInitExpr.List = append($$.ListInitExpr.List, $3) }
			| array_list_start EOL
	;


// array_list_item	: array_list_item COMMA array_elem
// 			{ $$.ListInitExpr.List = append($$.ListInitExpr.List, $3) }
// 		| array_elem
// 			{ $$ = ast.WrapListInitExpr(&ast.ListInitExpr{List: []ast.Node{$1}}) }
// 		| /* empty */
// 			{ $$ = ast.WrapListInitExpr(&ast.ListInitExpr{}) }
// 		;


map_init : map_init_start RIGHT_BRACE
		| empty_block
			{ $$ = ast.WrapMapInitExpr(&ast.MapInitExpr{}) }
		;

map_init_start: LEFT_BRACE expr COLON expr
		{ 
			$$ = ast.WrapMapInitExpr(&ast.MapInitExpr{
				KeyValeList: [][2]*ast.Node{{$2, $4}},
			})
		}
	| map_init_start COMMA expr COLON expr
		{
			
			mapInit := $1.MapInitExpr
			mapInit.KeyValeList = append(mapInit.KeyValeList, [2]*ast.Node{$3, $5})
		}
	| map_init_start EOL
	;



array_elem	: bool_literal
		| string_literal
		| nil_literal
		| number_literal
		| identifier
			{ $$ = ast.WrapIdentifier(&ast.Identifier{Name: $1.Val}) }
		;



/*
	literal:
		bool
		number (int float)
		nil
*/
bool_literal	: TRUE
			{ $$ = ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}) }
		| FALSE
			{ $$ = ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}) }
		;


string_literal	: STRING
			{ $$ = ast.WrapStringLiteral(
				&ast.StringLiteral{Val: yylex.(*parser).unquoteString($1.Val)}) }
		| MULTILINE_STRING
			{ $$ = ast.WrapStringLiteral(
				&ast.StringLiteral{Val: yylex.(*parser).unquoteMultilineString($1.Val)}) }
		;


nil_literal	: NIL
			{ $$ = ast.WrapNilLiteral(&ast.NilLiteral{}) }
		| NULL
			{ $$ = ast.WrapNilLiteral(&ast.NilLiteral{}) }
		;



number_literal	: NUMBER
			{ $$ = ast.WrapNumberLiteral(yylex.(*parser).number($1.Val)) }
		| unary_op NUMBER
			{
			num := yylex.(*parser).number($2.Val)
			switch $1.Typ {
			case ADD: // pass
			case SUB:
				if num.IsInt {
					num.Int = -num.Int
				} else {
					num.Float = -num.Float
				}
			}
			$$ = ast.WrapNumberLiteral(num)
			}
		;

identifier: ID
		| QUOTED_STRING
			{ $$.Val = yylex.(*parser).unquoteString($1.Val) }
		;


unary_op	: ADD | SUB ;


%%

