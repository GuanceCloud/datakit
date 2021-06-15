%{
package parser

import (
	"time"
	"math"
	"fmt"
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
%token <item>
AS ASC AUTO BY DESC TRUE FALSE FILTER
IDENTIFIER IN AND LINK LIMIT SLIMIT
OR NIL NULL OFFSET SOFFSET
ORDER RE INT FLOAT POINT TIMEZONE WITH

// start symbols for parser
%token START_STMTS START_BINARY_EXPRESSION START_FUNC_EXPRESSION

////////////////////////////////////////////////////
// grammar rules
////////////////////////////////////////////////////
%type <item>
	order_by_label
	identifier

%type<nodes>
	filter_list
	where_conditions

%type <node>
	group_by_stmt
	order_by_stmt
	limit_stmt
	offset_stmt
	slimit_stmt
	soffset_stmt
	timezone_stmt

%type <node>
	order_by_elem
	order_by_list
	binary_expr
	expr
	filter_elem
	regex
	static_cast
	time_expr
	time_range_opt
	time_range_start
	time_range_end
	group_by_interval
	number_literal

%type <duration> duration time_range_offset
%type <timestamp> timestamp date_literal

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

start: START_STMTS where_conditions
		 {
				yylex.(*parser).parseResult = $2
		 }
		 | START_BINARY_EXPRESSION binary_expr
		 {
				yylex.(*parser).parseResult = $2
		 }
		 | START_FUNC_EXPRESSION
		 {
				yylex.(*parser).parseResult = $2
		 }
		 | start EOF
		 | error
		 {
				yylex.(*parser).unexpected("", "")
		 }
		 ;

where_conditions: LEFT_BRACE filter_list RIGHT_BRACE
						 {
						   $$ = yylex.(*parser).newWhereConditions($2)
						 }
						 | /* empty */
						 {
						   $$ = yylex.(*parser).newWhereConditions(nil)
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

filter_elem: binary_expr
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
					 ;

time_range_opt: LEFT_BRACKET time_range_start time_range_end group_by_interval time_range_offset RIGHT_BRACKET
							{
								$$ = yylex.(*parser).newTimeRangeOpt($2, $3, $4, $5)
							}
							| /* empty */
							{
								$$ = yylex.(*parser).newTimeRangeOpt(nil, nil,nil, time.Duration(0))
							}
							;

time_range_start: time_expr
								{ $$ = $1}
								| /*empty*/
								{ $$ = nil }
								;

time_range_end: COLON time_expr
								{ $$ = $2}
								| COLON
								{ $$ = nil }
								| /*empty*/
								{ $$ = nil }
								;

time_range_offset: COMMA duration
								 { $$ = $2 }
								 | COMMA
								 { $$ = time.Duration(0) }
								 | /* empty */
								 { $$ = time.Duration(0) }
								 ;

group_by_interval: COLON duration
                 {
                 	 $$ = &TimeResolution{Duration: $2}
                 }
                 | COLON AUTO
                 {
                   $$ = yylex.(*parser).newTimeResolution(nil, true) // Deprecated
                 }
                 | COLON AUTO LEFT_PAREN RIGHT_PAREN
                 {
                   $$ = yylex.(*parser).newTimeResolution(nil, false)
                 }
                 | COLON AUTO LEFT_PAREN number_literal RIGHT_PAREN
                 {
                   $$ = yylex.(*parser).newTimeResolution($4.(*NumberLiteral), false)
                 }
                 | COLON
                 { $$ = nil }
                 | /* empty */
                 { $$ = nil }
                 ;

time_expr: duration
				 {
				 	 $$ = yylex.(*parser).newTimeExpr(&TimeExpr{IsDuration: true, Duration: $1})
				 }
				 | timestamp
				 {
				 	 $$ = yylex.(*parser).newTimeExpr(&TimeExpr{Time: $1})
				 }
				 ;

timezone_stmt: TIMEZONE LEFT_PAREN string_literal RIGHT_PAREN
             {
							 $$ = yylex.(*parser).newTimeZone($3.(*StringLiteral))
						 }
             | /* empty */
             { $$ = nil }
             ;

offset_stmt: OFFSET number_literal
           {
						 $$ = yylex.(*parser).newOffset($2.(*NumberLiteral))
				   }
           | /* empty */
           {
             $$ = yylex.(*parser).newOffset(nil)
           }
           ;

soffset_stmt: SOFFSET number_literal
          {
						$$ = yylex.(*parser).newSOffset($2.(*NumberLiteral))
					}
					| /* empty */
          {
            $$ = yylex.(*parser).newSOffset(nil)
          }
          ;

limit_stmt: LIMIT number_literal
          {
						$$ = yylex.(*parser).newLimit($2.(*NumberLiteral))
				  }
          | /* empty */
          {
					  $$ = yylex.(*parser).newLimit(nil)
					}
          ;

slimit_stmt: SLIMIT number_literal
           {
						 $$ = yylex.(*parser).newSLimit($2.(*NumberLiteral))
					 }
           | /* empty */
           {
						 $$ = yylex.(*parser).newSLimit(nil)
					 }
           ;

order_by_list: order_by_list COMMA order_by_elem
             {
               nl := $1.(NodeList)
               nl = append(nl, $3)
               $$ = nl
             }
             | order_by_elem
             {
             	 $$ = NodeList{$1}
             }
					   | /* empty */
					   {
					   	 $$ = NodeList{}
					   }
             ;

order_by_elem: identifier order_by_label
						 {
							 $$ = yylex.(*parser).newOrderByElem($1.Val, $2)
						 }
						 ;

order_by_label: ASC | DESC
							|/* empty */
							{ $$ = Item{Typ: ASC} }
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

date_literal: NUMBER SUB NUMBER SUB NUMBER NUMBER COLON NUMBER COLON NUMBER
						{
							timestr := fmt.Sprintf("%s-%02s-%02s %02s:%02s:%02s", $1.Val, $3.Val, $5.Val, $6.Val, $8.Val, $10.Val)
							t, err := time.ParseInLocation("2006-01-02 15:04:05", timestr, time.UTC)
							if err != nil {
								yylex.(*parser).addParseErrf($1.PositionRange(), "invalid date string: %s", timestr)
							}

							$$ = t
						}
						| NUMBER SUB NUMBER SUB NUMBER
						{
							timestr := fmt.Sprintf("%s-%02s-%02s", $1.Val, $3.Val, $5.Val)
							t, err := time.ParseInLocation("2006-01-02", timestr, time.UTC)
							if err != nil {
								yylex.(*parser).addParseErrf($1.PositionRange(), "invalid date string: %s", timestr)
							}
							$$ = t
						}
						;

duration: DURATION
				{
				  du, err := yylex.(*parser).parseDuration($1.Val)
					if err != nil {
								yylex.(*parser).addParseErr($1.PositionRange(), err)
					} else {
						$$ = du
					}
				}
				| unary_op DURATION
				{
					du, err := yylex.(*parser).parseDuration($2.Val)
					if err != nil {
								yylex.(*parser).addParseErr($2.PositionRange(), err)
					} else {
						switch $1.Typ {
						case ADD:
							$$ = du
						case SUB:
							$$ = -du
						}
					}
				}
				;

timestamp: date_literal
				 | number_literal
				 {
				  	nl := $1.(*NumberLiteral)
				  	var t time.Time
				  	if nl.IsInt {
				  		t = time.Unix(nl.Int, 0)
				  	} else {
				  		i, f := math.Modf(nl.Float)
				  		t = time.Unix(int64(i), int64(f * float64(time.Second)))
				  	}
				  	$$ = t
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

static_cast: INT LEFT_PAREN identifier RIGHT_PAREN
           {
             $$ = &StaticCast{IsInt: true, Val: &Identifier{Name: $3.Val}}
           }
           | FLOAT LEFT_PAREN identifier RIGHT_PAREN
           {
             $$ = &StaticCast{IsFloat: true, Val: &Identifier{Name: $3.Val}}
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
