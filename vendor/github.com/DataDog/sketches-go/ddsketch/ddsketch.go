// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2020 Datadog, Inc.

package ddsketch

import (
	"errors"
	"math"

	enc "github.com/DataDog/sketches-go/ddsketch/encoding"
	"github.com/DataDog/sketches-go/ddsketch/mapping"
	"github.com/DataDog/sketches-go/ddsketch/pb/sketchpb"
	"github.com/DataDog/sketches-go/ddsketch/store"
)

type DDSketch struct {
	mapping.IndexMapping
	positiveValueStore store.Store
	negativeValueStore store.Store
	zeroCount          float64
}

func NewDDSketchFromStoreProvider(indexMapping mapping.IndexMapping, storeProvider store.Provider) *DDSketch {
	return NewDDSketch(indexMapping, storeProvider(), storeProvider())
}

func NewDDSketch(indexMapping mapping.IndexMapping, positiveValueStore store.Store, negativeValueStore store.Store) *DDSketch {
	return &DDSketch{
		IndexMapping:       indexMapping,
		positiveValueStore: positiveValueStore,
		negativeValueStore: negativeValueStore,
	}
}

func NewDefaultDDSketch(relativeAccuracy float64) (*DDSketch, error) {
	return LogUnboundedDenseDDSketch(relativeAccuracy)
}

// Constructs an instance of DDSketch that offers constant-time insertion and whose size grows indefinitely
// to accommodate for the range of input values.
func LogUnboundedDenseDDSketch(relativeAccuracy float64) (*DDSketch, error) {
	indexMapping, err := mapping.NewLogarithmicMapping(relativeAccuracy)
	if err != nil {
		return nil, err
	}
	return NewDDSketch(indexMapping, store.NewDenseStore(), store.NewDenseStore()), nil
}

// Constructs an instance of DDSketch that offers constant-time insertion and whose size grows until the
// maximum number of bins is reached, at which point bins with lowest indices are collapsed, which causes the
// relative accuracy guarantee to be lost on lowest quantiles if values are all positive, or the mid-range
// quantiles for values closest to zero if values include negative numbers.
func LogCollapsingLowestDenseDDSketch(relativeAccuracy float64, maxNumBins int) (*DDSketch, error) {
	indexMapping, err := mapping.NewLogarithmicMapping(relativeAccuracy)
	if err != nil {
		return nil, err
	}
	return NewDDSketch(indexMapping, store.NewCollapsingLowestDenseStore(maxNumBins), store.NewCollapsingLowestDenseStore(maxNumBins)), nil
}

// Constructs an instance of DDSketch that offers constant-time insertion and whose size grows until the
// maximum number of bins is reached, at which point bins with highest indices are collapsed, which causes the
// relative accuracy guarantee to be lost on highest quantiles if values are all positive, or the lowest and
// highest quantiles if values include negative numbers.
func LogCollapsingHighestDenseDDSketch(relativeAccuracy float64, maxNumBins int) (*DDSketch, error) {
	indexMapping, err := mapping.NewLogarithmicMapping(relativeAccuracy)
	if err != nil {
		return nil, err
	}
	return NewDDSketch(indexMapping, store.NewCollapsingHighestDenseStore(maxNumBins), store.NewCollapsingHighestDenseStore(maxNumBins)), nil
}

// Adds a value to the sketch.
func (s *DDSketch) Add(value float64) error {
	return s.AddWithCount(value, float64(1))
}

// Adds a value to the sketch with a float64 count.
func (s *DDSketch) AddWithCount(value, count float64) error {
	if value < -s.MaxIndexableValue() || value > s.MaxIndexableValue() {
		return errors.New("The input value is outside the range that is tracked by the sketch.")
	}
	if count < 0 {
		return errors.New("The count cannot be negative.")
	}

	if value > s.MinIndexableValue() {
		s.positiveValueStore.AddWithCount(s.Index(value), count)
	} else if value < -s.MinIndexableValue() {
		s.negativeValueStore.AddWithCount(s.Index(-value), count)
	} else {
		s.zeroCount += count
	}
	return nil
}

// Return a (deep) copy of this sketch.
func (s *DDSketch) Copy() *DDSketch {
	return &DDSketch{
		IndexMapping:       s.IndexMapping,
		positiveValueStore: s.positiveValueStore.Copy(),
		negativeValueStore: s.negativeValueStore.Copy(),
		zeroCount:          s.zeroCount,
	}
}

// Clear empties the sketch while allowing reusing already allocated memory.
func (s *DDSketch) Clear() {
	s.positiveValueStore.Clear()
	s.negativeValueStore.Clear()
	s.zeroCount = 0
}

// Return the value at the specified quantile. Return a non-nil error if the quantile is invalid
// or if the sketch is empty.
func (s *DDSketch) GetValueAtQuantile(quantile float64) (float64, error) {
	if quantile < 0 || quantile > 1 {
		return math.NaN(), errors.New("The quantile must be between 0 and 1.")
	}

	count := s.GetCount()
	if count == 0 {
		return math.NaN(), errors.New("No such element exists")
	}

	rank := quantile * (count - 1)
	negativeValueCount := s.negativeValueStore.TotalCount()
	if rank < negativeValueCount {
		return -s.Value(s.negativeValueStore.KeyAtRank(negativeValueCount - 1 - rank)), nil
	} else if rank < s.zeroCount+negativeValueCount {
		return 0, nil
	} else {
		return s.Value(s.positiveValueStore.KeyAtRank(rank - s.zeroCount - negativeValueCount)), nil
	}
}

// Return the values at the respective specified quantiles. Return a non-nil error if any of the quantiles
// is invalid or if the sketch is empty.
func (s *DDSketch) GetValuesAtQuantiles(quantiles []float64) ([]float64, error) {
	values := make([]float64, len(quantiles))
	for i, q := range quantiles {
		val, err := s.GetValueAtQuantile(q)
		if err != nil {
			return nil, err
		}
		values[i] = val
	}
	return values, nil
}

// Return the total number of values that have been added to this sketch.
func (s *DDSketch) GetCount() float64 {
	return s.zeroCount + s.positiveValueStore.TotalCount() + s.negativeValueStore.TotalCount()
}

// Return true iff no value has been added to this sketch.
func (s *DDSketch) IsEmpty() bool {
	return s.zeroCount == 0 && s.positiveValueStore.IsEmpty() && s.negativeValueStore.IsEmpty()
}

// Return the maximum value that has been added to this sketch. Return a non-nil error if the sketch
// is empty.
func (s *DDSketch) GetMaxValue() (float64, error) {
	if !s.positiveValueStore.IsEmpty() {
		maxIndex, _ := s.positiveValueStore.MaxIndex()
		return s.Value(maxIndex), nil
	} else if s.zeroCount > 0 {
		return 0, nil
	} else {
		minIndex, err := s.negativeValueStore.MinIndex()
		if err != nil {
			return math.NaN(), err
		}
		return -s.Value(minIndex), nil
	}
}

// Return the minimum value that has been added to this sketch. Returns a non-nil error if the sketch
// is empty.
func (s *DDSketch) GetMinValue() (float64, error) {
	if !s.negativeValueStore.IsEmpty() {
		maxIndex, _ := s.negativeValueStore.MaxIndex()
		return -s.Value(maxIndex), nil
	} else if s.zeroCount > 0 {
		return 0, nil
	} else {
		minIndex, err := s.positiveValueStore.MinIndex()
		if err != nil {
			return math.NaN(), err
		}
		return s.Value(minIndex), nil
	}
}

// GetSum returns an approximation of the sum of the values that have been added to the sketch. If the
// values that have been added to the sketch all have the same sign, the approximation error has
// the relative accuracy guarantees of the mapping used for this sketch.
func (s *DDSketch) GetSum() (sum float64) {
	s.ForEach(func(value float64, count float64) (stop bool) {
		sum += value * count
		return false
	})
	return sum
}

// ForEach applies f on the bins of the sketches until f returns true.
// There is no guarantee on the bin iteration order.
func (s *DDSketch) ForEach(f func(value, count float64) (stop bool)) {
	if s.zeroCount != 0 && f(0, s.zeroCount) {
		return
	}
	stopped := false
	s.positiveValueStore.ForEach(func(index int, count float64) bool {
		stopped = f(s.IndexMapping.Value(index), count)
		return stopped
	})
	if stopped {
		return
	}
	s.negativeValueStore.ForEach(func(index int, count float64) bool {
		return f(-s.IndexMapping.Value(index), count)
	})
}

// Merges the other sketch into this one. After this operation, this sketch encodes the values that
// were added to both this and the other sketches.
func (s *DDSketch) MergeWith(other *DDSketch) error {
	if !s.IndexMapping.Equals(other.IndexMapping) {
		return errors.New("Cannot merge sketches with different index mappings.")
	}
	s.positiveValueStore.MergeWith(other.positiveValueStore)
	s.negativeValueStore.MergeWith(other.negativeValueStore)
	s.zeroCount += other.zeroCount
	return nil
}

// Generates a protobuf representation of this DDSketch.
func (s *DDSketch) ToProto() *sketchpb.DDSketch {
	return &sketchpb.DDSketch{
		Mapping:        s.IndexMapping.ToProto(),
		PositiveValues: s.positiveValueStore.ToProto(),
		NegativeValues: s.negativeValueStore.ToProto(),
		ZeroCount:      s.zeroCount,
	}
}

// FromProto builds a new instance of DDSketch based on the provided protobuf representation, using a Dense store.
func FromProto(pb *sketchpb.DDSketch) (*DDSketch, error) {
	return FromProtoWithStoreProvider(pb, store.DenseStoreConstructor)
}

func FromProtoWithStoreProvider(pb *sketchpb.DDSketch, storeProvider store.Provider) (*DDSketch, error) {
	positiveValueStore := storeProvider()
	store.MergeWithProto(positiveValueStore, pb.PositiveValues)
	negativeValueStore := storeProvider()
	store.MergeWithProto(negativeValueStore, pb.NegativeValues)
	m, err := mapping.FromProto(pb.Mapping)
	if err != nil {
		return nil, err
	}
	return &DDSketch{
		IndexMapping:       m,
		positiveValueStore: positiveValueStore,
		negativeValueStore: negativeValueStore,
		zeroCount:          pb.ZeroCount,
	}, nil
}

// Encode serializes the sketch and appends the serialized content to the provided []byte.
// If the capacity of the provided []byte is large enough, Encode does not allocate memory space.
// When the index mapping is known at the time of deserialization, omitIndexMapping can be set to true to avoid encoding it and to make the serialized content smaller.
// The encoding format is described in the encoding/flag module.
func (s *DDSketch) Encode(b *[]byte, omitIndexMapping bool) {
	if s.zeroCount != 0 {
		enc.EncodeFlag(b, enc.FlagZeroCountVarFloat)
		enc.EncodeVarfloat64(b, s.zeroCount)
	}

	if !omitIndexMapping {
		s.IndexMapping.Encode(b)
	}

	s.positiveValueStore.Encode(b, enc.FlagTypePositiveStore)
	s.negativeValueStore.Encode(b, enc.FlagTypeNegativeStore)
}

// DecodeDDSketch deserializes a sketch.
// Stores are built using storeProvider. The store type needs not match the
// store that the serialized sketch initially used. However, using the same
// store type may make decoding faster. In the absence of high performance
// requirements, store.BufferedPaginatedStoreConstructor is a sound enough
// choice of store provider.
// To avoid memory allocations, it is possible to use a store provider that
// reuses stores, by calling Clear() on previously used stores before providing
// the store.
// If the serialized content does not contain the index mapping, DecodeDDSketch
// returns an error.
func DecodeDDSketch(b []byte, storeProvider store.Provider) (*DDSketch, error) {
	return DecodeDDSketchWithIndexMapping(b, storeProvider, nil)
}

// DecodeDDSketchWithIndexMapping deserializes a sketch.
// Stores are built using storeProvider. The store type needs not match the
// store that the serialized sketch initially used. However, using the same
// store type may make decoding faster. In the absence of high performance
// requirements, store.BufferedPaginatedStoreConstructor is a sound enough
// choice of store provider.
// To avoid memory allocations, it is possible to use a store provider that
// reuses stores, by calling Clear() on previously used stores before providing
// the store.
// If the serialized content contains an index mapping that differs from the
// provided one, DecodeDDSketchWithIndexMapping returns an error.
func DecodeDDSketchWithIndexMapping(b []byte, storeProvider store.Provider, indexMapping mapping.IndexMapping) (*DDSketch, error) {
	s := &DDSketch{
		IndexMapping:       indexMapping,
		positiveValueStore: storeProvider(),
		negativeValueStore: storeProvider(),
		zeroCount:          float64(0),
	}
	err := s.DecodeAndMergeWith(b)
	return s, err
}

// DecodeAndMergeWith deserializes a sketch and merges its content in the
// receiver sketch.
// If the serialized content contains an index mapping that differs from the one
// of the receiver, DecodeAndMergeWith returns an error.
func (s *DDSketch) DecodeAndMergeWith(bb []byte) error {
	b := &bb
	for len(*b) > 0 {
		flag, err := enc.DecodeFlag(b)
		if err != nil {
			return err
		}
		switch flag.Type() {
		case enc.FlagTypePositiveStore:
			s.positiveValueStore.DecodeAndMergeWith(b, flag.SubFlag())
		case enc.FlagTypeNegativeStore:
			s.negativeValueStore.DecodeAndMergeWith(b, flag.SubFlag())
		case enc.FlagTypeIndexMapping:
			decodedIndexMapping, err := mapping.Decode(b, flag)
			if err != nil {
				return err
			}
			if s.IndexMapping != nil && !s.IndexMapping.Equals(decodedIndexMapping) {
				return errors.New("index mapping mismatch")
			}
			s.IndexMapping = decodedIndexMapping
		default:
			switch flag {

			case enc.FlagZeroCountVarFloat:
				decodedZeroCount, err := enc.DecodeVarfloat64(b)
				if err != nil {
					return err
				}
				s.zeroCount += decodedZeroCount

			default:
				return errors.New("unknown encoding flag")
			}
		}
	}

	if s.IndexMapping == nil {
		return errors.New("missing index mapping")
	}
	return nil
}

// ChangeMapping changes the store to a new mapping.
// it doesn't change s but returns a newly created sketch.
// positiveStore and negativeStore must be different stores, and be empty when the function is called.
// It is not the conversion that minimizes the loss in relative
// accuracy, but it avoids artefacts like empty bins that make the histograms look bad.
// scaleFactor allows to scale out / in all values. (changing units for eg)
func (s *DDSketch) ChangeMapping(newMapping mapping.IndexMapping, positiveStore store.Store, negativeStore store.Store, scaleFactor float64) *DDSketch {
	if scaleFactor == 1 && s.IndexMapping.Equals(newMapping) {
		return s.Copy()
	}
	changeStoreMapping(s.IndexMapping, newMapping, s.positiveValueStore, positiveStore, scaleFactor)
	changeStoreMapping(s.IndexMapping, newMapping, s.negativeValueStore, negativeStore, scaleFactor)
	newSketch := NewDDSketch(newMapping, positiveStore, negativeStore)
	newSketch.zeroCount = s.zeroCount
	return newSketch
}

func changeStoreMapping(oldMapping, newMapping mapping.IndexMapping, oldStore, newStore store.Store, scaleFactor float64) {
	oldStore.ForEach(func(index int, count float64) (stop bool) {
		inLowerBound := oldMapping.LowerBound(index) * scaleFactor
		inHigherBound := oldMapping.LowerBound(index+1) * scaleFactor
		inSize := inHigherBound - inLowerBound
		for outIndex := newMapping.Index(inLowerBound); newMapping.LowerBound(outIndex) < inHigherBound; outIndex++ {
			outLowerBound := newMapping.LowerBound(outIndex)
			outHigherBound := newMapping.LowerBound(outIndex + 1)
			lowerIntersectionBound := math.Max(outLowerBound, inLowerBound)
			higherIntersectionBound := math.Min(outHigherBound, inHigherBound)
			intersectionSize := higherIntersectionBound - lowerIntersectionBound
			proportion := intersectionSize / inSize
			newStore.AddWithCount(outIndex, proportion*count)
		}
		return false
	})
}

// Reweight multiplies all values from the sketch by w, but keeps the same global distribution.
// w has to be strictly greater than 0.
func (s *DDSketch) Reweight(w float64) error {
	if w <= 0 {
		return errors.New("can't reweight by a negative factor")
	}
	if w == 1 {
		return nil
	}
	s.zeroCount *= w
	if err := s.positiveValueStore.Reweight(w); err != nil {
		return err
	}
	if err := s.negativeValueStore.Reweight(w); err != nil {
		return err
	}
	return nil
}
