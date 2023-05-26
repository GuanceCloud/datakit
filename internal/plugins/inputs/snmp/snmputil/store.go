// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

// Store MetadataStore stores metadata scalarValues.
type Store struct {
	// map[<FIELD>]ResultValue
	scalarValues map[string]ResultValue

	// map[<FIELD>][<index>]ResultValue
	columnValues map[string]map[string]ResultValue

	// map[<RESOURCE>][<index>][]<TAG>
	resourceIDTags map[string]map[string][]string
}

// NewMetadataStore returns a new metadata Store.
func NewMetadataStore() *Store {
	return &Store{
		scalarValues:   make(map[string]ResultValue),
		columnValues:   make(map[string]map[string]ResultValue),
		resourceIDTags: make(map[string]map[string][]string),
	}
}

// AddScalarValue add scalar value to metadata store.
func (s Store) AddScalarValue(field string, value ResultValue) {
	s.scalarValues[field] = value
}

// AddColumnValue add column value to metadata store.
func (s Store) AddColumnValue(field string, index string, value ResultValue) {
	if _, ok := s.columnValues[field]; !ok {
		s.columnValues[field] = make(map[string]ResultValue)
	}
	s.columnValues[field][index] = value
}

// GetColumnAsString get column value as string.
func (s Store) GetColumnAsString(field string, index string) string {
	column, ok := s.columnValues[field]
	if !ok {
		return ""
	}
	value, ok := column[index]
	if !ok {
		return ""
	}
	strVal, err := value.ToString()
	if err != nil {
		l.Debugf("error converting value string `%v`: %s", value, err)
		return ""
	}
	return strVal
}

// GetColumnAsFloat get column value as float.
func (s Store) GetColumnAsFloat(field string, index string) float64 {
	column, ok := s.columnValues[field]
	if !ok {
		return 0
	}
	value, ok := column[index]
	if !ok {
		return 0
	}
	strVal, err := value.ToFloat64()
	if err != nil {
		l.Debugf("error converting value to float `%v`: %v", value, err)
		return 0
	}
	return strVal
}

// GetScalarAsString get scalar value as string.
func (s Store) GetScalarAsString(field string) string {
	value, ok := s.scalarValues[field]
	if !ok {
		return ""
	}
	strVal, err := value.ToString()
	if err != nil {
		l.Debugf("error parsing value `%v`: %v", value, err)
		return ""
	}
	return strVal
}

// ScalarFieldHasValue test if scalar field has value.
func (s Store) ScalarFieldHasValue(field string) bool {
	_, ok := s.scalarValues[field]
	return ok
}

// GetColumnIndexes get column indexes for a field.
func (s Store) GetColumnIndexes(field string) []string {
	column, ok := s.columnValues[field]
	if !ok {
		return nil
	}
	var indexes []string
	for key := range column {
		indexes = append(indexes, key)
	}
	return indexes
}

// GetIDTags get idTags for a specific resource and index.
func (s Store) GetIDTags(resource string, index string) []string {
	resTags, ok := s.resourceIDTags[resource]
	if !ok {
		return nil
	}
	tags, ok := resTags[index]
	if !ok {
		return nil
	}
	return tags
}

// AddIDTags add idTags for a specific resource and index.
func (s Store) AddIDTags(resource string, index string, tags []string) {
	_, ok := s.resourceIDTags[resource]
	if !ok {
		s.resourceIDTags[resource] = make(map[string][]string)
	}
	s.resourceIDTags[resource][index] = append(s.resourceIDTags[resource][index], tags...)
}
