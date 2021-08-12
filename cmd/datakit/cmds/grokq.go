package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/vjeantet/grok"
)

const (
	Low = iota
	Low_
	Med
	Med_
	High
	High_
)

type grokPrevilege struct {
	g string
	p int
}

var (
	G *grok.Grok

	patternPrevList = []*grokPrevilege{
		{g: "TIME", p: High},
		{g: "DATE_US", p: High},
		{g: "DATE_EU", p: High},
		{g: "ISO8601_TIMEZONE", p: High},
		{g: "ISO8601_SECOND", p: High},
		{g: "TIMESTAMP_ISO8601", p: High},
		{g: "DATE", p: High},
		{g: "DATESTAMP", p: High},
		{g: "TZ", p: High},
		{g: "DATESTAMP_RFC822", p: High},
		{g: "DATESTAMP_EVENTLOG", p: High},
		{g: "HTTPDERROR_DATE", p: High},
		{g: "SYSLOGTIMESTAMP", p: High},
		{g: "HTTPDATE", p: High},
		{g: "COMMONAPACHELOG", p: High},
		{g: "COMBINEDAPACHELOG", p: High},
		{g: "HTTPD20_ERRORLOG", p: High},
		{g: "HTTPD24_ERRORLOG", p: High},
		{g: "HTTPD_ERRORLOG", p: High},
		{g: "LOGLEVEL", p: High},
		{g: "HOSTPORT", p: High},
		{g: "TTY", p: High},
		{g: "WINPATH", p: High},
		{g: "URIPATH", p: High},
		{g: "URIPARAM", p: High},
		{g: "URIPATHPARAM", p: High},
		{g: "URI", p: High},
		{g: "EMAILADDRESS", p: High},
		{g: "UUID", p: High},
		{g: "MAC", p: High},
		{g: "CISCOMAC", p: High},
		{g: "WINDOWSMAC", p: High},
		{g: "COMMONMAC", p: High},
		{g: "QUOTEDSTRING", p: Med},
		{g: "IPV6", p: High_},
		{g: "IPV4", p: High_},
		{g: "IP", p: Med},
		{g: "PATH", p: Med},
		{g: "UNIXPATH", p: Med},
		{g: "URIPROTO", p: Med},
		{g: "MONTH", p: Med},
		{g: "MONTHNUM", p: Med},
		{g: "MONTHNUM2", p: Med},
		{g: "MONTHDAY", p: Med},
		{g: "DAY", p: Med},
		{g: "YEAR", p: Med},
		{g: "HOUR", p: Med},
		{g: "MINUTE", p: Med},
		{g: "SECOND", p: Med},
		{g: "DATESTAMP_OTHER", p: Med},
		{g: "QS", p: Med},
		{g: "INT", p: Low_},
		{g: "POSINT", p: Low_},
		{g: "NUMBER", p: Low_},
		{g: "BASE16NUM", p: Low_},
		{g: "BASE10NUM", p: Low_},
		{g: "HTTPDUSER", p: Low},
		{g: "EMAILLOCALPART", p: High},
		{g: "USERNAME", p: Low},
		{g: "USER", p: Low},
		{g: "NONNEGINT", p: Low},
		{g: "URIHOST", p: Low},
		{g: "HOSTNAME", p: Low},
		{g: "HOST", p: Low},
		{g: "IPORHOST", p: Low},
		{g: "NOTSPACE", p: Low},
		{g: "SPACE", p: Low},
		{g: "PROG", p: Low},
		{g: "SYSLOGPROG", p: Low},
		{g: "SYSLOGHOST", p: Low},
		{g: "SYSLOGFACILITY", p: Low},
		{g: "SYSLOGBASE", p: Low},
		{g: "WORD", p: Low},
		{g: "DATA", p: Low},
		{g: "GREEDYDATA", p: Low},
	}
)

func grokq() {

	var err error
	G, err = grok.NewWithConfig(&grok.Config{
		SkipDefaultPatterns: false,
	})

	if err != nil {
		l.Fatalf("grok.NewWithConfig: %s", err)
	}

	c, _ := newCompleter()

	p := prompt.New(
		run,
		c.Complete,
		prompt.OptionTitle("grokq: DataKit grok query"),
		prompt.OptionPrefix("grokq > "),
	)

	p.Run()
}

func run(txt string) {
	s := strings.Join(strings.Fields(strings.TrimSpace(txt)), " ") // remove dup spaces
	if s == "" {
		return
	}

	switch s {
	case "Q", "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		do(s)
	}
}

func do(txt string) {

	matchedGroks := [High_ + 1]map[string]interface{}{}

	for _, ptn := range patternPrevList {
		res, err := G.Parse("%{"+ptn.g+"}", txt)
		if err != nil {
			l.Warnf("parse %%{%s} failed: %s", ptn.g, err.Error())
			continue
		}

		if len(res) != 0 {
			for _, v := range res {
				if v == txt {
					if matchedGroks[ptn.p] == nil {
						matchedGroks[ptn.p] = map[string]interface{}{}
					}
					matchedGroks[ptn.p][ptn.g] = nil
				}
			}
		}
	}

	for i := High_; i >= 0; i-- {
		for ptn, _ := range matchedGroks[i] {
			fmt.Printf("\t%d %%{%s: ?}\n", i, ptn)
		}
	}
}
