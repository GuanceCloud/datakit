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
%token keywordsStart
%token <item>
AS ASC AUTO BY DESC TRUE FALSE FILTER
IDENTIFIER IN AND LINK LIMIT SLIMIT
OR NIL NULL OFFSET SOFFSET
ORDER RE INT FLOAT POINT TIMEZONE WITH
%token keywordsEnd

// start symbols for parser
%token startSymbolsStart
%token START_STMTS START_BINARY_EXPRESSION START_FUNC_EXPRESSION
%token startSymbolsEnd

////////////////////////////////////////////////////
// grammar rules
////////////////////////////////////////////////////
%type <item>
	unary_op
	lambda_label
	order_by_label
	function_name
	identifier
	namespace

%type<nodes>
  target_list
  lambda_query_list
	filter_list
	where_conditions
	target_clause
	function_args

%type <node>
	stmt
	stmts
	query_stmt
	query_base_stmt
	outerFunc_stmt
	lambda_stmt
	from_clause
	group_by_stmt
	order_by_stmt
	limit_stmt
	offset_stmt
	slimit_stmt
	soffset_stmt
	timezone_stmt

%type <node>
	array_elem
	array_list
	order_by_elem
	order_by_list
	attr_expr
	binary_expr
	expr
	filter_elem
	from_list
	function_arg
	function_expr
	naming_arg
	paren_expr
	regex
	static_cast
	target_elem
	time_expr
	time_range_opt
	time_range_start
	time_range_end
	group_by_interval
	columnref
	bool_literal
	string_literal
	nil_literal
	number_literal
	cascade_functions
	star

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
		 | START_FUNC_EXPRESSION function_expr
		 {
				yylex.(*parser).parseResult = $2
		 }
		 | start EOF
		 | error
		 {
				yylex.(*parser).unexpected("", "")
		 }
		 ;

stmt: query_stmt
		{ $$ = $1 }
		| outerFunc_stmt
	    { $$ = $1 }
		| lambda_stmt
		{ $$ = $1 }
		;

stmts: stmt
		 {
			 $$ = Stmts{$1}
		 }
		 | stmts SEMICOLON stmt
		 {
		 	 arr := $1.(Stmts)
			 arr = append(arr, $3)
			 $$ = arr
		 }
		 | stmts SEMICOLON
		 {
			 $$ = $1
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
						| static_cast
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

outerFunc_stmt: function_expr where_conditions time_range_opt limit_stmt offset_stmt
				 {
				   var cFuns *OuterFuncs
				   chainFuncs, err := yylex.(*parser).newOuterFunc(cFuns, $1.(*FuncExpr))
				   if err != nil {
						yylex.(*parser).addParseErr(nil, err)
				   } else{
						switch chainFuncs.(type){
						case *OuterFuncs:
							$$ = chainFuncs.(*OuterFuncs)
						case *Show:
							show := chainFuncs.(*Show)
							if $2 != nil {
								show.WhereCondition = $2
							}
							if $3 != nil {
								show.TimeRange = $3.(*TimeRange)
							}
							if $4 != nil {
								show.Limit = $4.(*Limit)
							}
							if $5 != nil {
								show.Offset= $5.(*Offset)
							}
							$$ = show
						case *DeleteFunc:
							$$ = chainFuncs.(*DeleteFunc)
						default:
						    yylex.(*parser).addParseErr(nil, fmt.Errorf("outer func error"))
						}
				      }
				 }
				 | outerFunc_stmt DOT function_expr
				 {
					cFuns := $1.(*OuterFuncs)
					chainFuncs, err := yylex.(*parser).newOuterFunc(cFuns, $3.(*FuncExpr))
					if err != nil {
						yylex.(*parser).addParseErr(nil, err)
				    }
				    $$ = chainFuncs.(*OuterFuncs)
				 }
				 ;


lambda_stmt: query_stmt lambda_label lambda_query_list WITH where_conditions
					 {
						 m := yylex.(*parser).newLambda($1.(*DFQuery), $2, $5)
						 for _, n := range $3 {
						 	 m.Right = append(m.Right, n.(*DFQuery))
						 }
						 $$ = m
					 }
					 ;

lambda_label: FILTER
						{
							$$ = Item{Typ: FILTER}
						}
						| LINK
						{
							$$ = Item{Typ: LINK}
						}
						;

lambda_query_list: query_stmt
							   {
							     $$ = []Node{$1}
                 }
							   | lambda_query_list lambda_label query_stmt
		             {
		               $$ = append($1, $3)
		             }
		             ;

query_stmt: query_base_stmt time_range_opt group_by_stmt order_by_stmt limit_stmt offset_stmt slimit_stmt soffset_stmt timezone_stmt
          {
					  m := $1.(*DFQuery)

          	if $2 != nil {
          		m.TimeRange = $2.(*TimeRange)
          	}

          	if $3 != nil {
          		m.GroupBy = $3.(*GroupBy)
          	}

          	if $4 != nil {
          		m.OrderBy = $4.(*OrderBy)
          	}

          	if $5 != nil {
          		m.Limit = $5.(*Limit)
          	}

          	if $6 != nil {
          		m.Offset = $6.(*Offset)
          	}

          	if $7 != nil {
          		m.SLimit = $7.(*SLimit)
          	}

          	if $8 != nil {
          		m.SOffset = $8.(*SOffset)
          	}

          	if $9 != nil {
          		m.TimeZone = $9.(*TimeZone)
          	}

          	$$ = m
          }
				  ;

query_base_stmt: from_clause target_clause where_conditions
               {
                 m := $1.(*DFQuery)
                 m.Targets = yylex.(*parser).newTargets($2)
                 m.WhereCondition = $3
                 $$ = m
               }
               ;

from_clause: from_list
					 {
					 	 $$ = $1
					 }
					 | namespace from_list
					 {
						 q := $2.(*DFQuery)
						 q.Namespace = $1.Val
						 $$ = q
					 }
					 ;

namespace: ID NAMESPACE
				 {
					 $$ = $1
				 }
				 ;

from_list: regex
				 {
					 q, err := yylex.(*parser).newQuery($1)
					 if err != nil  {
					 	 log.Errorf("newQuery: %s", err)
					 }
					 $$ = q
				 }
				 | attr_expr
				 {
				 	 // FIXME: only func:: support attr_expr in from-clause
					 x := $1.(*AttrExpr)
					 q, err := yylex.(*parser).newQuery(&StringLiteral{Val: fmt.Sprintf("%s__%s", x.Obj, x.Attr)})
					 if err != nil  {
					 	 log.Errorf("newQuery: %s", err)
					 }
					 $$ = q
				 }
				 | identifier
				 {
					 q, err := yylex.(*parser).newQuery($1)
					 if err != nil  {
					 	 log.Errorf("newQuery: %s", err)
					 }
					 $$ = q
				 }
				 | LEFT_PAREN query_stmt RIGHT_PAREN
				 {
					 $$ = yylex.(*parser).newSubquery($2.(*DFQuery))
				 }
				 | from_list COMMA regex
				 {
					 q := $1.(*DFQuery)
					 if err := q.appendFrom($3); err != nil {
					 	 log.Debugf("appendFrom: %s", err.Error())
					 }
					 $$ = q
				 }
				 | from_list COMMA identifier
				 {
				 	 q := $1.(*DFQuery)
					 if err := q.appendFrom($3.Val); err != nil {
					   log.Debugf("appendFrom: %s", err.Error())
					 }
					 $$ = q
				 }
				 | from_list COMMA
				 {
					 $$ = $1
				 }
				 ;

target_clause: COLON LEFT_PAREN target_list RIGHT_PAREN
						 {
						   $$ = $3
						 }
						 | /* empty */
						 { $$ = nil }
						 ;

target_list: target_elem
           {
           	 $$ = []Node{$1}
           }
           | target_list COMMA target_elem
           {
             $$ = append($1, $3)
           }
           | target_list COMMA
           | /* empty */
           { $$ = nil }
           ;

target_elem: expr
					 {
						 nl, err := yylex.(*parser).newTarget($1, "")
						 if err != nil {
							 yylex.(*parser).addParseErr(nil, err)
						 }
						 $$ = nl
					 }
					 | static_cast
					 {
						 nl, err := yylex.(*parser).newTarget($1, "")
						 if err != nil {
							 yylex.(*parser).addParseErr(nil, err)
						 }
						 $$ = nl
					 }
           | expr AS identifier
					 {
				     nl, err := yylex.(*parser).newTarget($1, $3.Val)
					   if err != nil {
							 yylex.(*parser).addParseErr(nil, err)
					   }
					   $$ = nl
					 }
           | static_cast AS identifier
					 {
				     nl, err := yylex.(*parser).newTarget($1, $3.Val)
					   if err != nil {
							 yylex.(*parser).addParseErr(nil, err)
					   }
					   $$ = nl
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

/* group by statement */
group_by_stmt: BY array_list
						 {
							 nl := $2.(NodeList)
						 	 if len(nl) == 0 {
							 	 yylex.(*parser).addParseErrf($1.PositionRange(), "group by list empty")
							 }

						   $$ = &GroupBy{List: nl}
						 }
						 | /* empty */
						 { $$ = nil }
						 ;

/* order by statement */
order_by_stmt: ORDER BY order_by_list
						 {
							 $$ = yylex.(*parser).newOrderBy($3.(NodeList))
						 }
						 | /* empty */
						 {
							 $$ = yylex.(*parser).newOrderBy(nil)
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
