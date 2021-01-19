// Copyright 2018 Guilherme Caruso. All rights reserved.
// License that can be found in the LICENSE file.

/*
Package Kair Used to facilitate the process of formatting dates and times using human language in your process.

Examples:

> If you need to get the current time:
	Kair.Now().Time

> If you need to get a formatted date:
	Kair.Date().Time

> If you need to get a formatted datetime:
	Kair.DateTime().Time


> Using standart formatters:
	Kair.DateTime("20,05,2018,10,20,00").Format("LT") // "10:20 AM"
	Kair.DateTime("20,05,2018,10,20,00").Format("LL") // "29/5/2018"
	Kair.DateTime("20,05,2018,10,20,00").Format("llll") // "Mon, May 29, 2018 10:20 AM"

-> Using custom formatter:
	Kair.DateTime("20,05,2018,10,20,00").PersonalFormat(""MMM/dd/YY h:m:s"") // "May/20/18 10:20:0"
*/

package kair

import (
	"regexp"
	"strings"
	"time"
)

var (
	units  = []string{"LT", "LTS", "L", "l", "LL", "ll", "LLL", "lll", "LLLL", "llll"}
	months = map[int]time.Month{
		1:  time.January,
		2:  time.February,
		3:  time.March,
		4:  time.April,
		5:  time.May,
		6:  time.June,
		7:  time.July,
		8:  time.August,
		9:  time.September,
		10: time.October,
		11: time.November,
		12: time.December,
	}
	formats = map[string]string{
		"MMMM": "January",
		"MMM":  "Jan",
		"MM":   "01",
		"M":    "1",
		"YYYY": "2006",
		"YY":   "06",
		"DD":   "Monday",
		"D":    "Mon",
		"dd":   "02",
		"d":    "2",
		"hh":   "03",
		"h":    "3",
		"mm":   "04",
		"m":    "4",
		"ss":   "05",
		"s":    "5",
	}
)

//SKair Used to standardize the use of functions
type SKair struct {
	Time time.Time
}

/*
Format Uses standard sequence for the time format.
Returns a string standart format if var is invalid

Standard formats :
	["LT", "LTS", "L", "l", "LL", "ll", "LLL", "lll", "LLLL", "llll"]
*/
func (k *SKair) Format(format string) string {
	timeLint := k.Time
	var timeriz string
	switch format {
	case units[0]:
		timeriz = timeLint.Format("3:04 PM")
	case units[1]:
		timeriz = timeLint.Format("3:04:05 PM")
	case units[2]:
		timeriz = timeLint.Format("02/01/2006")
	case units[3]:
		timeriz = timeLint.Format("02/1/2006")
	case units[4]:
		timeriz = timeLint.Format("January 2, 2006")
	case units[5]:
		timeriz = timeLint.Format("Jan 2, 2006")
	case units[6]:
		timeriz = timeLint.Format("January 2, 2006 3:04 PM")
	case units[7]:
		timeriz = timeLint.Format("Jan 2, 2006 3:04 PM")
	case units[8]:
		timeriz = timeLint.Format("Monday, January 2, 2006 3:04 PM")
	case units[9]:
		timeriz = timeLint.Format("Mon, Jan 2, 2006 3:04 PM")
	default:
		timeriz = timeLint.String()
	}
	return timeriz
}

/*
CustomFormat Uses custom sequence for the time format.

Returns a string custom datetime format

Custom formatters :
	"MMMM": Long Month,
	"MMM":  Month,
	"MM":   Zero Number Month,
	"M":    Number Month,
	"YYYY": Long Year,
	"YY":   Year,
	"DD":   Long Day,
	"D":    Day,
	"dd":  	Long Number Day,
	"d":    Number Day,
	"hh":   Long Hour,
	"h":   	Hour,
	"mm":   Long Minute,
	"m":    Minute,
	"ss":   Long Second,
	"s":    Second

*/
func (k *SKair) CustomFormat(pformat string) string {
	re := regexp.MustCompile(`(?m)(M{4})|(M{3})|(M{2})|(M{1})|(Y{4})|(Y{2})|(D{2})|(D{1})|(d{2})|(d{1})|(h{2})|(h{1})|(m{2})|(m{1})|(s{2})|(s{1})`)

	for _, match := range re.FindAllString(pformat, -1) {
		if val, ok := formats[match]; ok {
			pformat = strings.Replace(pformat, match, val, -1)
		}
	}

	timeriz := k.Time.Format(pformat)
	return timeriz

}

//Now - Retrieve current datetime
func Now() *SKair {
	var k SKair
	k.Time = time.Now()
	return &k
}

//Date - Retrieves a custom date
func Date(day int, month int, year int) *SKair {
	var k SKair
	k.Time = time.Date(year, months[month], day, 0, 0, 0, 0, time.UTC)
	return &k
}

// DateTime - Retrieves a custom datetime
func DateTime(day int, month int, year int, hour int, min int, sec int) *SKair {
	var k SKair
	k.Time = time.Date(year, months[month], day, hour, min, sec, 0, time.UTC)
	return &k
}
