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
	strings    []string
	float      float64
}

%token <item> SEMICOLON COMMA COMMENT DOT
	EOF ERROR ID LEFT_PAREN LEFT_BRACKET NUMBER
	RIGHT_PAREN RIGHT_BRACKET SPACE STRING QUOTED_STRING

// operator
%token operatorsStart
%token <item> ADD
	DIV GTE GT
	LT LTE MOD MUL
	NEQ EQ POW SUB
%token operatorsEnd

// keywords
%token keywordsStart
%token <item>
TRUE FALSE IDENTIFIER
AND OR NIL NULL RE JP
%token keywordsEnd


// start symbols for parser
%token startSymbolsStart
%token START_PIPELINE
%token startSymbolsEnd

////////////////////////////////////////////////////
// grammar rules
////////////////////////////////////////////////////
%type <item>
	unary_op
	function_name
	identifier

%type<nodes> function_args

%type <node>
	array_elem
	array_list
	binary_expr
	expr
	function_arg
	function_expr
	paren_expr
	regex
	bool_literal
	string_literal
	nil_literal
	number_literal
	//columnref
	index_expr
	attr_expr
	pipeline

%start start

// operator listed with increasing precedence
%left OR
%left AND
%left GTE GT NEQ EQ LTE LT
%left ADD SUB
%left MUL DIV MOD
%right POW

%%

start: START_PIPELINE pipeline
		 {
				yylex.(*parser).parseResult = $2
		 }
		 | start EOF
		 | error
		 {
				yylex.(*parser).unexpected("", "")
		 }
		 ;

pipeline: function_expr
		 {
		 	$$ = &Ast{ Functions: []*FuncExpr{$1.(*FuncExpr)} }
		 }
		 | pipeline function_expr
		 {
		 	ast := $1.(*Ast)
			ast.Functions = append(ast.Functions, $2.(*FuncExpr))
			$$ = $1
		 }
		 ;

/* expression */
expr:  array_elem | regex | paren_expr | function_expr | binary_expr | attr_expr
		;

unary_op: ADD
				| SUB
				;

string_literal: STRING
							{
							  $$ = &StringLiteral{Val: yylex.(*parser).unquoteString($1.Val)}
							}
							;

nil_literal: NIL
					 {
					 	 $$ = &NilLiteral{}
					 }
					 | NULL
					 {
					 	 $$ = &NilLiteral{}
					 }
					 ;

bool_literal: TRUE
						{
							$$ = &BoolLiteral{Val: true}
						}
						| FALSE
						{
							$$ = &BoolLiteral{Val: false}
						}
						;

paren_expr: LEFT_PAREN expr RIGHT_PAREN
					{
						$$ = &ParenExpr{Param: $2}
					}
					;

function_expr: function_name LEFT_PAREN function_args RIGHT_PAREN
							{
								$$ = yylex.(*parser).newFunc($1.Val, $3)
							}
							;

function_args: function_args COMMA function_arg
						 {
						   $$ = append($$, $3)
						 }
						 | function_args COMMA
						 | function_arg
						 {
						 	 $$ = []Node{$1}
						 }
						 | /* empty */
						 {
						 	 $$ = nil
						 }
						 ;


function_arg: expr
						{
							$$ = $1
						}
						| LEFT_BRACKET array_list RIGHT_BRACKET
                        {
                        	$$ = getFuncArgList($2.(NodeList))
                        }
						;

array_list: array_list COMMA array_elem
					{
						nl := $$.(NodeList)
						nl = append(nl, $3)
						$$ = nl
					}
					| array_elem
					{
						$$ = NodeList{$1}
					}
					| /* empty */
					{
						$$ = NodeList{}
					}
					;

array_elem: number_literal
					| string_literal
					//| columnref
					| nil_literal
					| bool_literal
					;

binary_expr: expr ADD expr
					 {
					   $$ = yylex.(*parser).newBinExpr($1, $3, $2)
					 }
					 | expr DIV expr
					 {
					   $$ = yylex.(*parser).newBinExpr($1, $3, $2)
					 }
					 | expr GTE expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 | expr GT expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 | expr AND expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 | expr OR expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 | expr LT expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 | expr LTE expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 | expr MOD expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 $$ = bexpr
					 }
					 | expr MUL expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 $$ = bexpr
					 }
					 | expr NEQ expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 | expr POW expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 $$ = bexpr
					 }
					 | expr SUB expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 $$ = bexpr
					 }
					 | expr EQ expr
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $3, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 ;


/* function names */
function_name: identifier
						 {
						 	$$ = $1
						 }
						;

/* literals */
number_literal: NUMBER
							{
								$$ = yylex.(*parser).number($1.Val)
							}
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

regex: RE LEFT_PAREN string_literal RIGHT_PAREN
          		 {
          		   $$ = &Regex{Regex: $3.(*StringLiteral).Val}
          		 }
          		 | RE LEFT_PAREN QUOTED_STRING RIGHT_PAREN
          		 {
          		   $$ = &Regex{Regex: yylex.(*parser).unquoteString($3.Val)}
          		 }
          		 ;

identifier: ID
					| QUOTED_STRING
					{
						$$.Val = yylex.(*parser).unquoteString($1.Val)
					}
					| IDENTIFIER LEFT_PAREN string_literal RIGHT_PAREN
					{
						$$.Val = $3.(*StringLiteral).Val
					}
					;

index_expr: identifier LEFT_BRACKET number_literal RIGHT_BRACKET
					{
						nl := $3.(*NumberLiteral)
						if !nl.IsInt {
							yylex.(*parser).addParseErr(nil,
								fmt.Errorf("array index should be int, got `%f'", nl.Float))
							$$ = nil
						} else {
							$$ = &IndexExpr{Obj: &Identifier{Name: $1.Val}, Index: []int64{nl.Int}}
						}
					}
					| index_expr LEFT_BRACKET number_literal RIGHT_BRACKET
					{

						nl := $3.(*NumberLiteral)
						if !nl.IsInt {
							yylex.(*parser).addParseErr(nil,
								fmt.Errorf("array index should be int, got `%f'", nl.Float))
							$$ = nil
						} else {
							in := $1.(*IndexExpr)
							in.Index = append(in.Index, nl.Int)
							$$ = in
						}
					}
					| LEFT_BRACKET number_literal RIGHT_BRACKET
					{
						nl := $2.(*NumberLiteral)
						if !nl.IsInt {
							yylex.(*parser).addParseErr(nil,
								fmt.Errorf("array index should be int, got `%f'", nl.Float))
							$$ = nil
						} else {
							$$ = &IndexExpr{ Index: []int64{nl.Int}}
						}
					}
					;

attr_expr: identifier
				 {
				 	$$ = &AttrExpr{Obj: &Identifier{Name: $1.Val}}
				 }
				 | index_expr
				 {
				 	$$ = &AttrExpr{Obj: $1}
				 }
				 | attr_expr DOT identifier
				 {
				 	$$ = &AttrExpr{Obj: $1, Attr: &Identifier{Name: $3.Val}}
				 }
				 | attr_expr DOT index_expr
				 {
				 	$$ = &AttrExpr{Obj: $1, Attr: $3}
				 }
%%
