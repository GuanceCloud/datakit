%{
package parser

%}

%union {
	node       Node
	nodes      []Node
	item       Item
	strings    []string
	float      float64
}

%token <item> SEMICOLON COMMA COMMENT
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
%token START_FUNC_EXPRESSION
%token startSymbolsEnd

////////////////////////////////////////////////////
// grammar rules
////////////////////////////////////////////////////
%type <item> 
	unary_op
	function_name
	identifier

%type<nodes> function_args  function_exprs

%type <node>
	array_elem
	array_list
	binary_expr
	expr
	function_arg
	function_expr
	paren_expr
	regex
	jpath
	bool_literal
	string_literal
	nil_literal
	number_literal
	columnref



%start start

// operator listed with increasing precedence
%left OR
%left AND
%left GTE GT NEQ EQ LTE LT 
%left ADD SUB
%left MUL DIV MOD
%right POW

%%

start: START_FUNC_EXPRESSION function_exprs
		 {
				yylex.(*parser).parseResult = $2
		 }
		 | start EOF
		 | error
		 {
				yylex.(*parser).unexpected("", "")
		 }
		 ;

function_exprs: function_expr
		 {
			 $$ = Funcs{$1}
		 }
		 | function_exprs SEMICOLON function_expr
		 {
             $1 = append($1, $3)
			 $$ = $1
		 }
		 | function_exprs SEMICOLON
		 {
			 $$ = $1
		 }
		 ;

/* expression */
expr:  array_elem | regex | jpath | paren_expr | function_expr | binary_expr
		;

columnref: identifier
				 {
				   $$ = &Identifier{Name: $1.Val}
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
					| columnref
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

jpath: JP LEFT_PAREN string_literal RIGHT_PAREN
          		 {
          		   $$ = &Jspath{Jspath: $3.(*StringLiteral).Val}
          		 }
          		 | JP LEFT_PAREN QUOTED_STRING RIGHT_PAREN
          		 {
          		   $$ = &Jspath{Jspath: yylex.(*parser).unquoteString($3.Val)}
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
