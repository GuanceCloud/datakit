%{
package parser

import (
	"fmt"
)
%}

%union {
	node       Node
	nodes      []Node
	item       Item
}

%token <item> SEMICOLON COMMA COMMENT DOT EOF ERROR ID NUMBER 
	LEFT_PAREN LEFT_BRACKET LEFT_BRACE RIGHT_BRACE
	RIGHT_PAREN RIGHT_BRACKET SPACE STRING QUOTED_STRING

// operator
%token operatorsStart
%token <item> ADD
	DIV GTE GT
	LT LTE MOD MUL
	NEQ EQ EQEQ POW SUB
%token operatorsEnd

// keywords
%token keywordsStart
%token <item>
TRUE FALSE IDENTIFIER AND OR 
NIL NULL RE JP IF ELIF ELSE
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
	function_name
	identifier

%type<nodes>
	function_args
	elif_list

%type <node>
	stmts
	stmt
	assignment_stmt
	ifelse_stmt
	ifelse_block_stmt
	function_stmt

%type <node>
	function_arg
	if_expr
	conditional_expr
	// computation_expr
	paren_expr
	index_expr
	attr_expr
	regex
	expr
	array_list
	array_elem
	bool_literal
	string_literal
	nil_literal
	number_literal
	//columnref

%start start

// operator listed with increasing precedence
%left OR
%left AND
%left GTE GT NEQ EQ EQEQ LTE LT
%left ADD SUB
%left MUL DIV MOD
%right POW

%%


start	: START_STMTS stmts
		{ yylex.(*parser).parseResult = $2 }
	| start EOF
	| error
		{ yylex.(*parser).unexpected("", "") }
	;


stmts	: stmt
		{ $$ = Stmts{$1} }
	| stmts stmt
		{
		s := $1.(Stmts)
		s = append(s, $2)
		$$ = s
		}
	;


stmt	: function_stmt
		{ $$ = Stmts{$1} }
	| ifelse_stmt
		{ $$ = Stmts{$1} }
	| assignment_stmt
		{ $$ = Stmts{$1} }
	;


/* expression */
expr	: array_elem | regex | paren_expr | conditional_expr | attr_expr; // computation_expr 


ifelse_stmt	: IF if_expr
           		{ $$ = yylex.(*parser).newIfelseStmt($2, nil, nil) }
           	| IF if_expr elif_list
           		{ $$ = yylex.(*parser).newIfelseStmt($2, $3, nil) }
           	| IF if_expr ELSE ifelse_block_stmt
           		{ $$ = yylex.(*parser).newIfelseStmt($2, nil, $4) }
           	| IF if_expr elif_list ELSE ifelse_block_stmt
           		{ $$ = yylex.(*parser).newIfelseStmt($2, $3, $5) }
           	;


elif_list	: elif_list ELIF if_expr
			{ $$ = append($1, $3) }
		| ELIF if_expr
			{ $$ = []Node{$2} }
		;


if_expr	: conditional_expr ifelse_block_stmt
		{ $$ = yylex.(*parser).newIfExpr($1, $2) }
	| bool_literal ifelse_block_stmt
		{ $$ = yylex.(*parser).newIfExpr($1, $2) }
	| ifelse_block_stmt
		{
		yylex.(*parser).addParseErr(nil, fmt.Errorf("if/elif expr not found condition"))
		$$ = nil 
		}
	;


ifelse_block_stmt	: LEFT_BRACE RIGHT_BRACE
				{ $$ = nil }
			| LEFT_BRACE stmts RIGHT_BRACE
				{ $$ = $2 }
			;


function_stmt	: function_name LEFT_PAREN function_args RIGHT_PAREN
			{
			f, err := yylex.(*parser).newFuncStmt($1.Val, $3)
			if err != nil {
				yylex.(*parser).addParseErr(nil, err)
				$$ = nil
			} else {
				$$ = f
			}
			}
		;


/* function names */
function_name	: identifier { $$ = $1 } ;


function_args	: function_args COMMA function_arg
			{
			$$ = append($$, $3)
			}
		| function_args COMMA
		| function_arg
			{ $$ = []Node{$1} }
		| /* empty */
			{ $$ = nil }
		;


function_arg	: assignment_stmt
			{ $$ = $1 }
		| array_elem
			{ $$ = $1 }
		| paren_expr
			{ $$ = $1 }
		// | computation_expr
		// 	{ $$ = $1 }
		| attr_expr 
			{ $$ = $1 }
		| LEFT_BRACKET array_list RIGHT_BRACKET
			{ $$ = getFuncArgList($2.(NodeList)) }
		;


assignment_stmt	: expr EQ expr
           		{ $$ = yylex.(*parser).newAssignmentStmt($1, $3) }
		;


conditional_expr	: expr GTE expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr GT expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr AND expr
				{ $$ = yylex.(*parser).newConditionalExpr($1, $3, $2) }
			| expr OR expr
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


/*
computation_expr	: expr ADD expr
				{ $$ = yylex.(*parser).newComputationExpr($1, $3, $2) }
			| expr SUB expr
				{ $$ = yylex.(*parser).newComputationExpr($1, $3, $2) }
			| expr MUL expr
				{ $$ = yylex.(*parser).newComputationExpr($1, $3, $2) }
			| expr DIV expr
				{ $$ = yylex.(*parser).newComputationExpr($1, $3, $2) }
			| expr MOD expr
				{ $$ = yylex.(*parser).newComputationExpr($1, $3, $2) }
			;
*/


paren_expr	: LEFT_PAREN expr RIGHT_PAREN
			{ $$ = &ParenExpr{Param: $2} }
		;


index_expr	: identifier LEFT_BRACKET number_literal RIGHT_BRACKET
			{ $$ = yylex.(*parser).newIndexExpr(&Identifier{Name: $1.Val}, $3) }
		| index_expr LEFT_BRACKET number_literal RIGHT_BRACKET
			{ $$ = yylex.(*parser).newIndexExpr($1, $3) }
		| DOT LEFT_BRACKET number_literal RIGHT_BRACKET
			{ $$ = yylex.(*parser).newIndexExpr(nil, $3) }
		;


attr_expr	: identifier DOT index_expr
			{ $$ = &AttrExpr{Obj: &Identifier{Name: $1.Val}, Attr: $3} }
		| identifier DOT identifier
			{ $$ = &AttrExpr{Obj: &Identifier{Name: $1.Val}, Attr: &Identifier{Name: $3.Val}} }
		| index_expr DOT index_expr
			{ $$ = &AttrExpr{Obj: $1, Attr: $3} }
	  	| index_expr DOT identifier
			{ $$ = &AttrExpr{Obj: $1, Attr: &Identifier{Name: $3.Val}} }
		| attr_expr DOT index_expr
			{ $$ = &AttrExpr{Obj: $1, Attr: $3} }
		| attr_expr DOT identifier
			{ $$ = &AttrExpr{Obj: $1, Attr: &Identifier{Name: $3.Val}} }
		;


regex	: RE LEFT_PAREN string_literal RIGHT_PAREN
		{ $$ = &Regex{Regex: $3.(*StringLiteral).Val} }
	| RE LEFT_PAREN QUOTED_STRING RIGHT_PAREN
		{ $$ = &Regex{Regex: yylex.(*parser).unquoteString($3.Val)} }
	;


array_list	: array_list COMMA array_elem
			{
			nl := $$.(NodeList)
			nl = append(nl, $3)
			$$ = nl
			}
		| array_elem
			{ $$ = NodeList{$1} }
		| /* empty */
			{ $$ = NodeList{} }
		;


array_elem	: bool_literal
		| string_literal
		| nil_literal
		| number_literal
		| identifier
			{ $$ = &Identifier{Name: $1.Val} }
		;


bool_literal	: TRUE
			{ $$ = &BoolLiteral{Val: true} }
		| FALSE
			{ $$ = &BoolLiteral{Val: false} }
		;


string_literal	: STRING
			{ $$ = &StringLiteral{Val: yylex.(*parser).unquoteString($1.Val)} }
		;


nil_literal	: NIL
			{ $$ = &NilLiteral{} }
		| NULL
			{ $$ = &NilLiteral{} }
		;


/* literals */
number_literal	: NUMBER
			{ $$ = yylex.(*parser).number($1.Val) }
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
			$$ = num
			}
		;


identifier	: ID
		| QUOTED_STRING
			{ $$.Val = yylex.(*parser).unquoteString($1.Val) }
		| IDENTIFIER LEFT_PAREN string_literal RIGHT_PAREN
			{ $$.Val = $3.(*StringLiteral).Val }
		;


unary_op	: ADD | SUB ;


%%
