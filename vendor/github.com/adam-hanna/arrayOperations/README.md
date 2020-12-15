[![Build Status](https://travis-ci.org/adam-hanna/arrayOperations.svg?branch=master)](https://travis-ci.org/adam-hanna/arrayOperations) [![Coverage Status](https://coveralls.io/repos/github/adam-hanna/arrayOperations/badge.svg?branch=master)](https://coveralls.io/github/adam-hanna/arrayOperations?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/adam-hanna/arrayOperations)](https://goreportcard.com/report/github.com/adam-hanna/arrayOperations) [![GoDoc](https://godoc.org/github.com/adam-hanna/arrayOperations?status.svg)](https://godoc.org/github.com/adam-hanna/arrayOperations)

# arrayOperations
Small library for performing union, intersect, difference and distinct operations on slices in goLang

I don't promise that these are optimized (not even close!), but they work :)

## Quickstart
~~~ go
var a = []int{1, 1, 2, 3}
var b = []int{2, 4}

z, ok := Difference(a, b)
if !ok {
	fmt.Println("Cannot find difference")
}

slice, ok := z.Interface().([]int)
if !ok {
	fmt.Println("Cannot convert to slice")
}
fmt.Println(slice, reflect.TypeOf(slice)) // [1, 3, 4] []int
~~~

## API
### [Difference](https://godoc.org/github.com/adam-hanna/arrayOperations#Difference)
~~~ go
func Difference(arrs ...interface{}) (reflect.Value, bool)
~~~
 Difference returns a slice of values that are only present in one of the input slices

[1, 2, 2, 4, 6] & [2, 4, 5] >> [1, 5, 6]

[1, 1, 3, 4, 5, 6] >> [1, 3, 4, 5, 6] 
~~~ go
// EXAMPLE
var a = []int{1, 1, 2, 3}
var b = []int{2, 4}

z, ok := Difference(a, b)
if !ok {
    fmt.Println("Cannot find difference")
}

slice, ok := z.Interface().([]int)
if !ok {
    fmt.Println("Cannot convert to slice")
}
fmt.Println(slice, reflect.TypeOf(slice)) // [1, 3, 4] []int
~~~

### [Distinct](https://godoc.org/github.com/adam-hanna/arrayOperations#Distinct)
~~~ go
func Distinct(arr interface{}) (reflect.Value, bool)
~~~

Distinct returns the unique vals of a slice

[1, 1, 2, 3] >> [1, 2, 3] 

~~~ go
// EXAMPLE
var a = []int{1, 1, 2, 3}

z, ok := Distinct(a)
if !ok {
    fmt.Println("Cannot find distinct")
}

slice, ok := z.Interface().([]int)
if !ok {
    fmt.Println("Cannot convert to slice")
}
fmt.Println(slice, reflect.TypeOf(slice)) // [1, 2, 3] []int
~~~

### [Intersect](https://godoc.org/github.com/adam-hanna/arrayOperations#Intersect)
~~~ go
func Intersect(arrs ...interface{}) (reflect.Value, bool)
~~~

Intersect returns a slice of values that are present in all of the input slices

[1, 1, 3, 4, 5, 6] & [2, 3, 6] >> [3, 6]

[1, 1, 3, 4, 5, 6] >> [1, 3, 4, 5, 6] 

~~~ go
// EXAMPLE
var a = []int{1, 1, 2, 3}
var b = []int{2, 4}

z, ok := Intersect(a, b)
if !ok {
    fmt.Println("Cannot find intersect")
}

slice, ok := z.Interface().([]int)
if !ok {
    fmt.Println("Cannot convert to slice")
}
fmt.Println(slice, reflect.TypeOf(slice)) // [2] []int
~~~

### [Union](https://godoc.org/github.com/adam-hanna/arrayOperations#Union)
~~~ go
func Union(arrs ...interface{}) (reflect.Value, bool)
~~~

Union returns a slice that contains the unique values of all the input slices

[1, 2, 2, 4, 6] & [2, 4, 5] >> [1, 2, 4, 5, 6]

[1, 1, 3, 4, 5, 6] >> [1, 3, 4, 5, 6] 

~~~ go
// EXAMPLE
var a = []int{1, 1, 2, 3}
var b = []int{2, 4}

z, ok := Union(a, b)
if !ok {
    fmt.Println("Cannot find union")
}

slice, ok := z.Interface().([]int)
if !ok {
    fmt.Println("Cannot convert to slice")
}
fmt.Println(slice, reflect.TypeOf(slice)) // [1, 2, 3, 4] []int
~~~

## License
MIT
