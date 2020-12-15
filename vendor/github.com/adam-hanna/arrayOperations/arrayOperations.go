package arrayOperations

import (
	"reflect"
)

// Distinct returns the unique vals of a slice
//
// [1, 1, 2, 3] >> [1, 2, 3]
func Distinct(arr interface{}) (reflect.Value, bool) {
	// create a slice from our input interface
	slice, ok := takeArg(arr, reflect.Slice)
	if !ok {
		return reflect.Value{}, ok
	}

	// put the values of our slice into a map
	// the key's of the map will be the slice's unique values
	c := slice.Len()
	m := make(map[interface{}]bool)
	for i := 0; i < c; i++ {
		m[slice.Index(i).Interface()] = true
	}
	mapLen := len(m)

	// create the output slice and populate it with the map's keys
	out := reflect.MakeSlice(reflect.TypeOf(arr), mapLen, mapLen)
	i := 0
	for k := range m {
		v := reflect.ValueOf(k)
		o := out.Index(i)
		o.Set(v)
		i++
	}

	return out, ok
}

// Intersect returns a slice of values that are present in all of the input slices
//
// [1, 1, 3, 4, 5, 6] & [2, 3, 6] >> [3, 6]
//
// [1, 1, 3, 4, 5, 6] >> [1, 3, 4, 5, 6]
func Intersect(arrs ...interface{}) (reflect.Value, bool) {
	// create a map to count all the instances of the slice elems
	arrLength := len(arrs)
	var kind reflect.Kind
	var kindHasBeenSet bool

	tempMap := make(map[interface{}]int)
	for _, arg := range arrs {
		tempArr, ok := Distinct(arg)
		if !ok {
			return reflect.Value{}, ok
		}

		// check to be sure the type hasn't changed
		if kindHasBeenSet && tempArr.Len() > 0 && tempArr.Index(0).Kind() != kind {
			return reflect.Value{}, false
		}
		if tempArr.Len() > 0 {
			kindHasBeenSet = true
			kind = tempArr.Index(0).Kind()
		}

		c := tempArr.Len()
		for idx := 0; idx < c; idx++ {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr.Index(idx).Interface()]; ok {
				tempMap[tempArr.Index(idx).Interface()]++
			} else {
				tempMap[tempArr.Index(idx).Interface()] = 1
			}
		}
	}

	// find the keys equal to the length of the input args
	numElems := 0
	for _, v := range tempMap {
		if v == arrLength {
			numElems++
		}
	}
	out := reflect.MakeSlice(reflect.TypeOf(arrs[0]), numElems, numElems)
	i := 0
	for key, val := range tempMap {
		if val == arrLength {
			v := reflect.ValueOf(key)
			o := out.Index(i)
			o.Set(v)
			i++
		}
	}

	return out, true
}

// Union returns a slice that contains the unique values of all the input slices
//
// [1, 2, 2, 4, 6] & [2, 4, 5] >> [1, 2, 4, 5, 6]
//
// [1, 1, 3, 4, 5, 6] >> [1, 3, 4, 5, 6]
func Union(arrs ...interface{}) (reflect.Value, bool) {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[interface{}]uint8)
	var kind reflect.Kind
	var kindHasBeenSet bool

	// write the contents of the arrays as keys to the map. The map values don't matter
	for _, arg := range arrs {
		tempArr, ok := Distinct(arg)
		if !ok {
			return reflect.Value{}, ok
		}

		// check to be sure the type hasn't changed
		if kindHasBeenSet && tempArr.Len() > 0 && tempArr.Index(0).Kind() != kind {
			return reflect.Value{}, false
		}
		if tempArr.Len() > 0 {
			kindHasBeenSet = true
			kind = tempArr.Index(0).Kind()
		}

		c := tempArr.Len()
		for idx := 0; idx < c; idx++ {
			tempMap[tempArr.Index(idx).Interface()] = 0
		}
	}

	// the map keys are now unique instances of all of the array contents
	mapLen := len(tempMap)
	out := reflect.MakeSlice(reflect.TypeOf(arrs[0]), mapLen, mapLen)
	i := 0
	for key := range tempMap {
		v := reflect.ValueOf(key)
		o := out.Index(i)
		o.Set(v)
		i++
	}

	return out, true
}

// Difference returns a slice of values that are only present in one of the input slices
//
// [1, 2, 2, 4, 6] & [2, 4, 5] >> [1, 5, 6]
//
// [1, 1, 3, 4, 5, 6] >> [1, 3, 4, 5, 6]
func Difference(arrs ...interface{}) (reflect.Value, bool) {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[interface{}]int)
	var kind reflect.Kind
	var kindHasBeenSet bool

	for _, arg := range arrs {
		tempArr, ok := Distinct(arg)
		if !ok {
			return reflect.Value{}, ok
		}

		// check to be sure the type hasn't changed
		if kindHasBeenSet && tempArr.Len() > 0 && tempArr.Index(0).Kind() != kind {
			return reflect.Value{}, false
		}
		if tempArr.Len() > 0 {
			kindHasBeenSet = true
			kind = tempArr.Index(0).Kind()
		}

		c := tempArr.Len()
		for idx := 0; idx < c; idx++ {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr.Index(idx).Interface()]; ok {
				tempMap[tempArr.Index(idx).Interface()]++
			} else {
				tempMap[tempArr.Index(idx).Interface()] = 1
			}
		}
	}

	// write the final val of the diffMap to an array and return
	numElems := 0
	for _, v := range tempMap {
		if v == 1 {
			numElems++
		}
	}
	out := reflect.MakeSlice(reflect.TypeOf(arrs[0]), numElems, numElems)
	i := 0
	for key, val := range tempMap {
		if val == 1 {
			v := reflect.ValueOf(key)
			o := out.Index(i)
			o.Set(v)
			i++
		}
	}

	return out, true
}

func takeArg(arg interface{}, kind reflect.Kind) (val reflect.Value, ok bool) {
	val = reflect.ValueOf(arg)
	if val.Kind() == kind {
		ok = true
	}
	return
}

/* ***************************************************************
*
* THE SECTIONS BELOW ARE DEPRECATED
*
/* *************************************************************** */

/* ***************************************************************
*
* THIS SECTION IS FOR STRINGS
*
/* *************************************************************** */

// IntersectString finds the intersection of two arrays.
//
// Deprecated: use Intersect instead.
func IntersectString(args ...[]string) []string {
	// create a map to count all the instances of the strings
	arrLength := len(args)
	tempMap := make(map[string]int)
	for _, arg := range args {
		tempArr := DistinctString(arg)
		for idx := range tempArr {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr[idx]]; ok {
				tempMap[tempArr[idx]]++
			} else {
				tempMap[tempArr[idx]] = 1
			}
		}
	}

	// find the keys equal to the length of the input args
	tempArray := make([]string, 0)
	for key, val := range tempMap {
		if val == arrLength {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// IntersectStringArr finds the intersection of two arrays using a multidimensional array as inputs
//
// Deprecated: use Intersect instead.
func IntersectStringArr(arr [][]string) []string {
	// create a map to count all the instances of the strings
	arrLength := len(arr)
	tempMap := make(map[string]int)
	for idx1 := range arr {
		tempArr := DistinctString(arr[idx1])
		for idx2 := range tempArr {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr[idx2]]; ok {
				tempMap[tempArr[idx2]]++
			} else {
				tempMap[tempArr[idx2]] = 1
			}
		}
	}

	// find the keys equal to the length of the input args
	tempArray := make([]string, 0)
	for key, val := range tempMap {
		if val == arrLength {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// UnionString finds the union of two arrays.
//
// Deprecated: use Union instead.
func UnionString(args ...[]string) []string {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[string]uint8)

	// write the contents of the arrays as keys to the map. The map values don't matter
	for _, arg := range args {
		for idx := range arg {
			tempMap[arg[idx]] = 0
		}
	}

	// the map keys are now unique instances of all of the array contents
	tempArray := make([]string, 0)
	for key := range tempMap {
		tempArray = append(tempArray, key)
	}

	return tempArray
}

// UnionStringArr finds the union of two arrays using a multidimensional array as inputs
//
// Deprecated: use Union instead.
func UnionStringArr(arr [][]string) []string {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[string]uint8)

	// write the contents of the arrays as keys to the map. The map values don't matter
	for idx1 := range arr {
		for idx2 := range arr[idx1] {
			tempMap[arr[idx1][idx2]] = 0
		}
	}

	// the map keys are now unique instances of all of the array contents
	tempArray := make([]string, 0)
	for key := range tempMap {
		tempArray = append(tempArray, key)
	}

	return tempArray
}

// DifferenceString finds the difference of two arrays.
//
// Deprecated: use Difference instead.
func DifferenceString(args ...[]string) []string {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[string]int)
	for _, arg := range args {
		tempArr := DistinctString(arg)
		for idx := range tempArr {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr[idx]]; ok {
				tempMap[tempArr[idx]]++
			} else {
				tempMap[tempArr[idx]] = 1
			}
		}
	}

	// write the final val of the diffMap to an array and return
	tempArray := make([]string, 0)
	for key, val := range tempMap {
		if val == 1 {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// DifferenceStringArr finds the difference of two arrays using a multidimensional array as inputs
//
// Deprecated: use Difference instead.
func DifferenceStringArr(arr [][]string) []string {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[string]int)
	for idx1 := range arr {
		tempArr := DistinctString(arr[idx1])
		for idx2 := range tempArr {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr[idx2]]; ok {
				tempMap[tempArr[idx2]]++
			} else {
				tempMap[tempArr[idx2]] = 1
			}
		}
	}

	// write the final val of the diffMap to an array and return
	tempArray := make([]string, 0)
	for key, val := range tempMap {
		if val == 1 {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// DistinctString removes duplicate values from one array.
//
// Deprecated: use Distinct instead.
func DistinctString(arg []string) []string {
	tempMap := make(map[string]uint8)

	for idx := range arg {
		tempMap[arg[idx]] = 0
	}

	tempArray := make([]string, 0)
	for key := range tempMap {
		tempArray = append(tempArray, key)
	}
	return tempArray
}

/* ***************************************************************
*
* THIS SECTION IS FOR uint64's
*
/* *************************************************************** */

// IntersectUint64 finds the intersection of two arrays.
//
// Deprecated: use Intersect instead.
func IntersectUint64(args ...[]uint64) []uint64 {
	// create a map to count all the instances of the strings
	arrLength := len(args)
	tempMap := make(map[uint64]int)
	for _, arg := range args {
		tempArr := DistinctUint64(arg)
		for idx := range tempArr {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr[idx]]; ok {
				tempMap[tempArr[idx]]++
			} else {
				tempMap[tempArr[idx]] = 1
			}
		}
	}

	// find the keys equal to the length of the input args
	tempArray := make([]uint64, 0)
	for key, val := range tempMap {
		if val == arrLength {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// DistinctIntersectUint64 finds the intersection of two arrays of distinct vals.
//
// Deprecated: use Intersect instead.
func DistinctIntersectUint64(args ...[]uint64) []uint64 {
	// create a map to count all the instances of the strings
	arrLength := len(args)
	tempMap := make(map[uint64]int)
	for _, arg := range args {
		for idx := range arg {
			// how many times have we encountered this elem?
			if _, ok := tempMap[arg[idx]]; ok {
				tempMap[arg[idx]]++
			} else {
				tempMap[arg[idx]] = 1
			}
		}
	}

	// find the keys equal to the length of the input args
	tempArray := make([]uint64, 0)
	for key, val := range tempMap {
		if val == arrLength {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

func sortedIntersectUintHelper(a1 []uint64, a2 []uint64) []uint64 {
	intersection := make([]uint64, 0)
	n1 := len(a1)
	n2 := len(a2)
	i := 0
	j := 0
	for i < n1 && j < n2 {
		switch {
		case a1[i] > a2[j]:
			j++
		case a2[j] > a1[i]:
			i++
		default:
			intersection = append(intersection, a1[i])
			i++
			j++
		}
	}
	return intersection
}

// SortedIntersectUint64 finds the intersection of two sorted arrays.
//
// Deprecated: use Intersect instead.
func SortedIntersectUint64(args ...[]uint64) []uint64 {
	// create an array to hold the intersection and write the first array to it
	tempIntersection := args[0]
	argsLen := len(args)

	for k := 1; k < argsLen; k++ {
		// do we have any intersections?
		switch len(tempIntersection) {
		case 0:
			// nope! Give them an empty array!
			return tempIntersection

		default:
			// yup, keep chugging
			tempIntersection = sortedIntersectUintHelper(tempIntersection, args[k])
		}
	}

	return tempIntersection
}

// IntersectUint64Arr finds the intersection of two arrays using a multidimensional array as inputs
//
// Deprecated: use Intersect instead.
func IntersectUint64Arr(arr [][]uint64) []uint64 {
	// create a map to count all the instances of the strings
	arrLength := len(arr)
	tempMap := make(map[uint64]int)
	for idx1 := range arr {
		tempArr := DistinctUint64(arr[idx1])
		for idx2 := range tempArr {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr[idx2]]; ok {
				tempMap[tempArr[idx2]]++
			} else {
				tempMap[tempArr[idx2]] = 1
			}
		}
	}

	// find the keys equal to the length of the input args
	tempArray := make([]uint64, 0)
	for key, val := range tempMap {
		if val == arrLength {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// SortedIntersectUint64Arr finds the intersection of two arrays using a multidimensional array as inputs
//
// Deprecated: use Intersect instead.
func SortedIntersectUint64Arr(arr [][]uint64) []uint64 {
	// create an array to hold the intersection and write the first array to it
	tempIntersection := arr[0]
	argsLen := len(arr)

	for k := 1; k < argsLen; k++ {
		// do we have any intersections?
		switch len(tempIntersection) {
		case 0:
			// nope! Give them an empty array!
			return tempIntersection

		default:
			// yup, keep chugging
			tempIntersection = sortedIntersectUintHelper(tempIntersection, arr[k])
		}
	}

	return tempIntersection
}

// DistinctIntersectUint64Arr finds the intersection of two distinct arrays using a multidimensional array as inputs
//
// Deprecated: use Distinct instead.
func DistinctIntersectUint64Arr(arr [][]uint64) []uint64 {
	// create a map to count all the instances of the strings
	arrLength := len(arr)
	tempMap := make(map[uint64]int)
	for idx1 := range arr {
		for idx2 := range arr[idx1] {
			// how many times have we encountered this elem?
			if _, ok := tempMap[arr[idx1][idx2]]; ok {
				tempMap[arr[idx1][idx2]]++
			} else {
				tempMap[arr[idx1][idx2]] = 1
			}
		}
	}

	// find the keys equal to the length of the input args
	tempArray := make([]uint64, 0)
	for key, val := range tempMap {
		if val == arrLength {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// UnionUint64 finds the union of two arrays.
//
// Deprecated: use Union instead.
func UnionUint64(args ...[]uint64) []uint64 {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[uint64]uint8)

	// write the contents of the arrays as keys to the map. The map values don't matter
	for _, arg := range args {
		for idx := range arg {
			tempMap[arg[idx]] = 0
		}
	}

	// the map keys are now unique instances of all of the array contents
	tempArray := make([]uint64, 0)
	for key := range tempMap {
		tempArray = append(tempArray, key)
	}

	return tempArray
}

// UnionUint64Arr finds the union of two arrays using a multidimensional array as inputs
//
// Deprecated: use Union instead.
func UnionUint64Arr(arr [][]uint64) []uint64 {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[uint64]uint8)

	// write the contents of the arrays as keys to the map. The map values don't matter
	for idx1 := range arr {
		for idx2 := range arr[idx1] {
			tempMap[arr[idx1][idx2]] = 0
		}
	}

	// the map keys are now unique instances of all of the array contents
	tempArray := make([]uint64, 0)
	for key := range tempMap {
		tempArray = append(tempArray, key)
	}

	return tempArray
}

// DifferenceUint64 finds the difference of two arrays.
//
// Deprecated: use Difference instead.
func DifferenceUint64(args ...[]uint64) []uint64 {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[uint64]int)
	for _, arg := range args {
		tempArr := DistinctUint64(arg)
		for idx := range tempArr {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr[idx]]; ok {
				tempMap[tempArr[idx]]++
			} else {
				tempMap[tempArr[idx]] = 1
			}
		}
	}

	// write the final val of the diffMap to an array and return
	tempArray := make([]uint64, 0)
	for key, val := range tempMap {
		if val == 1 {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// DifferenceUint64Arr finds the difference of two arrays using a multidimensional array as inputs.
//
// Deprecated: use Difference instead.
func DifferenceUint64Arr(arr [][]uint64) []uint64 {
	// create a temporary map to hold the contents of the arrays
	tempMap := make(map[uint64]int)
	for idx1 := range arr {
		tempArr := DistinctUint64(arr[idx1])
		for idx2 := range tempArr {
			// how many times have we encountered this elem?
			if _, ok := tempMap[tempArr[idx2]]; ok {
				tempMap[tempArr[idx2]]++
			} else {
				tempMap[tempArr[idx2]] = 1
			}
		}
	}

	// write the final val of the diffMap to an array and return
	tempArray := make([]uint64, 0)
	for key, val := range tempMap {
		if val == 1 {
			tempArray = append(tempArray, key)
		}
	}

	return tempArray
}

// DistinctUint64 removes duplicate values from one array.
//
// Deprecated: use Distinct instead.
func DistinctUint64(arg []uint64) []uint64 {
	tempMap := make(map[uint64]uint8)

	for idx := range arg {
		tempMap[arg[idx]] = 0
	}

	tempArray := make([]uint64, 0)
	for key := range tempMap {
		tempArray = append(tempArray, key)
	}
	return tempArray
}
