# Built-in Function {#functions}
---

Function parameter description:

- In function arguments, the anonymous argument (`_`) refers to the original input text data
- json path, expressed directly as `x.y.z`, without any other modifications. For example, `{"a":{"first":2.3, "second":2, "third":"abc", "forth":true}, "age":47}`, where the json path is `a.thrid` to indicate that the data to be manipulated is `abc`
- The relative order of all function arguments is fixed, and the engine will check it concretely
- All of the `key` parameters mentioned below refer to the `key` generated after the initial extraction (via `grok()` or `json()`)
- The path of the json to be processed, supports the writing of identifiers, and cannot use strings. If you are generating new keys, you need to use strings

## Function List {#function-list}

{{.PipelineFuncs}}
