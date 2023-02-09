### `sample()` {#fn-sample}

Function prototype: `fn sample(p)`

Function description: Choose to collect/discard data with probability p.

Function parameters:
- `p`: the probability that the sample function returns true, the value range is [0, 1]

Example:

```python
# process script
if !sample(0.3) { # sample(0.3) indicates that the sampling rate is 30%, that is, it returns true with a 30% probability, and 70% of the data will be discarded here
   drop() # mark the data to be discarded
   exit() # Exit the follow-up processing process
}
```
