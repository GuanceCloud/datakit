package parselogical

import (
	"strings"
)

import (
	"github.com/pkg/errors"
)

const (
	parseStateInitial int = iota // start state to determine the message type

	// prelude section
	parseStateRelation          // <schema>.<table>
	parseStateOperation         // INSERT/UPDATE/DELETE
	parseStateEscapedIdentifier // applicable to both the column names, and the relation for when "" is used

	// name and type of the column values
	parseStateColumnName
	parseStateColumnType
	parseStateOpenSquareBracket

	// value parsing
	parseStateColumnValue
	parseStateColumnQuotedValue

	// terminal states
	parseStateEnd
	parseStateNull
)

// ColumnValue is an annotated String type, containing the Postgres type of the String,
// and whether it was quoted within the original parsing.
// Quoting is useful for handling values like val[text]:null or val[text]:unchanged-toast-datum,
// which are special signifiers (denoting null and unchanged data, respectively).
type ColumnValue struct {
	Value  string
	Type   string
	Quoted bool
}

type parseState struct {
	msg *string

	current       int
	prev          int
	tokenStart    int
	oldKey        bool
	curColumnName string
	curColumnType string
}

// ParseResult is the result of parsing the string
type ParseResult struct {
	state parseState // for internal use

	Transaction string                 // filled if this is a transaction statement (BEGIN/COMMIT) with the value of the transaction
	Relation    string                 // filled if this is a DML statement with <schema>.<table>
	Operation   string                 // filled if this is a DML statement with the name of the operation (INSERT/UPDATE/DELETE)
	NoTupleData bool                   // true if this had no actual tuple data
	Columns     map[string]ColumnValue // if a DML statement, the fields affected by the operation with their values and types
	OldColumns  map[string]ColumnValue // if an UPDATE and REPLICA IDENTITY setting (FULL or USING INDEX both will add values here)
}

// NewParseResult creates the structure that is filled by the Parse operation
func NewParseResult(msg string) *ParseResult {
	pr := new(ParseResult)
	pr.state = parseState{msg: &msg, current: parseStateInitial, prev: parseStateInitial, tokenStart: 0, oldKey: false}
	pr.Columns = make(map[string]ColumnValue)
	pr.OldColumns = make(map[string]ColumnValue)
	return pr
}

// Parse parses a string produced from the test_decoding logical replication plugin into the ParseResult struct, above
func (pr *ParseResult) Parse() error {
	err := pr.parse(false)
	if err != nil {
		return err
	}
	return nil
}

// ParsePrelude allows more control over the parsing process
// if you want to avoid parsing the string for tables/schemas (via a filtering mechansim)
// you can run ParsePrelude to just fill the operation type, the schema and the table
func (pr *ParseResult) ParsePrelude() error {
	return pr.parse(true)
}

// ParseColumns also does some cool shit
func (pr *ParseResult) ParseColumns() error {
	return pr.parse(false)
}

// based on https://github.com/citusdata/pg_warp/blob/9e69814b85e18fbe5a3c89f0e17d22583c9a398c/consumer/consumer.go
// as opposed to a hackier earlier version
func (pr *ParseResult) parse(preludeOnly bool) error {
	state := pr.state
	message := *state.msg

	if state.current == parseStateInitial {
		if len(message) < 5 {
			return errors.Errorf("message too short: %s", message)
		}

		switch message[0:5] {
		case "BEGIN":
			fallthrough
		case "COMMI":
			fields := strings.Fields(message)

			if len(fields) != 2 {
				return errors.Errorf("unknown transaction message: %s", message)
			}

			pr.Operation = fields[0]
			pr.Transaction = fields[1]

			return nil
		case "table":
			// actual DML statement for which we need to parse out the table/operation
		default:
			return errors.Errorf("unknown logical message received: %s", message)
		}

		// we are parsing a table statement, so let's skip over the initial "table "
		state.tokenStart = 6
		state.current = parseStateRelation
	}

outer:
	for i := 0; i <= len(message); i++ {
		if i < state.tokenStart {
			i = state.tokenStart - 1
			continue
		}

		chr := byte('\000')
		if i < len(message) {
			chr = message[i]
		}
		chrNext := byte('\000')
		if i+1 < len(message) {
			chrNext = message[i+1]
		}

		switch state.current {
		case parseStateNull:
			return errors.Errorf("invalid parse state null: %+v", state)
		case parseStateRelation:
			if chr == ':' {
				if chrNext != ' ' {
					return errors.Errorf("invalid character ' ' at %d", i+1)
				}
				pr.Relation = message[state.tokenStart:i]
				state.tokenStart = i + 2
				state.current = parseStateOperation
			} else if chr == '"' {
				state.prev = state.current
				state.current = parseStateEscapedIdentifier
			}
		case parseStateOperation:
			if chr == ':' {
				if chrNext != ' ' {
					return errors.Errorf("invalid character ' ' at %d", i+1)
				}
				pr.Operation = message[state.tokenStart:i]
				state.tokenStart = i + 2
				state.current = parseStateColumnName

				if preludeOnly {
					break outer
				}
			}
		case parseStateColumnName:
			if chr == '[' {
				state.curColumnName = message[state.tokenStart:i]
				state.tokenStart = i + 1
				state.current = parseStateColumnType
			} else if chr == ':' {
				if message[state.tokenStart:i] == "old-key" {
					state.oldKey = true
				} else if message[state.tokenStart:i] == "new-tuple" {
					state.oldKey = false
				}
				state.tokenStart = i + 2
			} else if chr == '(' && message[state.tokenStart:len(message)] == "(no-tuple-data)" {
				pr.NoTupleData = true
				state.current = parseStateEnd
			} else if chr == '"' {
				state.prev = state.current
				state.current = parseStateEscapedIdentifier
			}
		case parseStateColumnType:
			if chr == ']' {
				if chrNext != ':' {
					return errors.Errorf("invalid character '%s' at %d", []byte{chrNext}, i+1)
				}
				state.curColumnType = message[state.tokenStart:i]
				state.tokenStart = i + 2
				state.current = parseStateColumnValue
			} else if chr == '"' {
				state.prev = state.current
				state.current = parseStateEscapedIdentifier
			} else if chr == '[' {
				state.prev = state.current
				state.current = parseStateOpenSquareBracket
			}
		case parseStateColumnValue:
			if chr == '\000' || chr == ' ' {
				quoted := state.prev == parseStateColumnQuotedValue
				startStr := state.tokenStart
				endStr := i

				if quoted {
					startStr++
					endStr--
				}

				unescapedValue := strings.Replace(message[startStr:endStr], "''", "'", -1) // the value is escape-quoted to us

				cv := ColumnValue{Value: unescapedValue, Quoted: quoted, Type: state.curColumnType}

				if state.oldKey {
					pr.OldColumns[state.curColumnName] = cv
				} else {
					pr.Columns[state.curColumnName] = cv
				}
			}

			if chr == '\000' {
				state.current = parseStateEnd
			} else if chr == ' ' {
				state.tokenStart = i + 1
				state.prev = state.current
				state.current = parseStateColumnName
			} else if chr == '\'' {
				state.prev = state.current
				state.current = parseStateColumnQuotedValue
			}
		case parseStateOpenSquareBracket:
			if chr == ']' {
				state.current = state.prev
				state.prev = parseStateNull
			}
		case parseStateEscapedIdentifier:
			if chr == '"' {
				if chrNext == '"' {
					i++
				} else {
					state.current = state.prev
					state.prev = parseStateNull
				}
			}
		case parseStateColumnQuotedValue:
			if chr == '\'' {
				if chrNext == '\'' {
					i++
				} else {
					prev := state.prev
					state.prev = state.current
					state.current = prev
				}
			}
		}
	}

	if (preludeOnly && state.current != parseStateColumnName) || (!preludeOnly && state.current != parseStateEnd) {
		return errors.Errorf("invalid parser end state: %+v", state.current)
	}

	return nil
}
