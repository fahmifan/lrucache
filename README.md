# LRU Cache

[![Report Card](https://goreportcard.com/badge/github.com/fahmifan/lrucache)](https://goreportcard.com/report/github.com/fahmifan/lrucache)

Implement LRU Cache in Golang with concurrency safety

## Benchmark
### V1

```
go test -benchmem -run=^$ -bench ^(BenchmarkLRUCacher)$ github.com/fahmifan/lrucache

goos: linux
goarch: amd64
pkg: github.com/fahmifan/lrucache
cpu: Intel(R) Core(TM) i5-7400 CPU @ 3.00GHz
BenchmarkLRUCacher/Put-4         	 2735919	       426.2  ns/op	      89 B/op	       4 allocs/op
BenchmarkLRUCacher/Get-4         	 9175941	       130.2  ns/op	      16 B/op	       1 allocs/op
BenchmarkLRUCacher/Del-4         	11998616	        99.73 ns/op	      12 B/op	       1 allocs/op
PASS
ok  	github.com/fahmifan/lrucache	4.245s
```

### V2
```
go test -benchmem -run=^$ -bench ^BenchmarkLRUCacher$ github.com/fahmifan/lrucache/v2

goos: linux
goarch: amd64
pkg: github.com/fahmifan/lrucache/v2
cpu: Intel(R) Core(TM) i5-7400 CPU @ 3.00GHz
BenchmarkLRUCacher/Put-4         	 1388715	      818.9  ns/op	     210 B/op	       5 allocs/op
BenchmarkLRUCacher/Get-4         	 9343468	      131.4  ns/op	      16 B/op	       1 allocs/op
BenchmarkLRUCacher/Del-4         	12120594	      99.69  ns/op	      12 B/op	       1 allocs/op
PASS
ok  	github.com/fahmifan/lrucache/v2	4.702s
```