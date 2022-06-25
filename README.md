# dag

[![Build Status](https://github.com/pipego/dag/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/pipego/dag/actions?query=workflow%3Aci)
[![codecov](https://codecov.io/gh/pipego/dag/branch/main/graph/badge.svg?token=t31YICk0ek)](https://codecov.io/gh/pipego/dag)
[![License](https://img.shields.io/github/license/pipego/dag.svg)](https://github.com/pipego/dag/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/pipego/dag.svg)](https://github.com/pipego/dag/tags)



## Introduction

*dag* is the DAG of [pipego](https://github.com/pipego) written in Go.



## Prerequisites

- Go >= 1.18.0



## Example

```go
var r runner.Runner

r.AddVertex("task1", func(args []string) error {
	fmt.Printf("name %s args %s\n", args[0], args[1])
	return nil
}, []string{"echo", "task1"})

r.AddVertex("task2", func(args []string) error {
	fmt.Printf("name %s args %s\n", args[0], args[1])
	return nil
}, []string{"echo", "task2"})

r.AddVertex("task3", func(args []string) error {
	fmt.Printf("name %s args %s\n", args[0], args[1])
	return nil
}, []string{"echo", "task3"})

r.AddEdge("task1", "task3")
r.AddEdge("task2", "task3")

fmt.Printf("the runner terminated with: %v\n", r.Run())
```



## License

Project License can be found [here](LICENSE).



## Reference

- [drone-dag](https://github.com/drone/dag)
- [drone-pipeline](https://docs.drone.io/pipeline/overview/)
- [machinery](https://github.com/RichardKnop/machinery/blob/master/v2/example/go-redis/main.go)
- [wiki-dag](https://en.wikipedia.org/wiki/Directed_acyclic_graph)
