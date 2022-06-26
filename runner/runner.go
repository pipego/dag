// Package runner implements a directed acyclic graph task runner with deterministic teardown.
// it is similar to package errgroup, in that it runs multiple tasks in parallel and returns
// the first error it encounters. Users define a Runner as a set vertices (functions) and edges
// between them. During Run, the directed acyclec graph will be validated and each vertex
// will run in parallel as soon as it's dependencies have been resolved. The Runner will only
// return after all running goroutines have stopped.
package runner

import (
	"github.com/pkg/errors"
)

// Runner collects functions and arranges them as vertices and edges of a directed acyclic graph.
// Upon validation of the graph, functions are run in parallel topological order. The zero value
// is useful.
type Runner struct {
	fn    map[string]function
	graph map[string][]string
}

type Result struct {
	Output []Output
	Error  string
}

type Output struct {
	Pos     int64
	Time    int64
	Message string
}

type function struct {
	args []string
	name func(string, []string) Result
}

type result struct {
	name   string
	result Result
}

var errMissingVertex = errors.New("missing vertex")
var errCycleDetected = errors.New("dependency cycle detected")

// AddVertex adds a function as a vertex in the graph. Only functions which have been added in this
// way will be executed during Run.
func (r *Runner) AddVertex(name string, fn func(string, []string) Result, args []string) {
	if r.fn == nil {
		r.fn = make(map[string]function)
	}

	r.fn[name] = function{
		args: args,
		name: fn,
	}
}

// AddEdge establishes a dependency between two vertices in the graph. Both from and to must exist
// in the graph, or Run will err. The vertex at from will execute before the vertex at to.
func (r *Runner) AddEdge(from, to string) {
	if r.graph == nil {
		r.graph = make(map[string][]string)
	}

	r.graph[from] = append(r.graph[from], to)
}

// Run will validate that all edges in the graph point to existing vertices, and that there are
// no dependency cycles. After validation, each vertex will be run, deterministically, in parallel
// topological order. If any vertex returns an error, no more vertices will be scheduled and
// Run will exit and return that error once all in-flight functions finish execution.
func (r *Runner) Run() error {
	var err error

	// sanity check
	if len(r.fn) == 0 {
		return nil
	}

	// count how many deps each vertex has
	deps := make(map[string]int)
	for vertex, edges := range r.graph {
		// every vertex along every edge must have an associated fn
		if _, ok := r.fn[vertex]; !ok {
			return errMissingVertex
		}
		for _, vertex := range edges {
			if _, ok := r.fn[vertex]; !ok {
				return errMissingVertex
			}
			deps[vertex]++
		}
	}

	if r.detectCycles() {
		return errCycleDetected
	}

	resc := make(chan result, len(r.fn))
	running := 0

	// start any vertex that has no deps
	for name := range r.fn {
		if deps[name] == 0 {
			running++
			r.start(name, r.fn[name], resc)
		}
	}

	// wait for all running work to complete
	for running > 0 {
		res := <-resc
		running--

		// capture the first error
		if res.result.Error != "" && err == nil {
			err = errors.New(res.result.Error)
		}

		// don't enqueue any more work on if there's been an error
		if err != nil {
			continue
		}

		// start any vertex whose deps are fully resolved
		for _, vertex := range r.graph[res.name] {
			deps[vertex]--
			if deps[vertex] == 0 {
				running++
				r.start(vertex, r.fn[vertex], resc)
			}
		}
	}

	return err
}

func (r *Runner) detectCycles() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for vertex := range r.graph {
		if !visited[vertex] {
			if r.detectCyclesHelper(vertex, visited, recStack) {
				return true
			}
		}
	}

	return false
}

func (r *Runner) detectCyclesHelper(vertex string, visited, recStack map[string]bool) bool {
	visited[vertex] = true
	recStack[vertex] = true

	for _, v := range r.graph[vertex] {
		// only check cycles on a vertex one time
		if !visited[v] {
			if r.detectCyclesHelper(v, visited, recStack) {
				return true
			}
			// if we've visited this vertex in this recursion stack, then we have a cycle
		} else if recStack[v] {
			return true
		}
	}

	recStack[vertex] = false

	return false
}

func (r *Runner) start(name string, fn function, resc chan<- result) {
	go func() {
		resc <- result{
			name:   name,
			result: fn.name(name, fn.args),
		}
	}()
}
