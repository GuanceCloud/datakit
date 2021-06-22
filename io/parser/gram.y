%{
package parser

import (
	"time"
)
%}

%union {
	node       Node
	nodes      []Node

	item       Item

	strings    []string
	float      float64
	duration   time.Duration
	timestamp  time.Time
}

%token <item> EQ COLON SEMICOLON COMMA COMMENT DURATION
	EOF ERROR ID LEFT_BRACE LEFT_BRACKET
	LEFT_PAREN NUMBER RIGHT_BRACE RIGHT_BRACKET
	RIGHT_PAREN SPACE STRING QUOTED_STRING NAMESPACE
	DOT

// operator
%token operatorsStart
%token <item> ADD
	DIV GTE GT
	LT LTE MOD MUL
	NEQ POW SUB
%token operatorsEnd

// keywords
%token keywordsStart
%token <item>
AS ASC AUTO BY DESC TRUE FALSE FILTER
IDENTIFIER IN NOTIN AND LINK LIMIT SLIMIT
OR NIL NULL OFFSET SOFFSET
ORDER RE INT FLOAT POINT TIMEZONE WITH
%token keywordsEnd

// start symbols for parser
%token startSymbolsStart
%token START_STMTS START_BINARY_EXPRESSION START_FUNC_EXPRESSION START_WHERE_CONDITION
%token startSymbolsEnd

////////////////////////////////////////////////////
// grammar rules
////////////////////////////////////////////////////
%type <item>
	unary_op
	function_name
	identifier

%type<nodes>
	filter_list
	function_args

%type <node>
	stmts
	where_conditions
	array_elem
	array_list
	attr_expr
	binary_expr
	expr
	function_arg
	function_expr
	naming_arg
	paren_expr
	filter_elem
	regex
	columnref
	bool_literal
	string_literal
	nil_literal
	number_literal
	cascade_functions
	star

%start start

// operator listed with increasing precedence
%left OR
%left AND
%left GTE GT NEQ EQ LTE LT
%left ADD SUB
%left MUL DIV MOD
%right POW

// `offset` do not have associativity
%nonassoc OFFSET
%right LEFT_BRACKET

%%

start: START_WHERE_CONDITION stmts
		 {
				yylex.(*parser).parseResult = $2
		 }
		 | start EOF
		 | error
		 {
				yylex.(*parser).unexpected("", "")
		 }
		 ;

stmts: where_conditions
		 {
		 $$ = WhereConditions{$1}
		 }
		 | stmts SEMICOLON where_conditions
		 {
			 if $3 != nil {
				arr := $1.(WhereConditions)
				arr = append(arr, $3)
				$$ = arr
			 } else {
			 	$$ = $1
			 }
		 }
		 ;

/* expression */
expr: array_elem | regex | paren_expr | function_expr | binary_expr | cascade_functions
		;

columnref: identifier
				 {
				   $$ = &Identifier{Name: $1.Val}
				 }
				 | attr_expr
				 {
					 $$ = $1
				 }
				 ;

attr_expr: identifier DOT identifier
				 {
				 	 $$ = &AttrExpr{
					 	 Obj: &Identifier{Name: $1.Val},
					 	 Attr: &Identifier{Name: $3.Val},
					 }
				 }
				 | attr_expr DOT identifier
				 {
				 	 $$ = &AttrExpr{
						 Obj: $1.(*AttrExpr),
						 Attr: &Identifier{Name: $3.Val},
					 }
				 }
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

cascade_functions: function_expr DOT function_expr
								 {
								 	$$ = &CascadeFunctions{Funcs: []*FuncExpr{$1.(*FuncExpr), $3.(*FuncExpr)}}
								 }
								 | cascade_functions DOT function_expr
								 {
								 	fc := $1.(*CascadeFunctions)
									fc.Funcs = append(fc.Funcs, $3.(*FuncExpr))
									$$ = fc
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
					| columnref
					| nil_literal
					| bool_literal
					| star
					;

star : MUL
		 {
		 		$$ = &Star{}
		 }
		 ;

function_arg: naming_arg
						| expr
						| LEFT_BRACKET array_list RIGHT_BRACKET
						{
							$$ = getFuncArgList($2.(NodeList))
						}
						;

naming_arg: identifier EQ expr
						{
							$$ = &FuncArg{ArgName: $1.Val, ArgVal: $3}
						}
						| identifier EQ LEFT_BRACKET array_list RIGHT_BRACKET
						{
							$$ = &FuncArg{
								ArgName: $1.Val,
								ArgVal: getFuncArgList($4.(NodeList)),
							}
						}
						;

where_conditions: LEFT_BRACE filter_list RIGHT_BRACE
						 {
						   $$ = yylex.(*parser).newWhereConditions($2)
						 }
						 | /* empty */
						 {
						   $$ = nil
						 }
						 ;

/* filter list */
filter_list: filter_elem
					 {
					  	$$ = []Node{ $1 }
					 }
					 | filter_list COMMA filter_elem
					 {
					  	$$ = append($$, $3)
					 }
					 | filter_list COMMA
					 | /* empty */
					 { $$ = nil }
					 ;

filter_elem: binary_expr | paren_expr
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
					 | columnref IN LEFT_BRACKET array_list RIGHT_BRACKET
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $4, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 | columnref NOTIN LEFT_BRACKET array_list RIGHT_BRACKET
					 {
						 bexpr := yylex.(*parser).newBinExpr($1, $4, $2)
						 bexpr.ReturnBool = true
						 $$ = bexpr
					 }
					 ;

/* function names */
function_name: identifier
						 {
						 	$$ = $1
						 }
						 | attr_expr
						 {
						 	$$ = Item{Val: $1.(*AttrExpr).String()}
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
%%
