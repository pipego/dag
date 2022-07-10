package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/pipego/dag/runner"
)

const (
	LIVELOG = 5000
	TIMEOUT = 10 * time.Second
)

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

func main() {
	var r runner.Runner

	l := runner.Livelog{
		Error: make(chan error, LIVELOG),
		Line:  make(chan *runner.Line, LIVELOG),
	}

	d := initDag()
	_ = runDag(r, d, l)

	done := make(chan bool, 1)
	go printer(l, done)

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

L:
	for {
		select {
		case <-done:
			break L
		case <-ctx.Done():
			break L
		}
	}

	close(l.Error)
	close(l.Line)

	fmt.Println("done.")
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

func runDag(run runner.Runner, dag Dag, log runner.Livelog) error {
	for _, vertex := range dag.Vertex {
		run.AddVertex(vertex.Name, runHelper, vertex.Run)
	}

	for _, edge := range dag.Edge {
		run.AddEdge(edge.From, edge.To)
	}

	return run.Run(log)
}

func runHelper(_ string, args []string, log runner.Livelog) error {
	var a []string
	var n string

	n, _ = exec.LookPath(args[0])
	a = args[1:]

	cmd := exec.Command(n, a...)
	stdout, _ := cmd.StdoutPipe()

	_ = cmd.Start()

	scanner := bufio.NewScanner(stdout)
	routine(scanner, log)

	go func() {
		_ = cmd.Wait()
	}()

	return nil
}

func routine(scanner *bufio.Scanner, log runner.Livelog) {
	go func() {
		p := 1
		for scanner.Scan() {
			l := runner.Line{Pos: int64(p), Time: time.Now().Unix(), Message: scanner.Text()}
			log.Line <- &l
			p += 1
		}
	}()
}

func printer(log runner.Livelog, done chan<- bool) {
	for range tasks {
		line := <-log.Line
		fmt.Println(line)
	}

	done <- true
}
