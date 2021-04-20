package mathjax

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const common = `
\documentclass[preview]{standalone}
\usepackage{amsmath}
\usepackage{amssymb}
\usepackage{stmaryrd}
\begin{document}
%s
\end{document}
`

const displayInlineFormula = `
\documentclass[11pt]{article}
\usepackage[active,tightpage,textmath]{preview}
\usepackage{amsmath}
\usepackage{amssymb}
\usepackage{stmaryrd}
\begin{document}
\(%s\)
\end{document}
`

const displayBlockFormula = `
\documentclass[11pt,preview]{standalone}
\usepackage{amsmath}
\usepackage{amssymb}
\usepackage{stmaryrd}
\begin{document}
\begin{equation*}
%s
\end{equation*}
\end{document}
`

const tmpl = `
\documentclass[11pt]{article}
\usepackage[paperwidth=180in,paperheight=180in]{geometry}
\batchmode
\usepackage[utf8]{inputenc}
\usepackage{amsmath}
\usepackage{amssymb}
\usepackage{stmaryrd}
\usepackage[verbose]{newunicodechar}
\pagestyle{empty}
\setlength{\topskip}{0pt}
\setlength{\parindent}{0pt}
\setlength{\abovedisplayskip}{0pt}
\setlength{\belowdisplayskip}{0pt}
\begin{document}
%s
\end{document}
`

const tikz = `
\documentclass[11pt]{article}
\usepackage{tikz}
\usepackage{lipsum}
\usepackage{paralist,pst-func, pst-plot, pst-math, pstricks-add,pgfplots}
\usetikzlibrary{patterns,matrix,arrows}
\usepackage[active,tightpage]{preview}
\PreviewEnvironment{tikzpicture}
\setlength\PreviewBorder{1pt}

\begin{document}
%s
\end{document}
`

type TexRenderer struct {
	texPath           string
	docTemplate       string
	inlineFormulaImpl string
	commonBlockTmpl   string
	blockFormulaTmpl  string
	tikzTmpl          string
	tmpDir            string
}

func NewDefaultTexRenderer() *TexRenderer {
	var wd, _ = os.Getwd()
	var texPath = os.Getenv("TEX_PATH")

	var tmpDir = wd + "/tmp/"

	var defaultRenderer = &TexRenderer{
		texPath:           texPath,
		docTemplate:       tmpl,
		inlineFormulaImpl: displayInlineFormula,
		commonBlockTmpl:   common,
		blockFormulaTmpl:  displayBlockFormula,
		tmpDir:            tmpDir,
		tikzTmpl:          tikz,
	}
	return defaultRenderer
}

func (r *TexRenderer) RunInline(formula string) []byte {
	return r.runRaw(fmt.Sprintf(r.inlineFormulaImpl, formula))
}

func (r *TexRenderer) Run(formula string) []byte {
	var tmpl string
	formula = strings.TrimSpace(formula)
	if strings.Contains(formula, `\begin{tikzpicture}`) {
		tmpl = r.tikzTmpl
	} else if (strings.HasPrefix(formula, `\begin{`)) {
		tmpl = r.commonBlockTmpl
	} else {
		tmpl = r.blockFormulaTmpl
	}
	return r.runRaw(fmt.Sprintf(tmpl, strings.TrimSpace(formula)))
}

func (r *TexRenderer) runRaw(formula string) []byte {
	f, err := ioutil.TempFile(r.tmpDir, "doc")
	if err != nil {
		log.Fatalf("%v", err)
	}
	f.WriteString(formula)
	f.Sync()
	f.Close()
	r.runPdfLatex(f.Name())
	r.runPdf2Svg(f.Name())
	svgf, err := os.Open(f.Name() + ".svg")
	if err != nil {
		return nil
	}
	svg, err := ioutil.ReadAll(svgf)
	if err != nil {
		return nil
	}
	return svg
}

func (r *TexRenderer) runDvi2Svg(fname string) {
	// fmt.Println([]string{fmt.Sprintf("%sdvisvgm", r.texPath), fmt.Sprintf("%s.dvi", fname), "-o", fmt.Sprintf("%s.svg", fname), "-n", "--exact", "-v0", "--relative", "--zoom=1.2546875"})

	cmd := exec.Command(fmt.Sprintf("%sdvisvgm", r.texPath), fmt.Sprintf("%s.dvi", fname), "-o", fmt.Sprintf("%s.svg", fname), "-n", "--exact", "-v0", "--relative", "--zoom=1.2546875")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("dvi2svg cmd.Run() failed with %s\n", err)
	}
	// outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	// fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
}

func (r *TexRenderer) runLatex(fname string) {
	cmd := exec.Command(fmt.Sprintf("%slatex", r.texPath), "-output-directory", r.tmpDir, fname)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("latex cmd.Run() failed with %s\n", err)
	}
	// outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	// fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
}

func (r *TexRenderer) runPdfLatex(fname string) {
	cmd := exec.Command(fmt.Sprintf("%spdflatex", r.texPath), "-output-directory", r.tmpDir, fname)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("pdflatex %s cmd.Run() failed with %s\n", fname, err)
	}
}

func (r *TexRenderer) runPdf2Svg(fname string) {
	cmd := exec.Command("pdf2svg", fmt.Sprintf("%s.pdf", fname), fmt.Sprintf("%s.svg", fname))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("pdf2svg cmd.Run() failed with %s\n", err)
	}
}
