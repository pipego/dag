package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"time"

	"github.com/pipego/dag/runner"
)

type Task struct {
	Name     string
	Commands []string
	Depends  []string
}

type Dag struct {
	Vertex []Vertex
	Edge   []Edge
}

type Vertex struct {
	Name string
	Run  []string
}

type Edge struct {
	From string
	To   string
}

var (
	tasks = []Task{
		{
			Name:     "task1",
			Commands: []string{"echo", "task1"},
			Depends:  []string{},
		},
		{
			Name:     "task2",
			Commands: []string{"echo", "task2"},
			Depends:  []string{},
		},
		{
			Name:     "task3",
			Commands: []string{"echo", "task3"},
			Depends:  []string{"task1", "task2"},
		},
	}
)

func main() {
	var r runner.Runner

	d := initDag()
	_ = runDag(r, d)
}

func initDag() Dag {
	var dag Dag

	for _, task := range tasks {
		d := Vertex{
			Name: task.Name,
			Run:  task.Commands,
		}
		dag.Vertex = append(dag.Vertex, d)

		for _, dep := range task.Depends {
			e := Edge{
				From: dep,
				To:   task.Name,
			}
			dag.Edge = append(dag.Edge, e)
		}
	}

	return dag
}

func runDag(run runner.Runner, dag Dag) error {
	for _, vertex := range dag.Vertex {
		run.AddVertex(vertex.Name, routine, vertex.Run)
	}

	for _, edge := range dag.Edge {
		run.AddEdge(edge.From, edge.To)
	}

	return run.Run()
}

func routine(_ string, args []string) runner.Result {
	var a []string
	var n string
	var out []runner.Output

	n, _ = exec.LookPath(args[0])
	a = args[1:]

	cmd := exec.Command(n, a...)
	stdout, _ := cmd.StdoutPipe()

	_ = cmd.Start()

	scanner := bufio.NewScanner(stdout)
	p := 1

	for scanner.Scan() {
		o := runner.Output{Pos: int64(p), Time: time.Now().Unix(), Message: scanner.Text()}
		fmt.Println(o)
		out = append(out, o)
		p += 1
	}

	_ = cmd.Wait()

	return runner.Result{out, ""}
}
