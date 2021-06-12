# LRU Cache

[![Report Card](https://goreportcard.com/badge/github.com/fahmifan/lrucache)](https://goreportcard.com/report/github.com/fahmifan/lrucache)

Implement LRU Cache in Golang with concurrency safety

## Benchmark
```
go test -benchmem -run=^$ -bench ^(BenchmarkLRUCacher)$ github.com/fahmifan/lrucache

goos: linux
goarch: amd64
pkg: github.com/fahmifan/lrucache
cpu: Intel(R) Core(TM) i5-7400 CPU @ 3.00GHz
BenchmarkLRUCacher/Put-4         	 2777412	       415.7 ns/op	      89 B/op	       4 allocs/op
BenchmarkLRUCacher/Get-4         	 9061254	       130.3 ns/op	      16 B/op	       1 allocs/op
BenchmarkLRUCacher/Del-4         	11411762	       105.7 ns/op	      12 B/op	       1 allocs/op
PASS
ok  	github.com/fahmifan/lrucache	4.228s
```