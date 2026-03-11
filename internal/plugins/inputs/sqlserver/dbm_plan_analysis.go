// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type PlanAnalysis struct {
	// Core metrics
	TotalCost     float64
	CompileTime   int64
	EstimatedRows float64

	// List information
	Tables      []string
	Warnings    []string
	IndexesUsed []string

	// Structured information (JSON serialized)
	Operators      []*OperatorInfo
	MissingIndexes []*MissingIndexGroupInfo
}

type MissingIndexGroupInfo struct {
	Impact         float64
	MissingIndexes []*MissingIndexInfo
}

type MissingIndexInfo struct {
	Table    string
	Keys     []string
	Includes []string
}
type OperatorInfo struct {
	NodeID            int             `json:"node_id"`
	PhysicalOp        string          `json:"physical_op"`
	LogicalOp         string          `json:"logical_op"`
	Cost              float64         `json:"cost"`
	EstimatedRows     float64         `json:"estimated_rows,omitempty"`
	EstimatedRowsRead float64         `json:"estimated_rows_read,omitempty"`
	ActualRows        float64         `json:"actual_rows,omitempty"`
	TableName         string          `json:"table_name,omitempty"`
	IndexName         string          `json:"index_name,omitempty"`
	Children          []*OperatorInfo `json:"children,omitempty"`
}

type planParser struct {
	analysis                 *PlanAnalysis
	operatorStack            []*OperatorInfo
	currentMissingIndex      *MissingIndexInfo
	currentMissingIndexGroup *MissingIndexGroupInfo
	currentColumnGroupUsage  string // "EQUALITY", "RANGE", or "INCLUDE"
	inQueryPlan              bool
	inMissingIndexes         bool
	inWarnings               bool
	tablesMap                map[string]struct{}
	indexesMap               map[string]struct{}
	warningsMap              map[string]struct{}
}

// parseExecutionPlan parses XML execution plan and extracts key information.
func parseExecutionPlan(xmlPlan string) (*PlanAnalysis, error) {
	parser := &planParser{
		analysis:    &PlanAnalysis{},
		tablesMap:   make(map[string]struct{}),
		indexesMap:  make(map[string]struct{}),
		warningsMap: make(map[string]struct{}),
	}

	decoder := xml.NewDecoder(strings.NewReader(xmlPlan))
	decoder.Strict = false

	if err := parser.parse(decoder); err != nil {
		return nil, err
	}

	return parser.analysis, nil
}

// parse processes XML tokens.
func (p *planParser) parse(decoder *xml.Decoder) error {
	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("failed to decode XML token: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if err := p.handleStartElement(t); err != nil {
				return err
			}
		case xml.EndElement:
			p.handleEndElement(t)
		}
	}
	return nil
}

// handleStartElement processes XML start elements.
func (p *planParser) handleStartElement(t xml.StartElement) error {
	switch t.Name.Local {
	case "StmtSimple":
		p.parseStmtSimple(t)
	case "QueryPlan":
		p.parseQueryPlan(t)
	case "RelOp":
		p.parseRelOp(t)
	case "MissingIndexes":
		p.inMissingIndexes = true
	case "MissingIndex":
		p.parseMissingIndex(t)
	case "MissingIndexGroup":
		p.parseMissingIndexGroup(t)
	case "ColumnGroup":
		p.parseColumnGroup(t)
	case "Column":
		p.parseColumn(t)
	case "Warnings":
		p.inWarnings = true
	case "Object":
		p.parseObject(t)
	default:
		// If inside Warnings section, treat any element as a warning type
		if p.inWarnings {
			p.parseWarning(t)
		}
	}
	return nil
}

// handleEndElement processes XML end elements.
func (p *planParser) handleEndElement(t xml.EndElement) {
	switch t.Name.Local {
	case "RelOp":
		p.endRelOp()
	case "MissingIndex":
		p.endMissingIndex()
	case "MissingIndexGroup":
		p.endMissingIndexGroup()
	case "ColumnGroup":
		p.endColumnGroup()
	case "MissingIndexes":
		p.inMissingIndexes = false
	case "Warnings":
		p.inWarnings = false
	case "QueryPlan":
		p.inQueryPlan = false
	}
}

// parseStmtSimple extracts statement-level attributes.
func (p *planParser) parseStmtSimple(t xml.StartElement) {
	for _, attr := range t.Attr {
		switch attr.Name.Local {
		case "StatementEstRows":
			if rows, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				p.analysis.EstimatedRows = rows
			}
		case "StatementSubTreeCost":
			if cost, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				p.analysis.TotalCost = cost
			}
		}
	}
}

// parseQueryPlan extracts query plan attributes.
func (p *planParser) parseQueryPlan(t xml.StartElement) {
	for _, attr := range t.Attr {
		if attr.Name.Local == "CompileTime" {
			if compileTime, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
				p.analysis.CompileTime = compileTime
			}
		}
	}
	p.inQueryPlan = true
}

// parseRelOp processes a RelOp element.
func (p *planParser) parseRelOp(t xml.StartElement) {
	if !p.inQueryPlan {
		return
	}

	op := &OperatorInfo{
		Children: make([]*OperatorInfo, 0),
	}

	for _, attr := range t.Attr {
		switch attr.Name.Local {
		case "NodeId":
			if nodeID, err := strconv.Atoi(attr.Value); err == nil {
				op.NodeID = nodeID
			}
		case "PhysicalOp":
			op.PhysicalOp = attr.Value
		case "LogicalOp":
			op.LogicalOp = attr.Value
		case "EstimatedTotalSubtreeCost":
			if cost, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				op.Cost = cost
			}
		case "EstimateRows":
			if rows, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				op.EstimatedRows = rows
			}
		case "ActualRows":
			if rows, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				op.ActualRows = rows
			}
		case "EstimatedRowsRead":
			if rows, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				op.EstimatedRowsRead = rows
			}
		}
	}

	if len(p.operatorStack) > 0 {
		parent := p.operatorStack[len(p.operatorStack)-1]
		parent.Children = append(parent.Children, op)
	} else {
		p.analysis.Operators = append(p.analysis.Operators, op)
	}

	p.operatorStack = append(p.operatorStack, op)
}

// endRelOp handles RelOp end element.
func (p *planParser) endRelOp() {
	if len(p.operatorStack) > 0 {
		p.operatorStack = p.operatorStack[:len(p.operatorStack)-1]
	}
}

// parseMissingIndex processes MissingIndex element.
func (p *planParser) parseMissingIndex(t xml.StartElement) {
	if !p.inMissingIndexes {
		return
	}
	p.currentMissingIndex = &MissingIndexInfo{
		Keys:     make([]string, 0),
		Includes: make([]string, 0),
	}

	for _, attr := range t.Attr {
		if attr.Name.Local == "Table" {
			p.currentMissingIndex.Table = strings.Trim(attr.Value, "[]")
		}
	}
}

// parseMissingIndexGroup processes MissingIndexGroup element.
func (p *planParser) parseMissingIndexGroup(t xml.StartElement) {
	if !p.inMissingIndexes {
		return
	}
	p.currentMissingIndexGroup = &MissingIndexGroupInfo{
		MissingIndexes: make([]*MissingIndexInfo, 0),
	}
	for _, attr := range t.Attr {
		if attr.Name.Local == "Impact" {
			if impact, err := strconv.ParseFloat(attr.Value, 64); err == nil {
				p.currentMissingIndexGroup.Impact = impact
			}
		}
	}
}

// endMissingIndex handles MissingIndex end element.
func (p *planParser) endMissingIndex() {
	if p.currentMissingIndex == nil {
		return
	}
	// Only add MissingIndex if it has a Table and we're inside a MissingIndexGroup
	if p.currentMissingIndex.Table != "" && p.currentMissingIndexGroup != nil {
		p.currentMissingIndexGroup.MissingIndexes = append(p.currentMissingIndexGroup.MissingIndexes, p.currentMissingIndex)
	}
	p.currentMissingIndex = nil
}

// endMissingIndexGroup handles MissingIndexGroup end element.
func (p *planParser) endMissingIndexGroup() {
	if p.currentMissingIndexGroup == nil {
		return
	}
	// Only add MissingIndexGroup if it has at least one MissingIndex
	if len(p.currentMissingIndexGroup.MissingIndexes) > 0 {
		p.analysis.MissingIndexes = append(p.analysis.MissingIndexes, p.currentMissingIndexGroup)
	}
	p.currentMissingIndexGroup = nil
}

// parseColumnGroup processes ColumnGroup element.
func (p *planParser) parseColumnGroup(t xml.StartElement) {
	if p.currentMissingIndex == nil {
		return
	}
	// Reset current column group usage
	p.currentColumnGroupUsage = ""
	for _, attr := range t.Attr {
		if attr.Name.Local == "Usage" {
			p.currentColumnGroupUsage = attr.Value // "EQUALITY", "RANGE", or "INCLUDE"
			break
		}
	}
}

// endColumnGroup handles ColumnGroup end element.
func (p *planParser) endColumnGroup() {
	// Clear current column group usage
	p.currentColumnGroupUsage = ""
}

// parseColumn processes Column element.
func (p *planParser) parseColumn(t xml.StartElement) {
	if p.currentMissingIndex == nil {
		return
	}

	var columnName string
	for _, attr := range t.Attr {
		if attr.Name.Local == "Name" {
			// Remove brackets, e.g., [category] -> category
			columnName = strings.Trim(attr.Value, "[]")
			break
		}
	}

	if columnName == "" {
		return
	}

	// Add to Keys if Usage is EQUALITY or RANGE, otherwise add to Includes
	if p.currentColumnGroupUsage == "EQUALITY" || p.currentColumnGroupUsage == "RANGE" {
		p.currentMissingIndex.Keys = append(p.currentMissingIndex.Keys, columnName)
	} else if p.currentColumnGroupUsage == "INCLUDE" {
		p.currentMissingIndex.Includes = append(p.currentMissingIndex.Includes, columnName)
	}
}

// parseObject processes Object element to extract table and index information.
func (p *planParser) parseObject(t xml.StartElement) {
	if len(p.operatorStack) == 0 {
		return
	}

	current := p.operatorStack[len(p.operatorStack)-1]

	var tableName, indexName string
	for _, attr := range t.Attr {
		switch attr.Name.Local {
		case "Table":
			// Remove all brackets and dots, e.g., [dbo].[Users] -> dbo.Users
			tableName = strings.Trim(attr.Value, "[]")
			if current.TableName == "" {
				current.TableName = tableName
			}
		case "Index":
			// Remove all brackets, e.g., [PK_Users] -> PK_Users
			indexName = strings.Trim(attr.Value, "[]")
			if current.IndexName == "" {
				current.IndexName = indexName
			}
		}
	}

	if tableName != "" {
		if _, exists := p.tablesMap[tableName]; !exists {
			p.tablesMap[tableName] = struct{}{}
			p.analysis.Tables = append(p.analysis.Tables, tableName)
		}
	}

	if tableName != "" && indexName != "" {
		indexKey := fmt.Sprintf("%s.%s", tableName, indexName)
		if _, exists := p.indexesMap[indexKey]; !exists {
			p.indexesMap[indexKey] = struct{}{}
			p.analysis.IndexesUsed = append(p.analysis.IndexesUsed, indexKey)
		}
	}
}

// parseWarning processes a warning element.
func (p *planParser) parseWarning(t xml.StartElement) {
	warningType := t.Name.Local
	if _, exists := p.warningsMap[warningType]; !exists {
		p.warningsMap[warningType] = struct{}{}
		p.analysis.Warnings = append(p.analysis.Warnings, warningType)
	}
}
