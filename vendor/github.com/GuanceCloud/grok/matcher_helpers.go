package grok

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var monthNameValues = [...]string{
	"Jan", "January",
	"Feb", "February",
	"Mar", "March",
	"Apr", "April",
	"May",
	"Jun", "June",
	"Jul", "July",
	"Aug", "August",
	"Sep", "September",
	"Oct", "October",
	"Nov", "November",
	"Dec", "December",
}

var dayNameValues = [...]string{
	"Mon", "Monday",
	"Tue", "Tuesday",
	"Wed", "Wednesday",
	"Thu", "Thursday",
	"Fri", "Friday",
	"Sat", "Saturday",
	"Sun", "Sunday",
}

var logLevelValues = [...]string{
	"alert", "trace", "debug", "notice", "info", "warn", "warning",
	"war", "err", "er", "error", "crit", "cri", "critical", "fatal", "severe", "emerg", "emergency",
}

func isApacheWord(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if isWordByte(s[i]) {
			continue
		}
		return false
	}
	return true
}

func isApacheNumber(s string) bool {
	if len(s) == 0 {
		return false
	}

	i := 0
	signed := false
	if s[0] == '+' || s[0] == '-' {
		signed = true
		i++
		if i == len(s) {
			return false
		}
	}
	if s[i] == '.' {
		if signed {
			return false
		}
		i++
		if i == len(s) {
			return false
		}
		for ; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return false
			}
		}
		return true
	}

	digits := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		digits++
		i++
	}
	if digits == 0 {
		return false
	}
	if i == len(s) {
		return true
	}
	if s[i] != '.' {
		return false
	}
	i++
	if i == len(s) || s[i] < '0' || s[i] > '9' {
		return false
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isApacheInt(s string) bool {
	if len(s) == 0 {
		return false
	}
	i := 0
	if s[0] == '+' || s[0] == '-' {
		i++
	}
	if i == len(s) {
		return false
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isApachePosInt(s string) bool {
	if len(s) == 0 || s[0] < '1' || s[0] > '9' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isApacheNonNegInt(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func maybeTrim(s string, trim bool) string {
	if !trim {
		return s
	}
	return trimMatch(s)
}

func consumeTimestampISO8601(s string, start int) (int, bool) {
	i := start
	yearStart := i
	if !consumeOneOrTwoYearChunks(s, &i) || i >= len(s) || s[i] != '-' {
		return 0, false
	}
	i++
	if !consumeTwoDigitRange(s, &i, 1, 12, true) || i >= len(s) || s[i] != '-' {
		return 0, false
	}
	i++
	if !consumeTwoDigitRange(s, &i, 1, 31, true) || i >= len(s) || (s[i] != 'T' && s[i] != ' ') {
		return 0, false
	}
	i++
	if !consumeTwoDigitRange(s, &i, 0, 23, true) {
		return 0, false
	}
	if i < len(s) && s[i] == ':' {
		i++
	}
	if !consumeTwoDigitRange(s, &i, 0, 59, false) {
		return 0, false
	}
	if i < len(s) && (s[i] == ':' || isASCIIDigit(s[i])) {
		if s[i] == ':' {
			i++
		}
		if !consumeSecondValue(s, &i) {
			return 0, false
		}
	}

	if i < len(s) {
		switch s[i] {
		case 'Z':
			i++
		case '+', '-':
			i++
			if !consumeTwoDigitRange(s, &i, 0, 23, true) {
				return 0, false
			}
			if i < len(s) && s[i] == ':' {
				i++
			}
			if !consumeTwoDigitRange(s, &i, 0, 59, false) {
				return 0, false
			}
		}
	}

	if i == yearStart {
		return 0, false
	}
	return i, true
}

func consumeHTTPDate(s string, start int) (int, bool) {
	if next, ok := consumeCanonicalHTTPDate(s, start); ok {
		return next, true
	}

	i := start
	if !consumeTwoDigitRange(s, &i, 1, 31, true) || i >= len(s) || s[i] != '/' {
		return 0, false
	}
	i++
	monthStart := i
	for i < len(s) {
		r := s[i]
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r >= utf8.RuneSelf {
			i++
			continue
		}
		break
	}
	if i == monthStart || !isMonthNameValue(s[monthStart:i]) || i >= len(s) || s[i] != '/' {
		return 0, false
	}
	i++
	if !consumeOneOrTwoYearChunks(s, &i) || i >= len(s) || s[i] != ':' {
		return 0, false
	}
	i++
	next, ok := consumeTimeOfDay(s, i)
	if !ok {
		return 0, false
	}
	i = next
	if i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++
	if !consumeSignedInt(s, &i) {
		return 0, false
	}
	return i, true
}

func consumeCanonicalHTTPDate(s string, start int) (int, bool) {
	i := start
	if i+26 > len(s) {
		return 0, false
	}

	if !isASCIIDigit(s[i]) || !isASCIIDigit(s[i+1]) || s[i+2] != '/' {
		return 0, false
	}
	day := int(s[i]-'0')*10 + int(s[i+1]-'0')
	if day < 1 || day > 31 {
		return 0, false
	}
	i += 3

	if !isCanonicalMonthAbbrev(s[i:i+3]) || s[i+3] != '/' {
		return 0, false
	}
	i += 4

	if !consumeNDigits(s, &i, 4) || i >= len(s) || s[i] != ':' {
		return 0, false
	}
	i++

	if !consumeCanonicalTwoDigitRange(s, &i, 0, 23) || i >= len(s) || s[i] != ':' {
		return 0, false
	}
	i++
	if !consumeCanonicalTwoDigitRange(s, &i, 0, 59) || i >= len(s) || s[i] != ':' {
		return 0, false
	}
	i++
	if !consumeCanonicalTwoDigitRange(s, &i, 0, 60) || i >= len(s) || s[i] != ' ' {
		return 0, false
	}
	i++

	if i+5 > len(s) || (s[i] != '+' && s[i] != '-') {
		return 0, false
	}
	i++
	if !consumeNDigits(s, &i, 4) {
		return 0, false
	}
	return i, true
}

func consumeCanonicalTwoDigitRange(s string, i *int, min int, max int) bool {
	if *i+2 > len(s) {
		return false
	}
	a := s[*i]
	b := s[*i+1]
	if !isASCIIDigit(a) || !isASCIIDigit(b) {
		return false
	}
	v := int(a-'0')*10 + int(b-'0')
	if v < min || v > max {
		return false
	}
	*i += 2
	return true
}

func isCanonicalMonthAbbrev(s string) bool {
	if len(s) != 3 {
		return false
	}
	if !isCanonicalAlphaCase(s) {
		return false
	}
	switch s[0] | 0x20 {
	case 'a':
		return (s[1]|0x20) == 'p' && (s[2]|0x20) == 'r' || (s[1]|0x20) == 'u' && (s[2]|0x20) == 'g'
	case 'd':
		return (s[1]|0x20) == 'e' && (s[2]|0x20) == 'c'
	case 'f':
		return (s[1]|0x20) == 'e' && (s[2]|0x20) == 'b'
	case 'j':
		b1 := s[1] | 0x20
		b2 := s[2] | 0x20
		return (b1 == 'a' && b2 == 'n') || (b1 == 'u' && b2 == 'n') || (b1 == 'u' && b2 == 'l')
	case 'm':
		return (s[1]|0x20) == 'a' && (s[2]|0x20) == 'r'
	case 'n':
		return (s[1]|0x20) == 'o' && (s[2]|0x20) == 'v'
	case 'o':
		return (s[1]|0x20) == 'c' && (s[2]|0x20) == 't'
	case 's':
		return (s[1]|0x20) == 'e' && (s[2]|0x20) == 'p'
	default:
		return false
	}
}

func isCanonicalAlphaCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	title := true

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 'A' || (c > 'Z' && c < 'a') || c > 'z' {
			return false
		}
		if i == 0 {
			if c < 'A' || c > 'Z' {
				title = false
			}
		} else if c < 'a' || c > 'z' {
			title = false
		}
	}

	return title
}

func consumeTimeOfDay(s string, start int) (int, bool) {
	i := start
	if !consumeTwoDigitRange(s, &i, 0, 23, true) || i >= len(s) || s[i] != ':' {
		return 0, false
	}
	i++
	if !consumeTwoDigitRange(s, &i, 0, 59, false) {
		return 0, false
	}
	if i >= len(s) || s[i] != ':' {
		return 0, false
	}
	i++
	if !consumeSecondValue(s, &i) {
		return 0, false
	}
	return i, true
}

func consumePatternTimeOfDay(s string, start int, allowLeadingNonDigit bool, allowTrailingNonDigit bool) (int, bool) {
	i := start
	if allowLeadingNonDigit && i < len(s) && !isASCIIDigit(s[i]) {
		i++
	}
	next, ok := consumeTimeOfDay(s, i)
	if !ok {
		return 0, false
	}
	i = next
	if i < len(s) && isASCIIDigit(s[i]) {
		return 0, false
	}
	if allowTrailingNonDigit && i < len(s) && !isASCIIDigit(s[i]) {
		i++
	}
	return i, true
}

func consumeNDigits(s string, i *int, n int) bool {
	if *i+n > len(s) {
		return false
	}
	for j := 0; j < n; j++ {
		if s[*i+j] < '0' || s[*i+j] > '9' {
			return false
		}
	}
	*i += n
	return true
}

func consumeOneOrTwoDigits(s string, i *int) bool {
	if *i >= len(s) || s[*i] < '0' || s[*i] > '9' {
		return false
	}
	*i = *i + 1
	if *i < len(s) && s[*i] >= '0' && s[*i] <= '9' {
		*i = *i + 1
	}
	return true
}

func consumeAtLeastOneDigit(s string, i *int) bool {
	start := *i
	for *i < len(s) && s[*i] >= '0' && s[*i] <= '9' {
		*i = *i + 1
	}
	return *i > start
}

func consumeOneOrTwoYearChunks(s string, i *int) bool {
	start := *i
	if !consumeNDigits(s, i, 2) {
		return false
	}
	if *i+2 <= len(s) {
		allDigits := true
		for j := *i; j < *i+2; j++ {
			if s[j] < '0' || s[j] > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			*i += 2
		}
	}
	return *i == start+2 || *i == start+4
}

func consumeTwoDigitRange(s string, i *int, min int, max int, allowSingle bool) bool {
	start := *i
	if allowSingle {
		if !consumeOneOrTwoDigits(s, i) {
			return false
		}
	} else {
		if !consumeNDigits(s, i, 2) {
			return false
		}
	}
	v, ok := parsePositiveInt(s[start:*i])
	return ok && v >= min && v <= max
}

func consumeSecondValue(s string, i *int) bool {
	start := *i
	if !consumeNDigits(s, i, 2) {
		return false
	}
	v, ok := parsePositiveInt(s[start:*i])
	if !ok || v < 0 || v > 60 {
		return false
	}
	if *i < len(s) && (s[*i] == '.' || s[*i] == ',' || s[*i] == ':') {
		*i++
		if !consumeAtLeastOneDigit(s, i) {
			return false
		}
	}
	return true
}

func consumeSignedInt(s string, i *int) bool {
	if *i < len(s) && (s[*i] == '+' || s[*i] == '-') {
		*i++
	}
	return consumeAtLeastOneDigit(s, i)
}

func isASCIIDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func sliceAlphaToken(content string, pos int) (string, int, bool) {
	start := pos
	for pos < len(content) {
		r, size := utf8.DecodeRuneInString(content[pos:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		if !unicode.IsLetter(r) {
			break
		}
		pos += size
	}
	if pos == start {
		return "", 0, false
	}
	return content[start:pos], pos, true
}

func isMonthNameValue(s string) bool {
	for _, candidate := range monthNameValues {
		if s == candidate {
			return true
		}
	}
	return false
}

func isDayNameValue(s string) bool {
	for _, candidate := range dayNameValues {
		if s == candidate {
			return true
		}
	}
	return false
}

func isLogLevelValue(s string) bool {
	for _, candidate := range logLevelValues {
		if matchesLogLevelCase(candidate, s) {
			return true
		}
	}
	return false
}

func matchesLogLevelCase(lower, actual string) bool {
	if len(actual) != len(lower) {
		return false
	}

	lowerMatch := true
	upperMatch := true
	titleMatch := true

	for i := 0; i < len(lower); i++ {
		wantLower := lower[i]
		got := actual[i]

		if got != wantLower {
			lowerMatch = false
		}

		wantUpper := wantLower
		if 'a' <= wantUpper && wantUpper <= 'z' {
			wantUpper -= 'a' - 'A'
		}
		if got != wantUpper {
			upperMatch = false
		}

		wantTitle := wantLower
		if i == 0 && 'a' <= wantTitle && wantTitle <= 'z' {
			wantTitle -= 'a' - 'A'
		}
		if got != wantTitle {
			titleMatch = false
		}

		if !lowerMatch && !upperMatch && !titleMatch {
			return false
		}
	}

	return lowerMatch || upperMatch || titleMatch
}

func equalFoldLiteral(s, literal string) bool {
	if len(s) != len(literal) {
		return strings.EqualFold(s, literal)
	}
	for i := 0; i < len(s); i++ {
		a := s[i]
		b := literal[i]
		if a == b {
			continue
		}
		if a >= utf8.RuneSelf || b >= utf8.RuneSelf {
			return strings.EqualFold(s, literal)
		}
		if 'A' <= a && a <= 'Z' {
			a += 'a' - 'A'
		}
		if 'A' <= b && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}
