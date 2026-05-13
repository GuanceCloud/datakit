package grok

import internalscan "github.com/GuanceCloud/grok/internal/scan"

type atomScanner = internalscan.Scanner
type atomScanResult = internalscan.Result

func newAtomScanner(atoms []string) *atomScanner {
	return internalscan.New(atoms)
}
