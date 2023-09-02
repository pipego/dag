package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pipego/dag/runner"
)

const (
	Log = 5000
)

type Task struct {
	Name     string
	File     runner.File
	Params   []runner.Param
	Commands []string
	Livelog  int64
	Depends  []string
}

type Dag struct {
	Vertex []Vertex
	Edge   []Edge
}

type Vertex struct {
	Name     string
	File     runner.File
	Params   []runner.Param
	Commands []string
	Livelog  int64
}

type Edge struct {
	From string
	To   string
}

var (
	tasks = []Task{
		{
			Name: "task1",
			File: runner.File{Content: "", Gzip: false},
			Params: []runner.Param{
				{
					Name:  "env1",
					Value: "val1",
				},
			},
			Commands: []string{"echo", "$env1"},
			Livelog:  Log,
			Depends:  []string{},
		},
		{
			Name: "task2",
			File: runner.File{Content: "", Gzip: false},
			Params: []runner.Param{
				{
					Name:  "env2",
					Value: "val2",
				},
			},
			Commands: []string{"echo", "$env2"},
			Livelog:  Log,
			Depends:  []string{},
		},
		{
			Name: "task3",
			File: runner.File{Content: "", Gzip: false},
			Params: []runner.Param{
				{
					Name:  "env3",
					Value: "val3",
				},
			},
			Commands: []string{"echo", "$env3"},
			Livelog:  Log,
			Depends:  []string{"task1", "task2"},
		},
	}
)

func main() {
	var r runner.Runner

	l := runner.Livelog{
		Error: make(chan error, Log),
		Line:  make(chan *runner.Line, Log),
	}

	d := initDag()
	_ = runDag(r, d, l)

	done := make(chan bool, 1)
	go printer(l, done)

L:
	for {
		select {
		case <-done:
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
			Name:     task.Name,
			File:     task.File,
			Params:   task.Params,
			Commands: task.Commands,
			Livelog:  task.Livelog,
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
		run.AddVertex(vertex.Name, runHelper, vertex.File, vertex.Params, vertex.Commands, vertex.Livelog)
	}

	for _, edge := range dag.Edge {
		run.AddEdge(edge.From, edge.To)
	}

	return run.Run(log)
}

func runHelper(_ string, _ runner.File, params []runner.Param, cmds []string, _ int64, log runner.Livelog) error {
	var a, args []string
	var n string

	args = buildArgs(cmds)

	n, _ = exec.LookPath(args[0])
	a = args[1:]

	cmd := exec.Command(n, a...)
	cmd.Env = append(cmd.Environ(), buildEnvs(params)...)
	stdout, _ := cmd.StdoutPipe()

	_ = cmd.Start()

	scanner := bufio.NewScanner(stdout)
	routine(scanner, log)

	go func(cmd *exec.Cmd) {
		_ = cmd.Wait()
	}(cmd)

	return nil
}

func buildArgs(cmds []string) []string {
	return []string{"bash", "-c", strings.Join(cmds, " ")}
}

func buildEnvs(params []runner.Param) []string {
	var buf []string

	for _, item := range params {
		buf = append(buf, item.Name+"="+item.Value)
	}

	return buf
}

func routine(scanner *bufio.Scanner, log runner.Livelog) {
	done := make(chan bool)

	go func(scanner *bufio.Scanner, log runner.Livelog, done chan bool) {
		p := 1
		for scanner.Scan() {
			l := runner.Line{Pos: int64(p), Time: time.Now().Unix(), Message: scanner.Text()}
			log.Line <- &l
			p += 1
		}
		done <- true
	}(scanner, log, done)

	<-done
}

func printer(log runner.Livelog, done chan<- bool) {
	for range tasks {
		line := <-log.Line
		fmt.Println(line)
	}

	done <- true
}
