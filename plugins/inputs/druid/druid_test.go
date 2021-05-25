package druid

import "testing"

func TestExtract(t *testing.T) {

	const data = `
[
  {
    "feed": "metrics",
    "timestamp": "2020-06-24T06:23:03.292Z",
    "service": "druid/middleManager",
    "host": "localhost:8091",
    "version": "0.18.1",
    "metric": "jvm/mem/max",
    "value": 67108864,
    "memKind": "heap"
  },
  {
    "feed": "metrics",
    "timestamp": "2020-06-24T06:23:03.293Z",
    "service": "druid/middleManager",
    "host": "localhost:8091",
    "version": "0.18.1",
    "metric": "jvm/mem/committed",
    "value": 67108864,
    "memKind": "heap"
  },
  {
    "feed": "metrics",
    "timestamp": "2020-06-24T06:23:03.294Z",
    "service": "druid/middleManager",
    "host": "localhost:8091",
    "version": "0.18.1",
    "metric": "jvm/mem/used",
    "value": 45955760,
    "memKind": "heap"
  },
  {
    "feed": "metrics",
    "timestamp": "2020-06-24T06:23:03.294Z",
    "service": "druid/middleManager",
    "host": "localhost:8091",
    "version": "0.18.1",
    "metric": "jvm/mem/init",
    "value": 67108864,
    "memKind": "heap"
  }
]
`
	t.Logf("%v\n", extract([]byte(data), nil))
}
