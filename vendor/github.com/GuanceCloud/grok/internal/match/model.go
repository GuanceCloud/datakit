package match

type StructuredIRInfo struct {
	MinWidth          int
	Nullable          bool
	FirstLiteral      string
	FirstLiteralExact bool
	LastLiteral       string
	LastLiteralExact  bool
}

type ASCIICharClass struct {
	Table [256]bool
}

type StructuredKind uint8

const (
	StructuredWord StructuredKind = iota
	StructuredNotSpace
	StructuredHostName
	StructuredIPOrHost
	StructuredNumber
	StructuredInt
	StructuredPosInt
	StructuredNonNegInt
	StructuredCharClass
	StructuredMonthName
	StructuredDayName
	StructuredTimeOfDay
	StructuredYear
	StructuredMonthNum
	StructuredMonthDay
	StructuredHour
	StructuredMinute
	StructuredSecond
	StructuredQuoted
	StructuredUntilLiteral
	StructuredGreedyUntilLiteral
	StructuredSpaceOne
	StructuredSpacePlus
	StructuredSpaceStar
	StructuredTimestampISO8601
	StructuredURIPath
	StructuredURIPathParam
	StructuredHTTPDate
	StructuredLogLevel
)
