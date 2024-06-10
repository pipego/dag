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
	Count = 5000
	Width = 500
)

type Task struct {
	Name     string
	File     runner.File
	Params   []runner.Param
	Commands []string
	Width    int64
	Language runner.Language
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
	Width    int64
	Language runner.Language
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
			Width:    Width,
			Language: runner.Language{},
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
			Width:    Width,
			Language: runner.Language{},
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
			Width:    Width,
			Language: runner.Language{},
			Depends:  []string{"task1", "task2"},
		},
	}
)

func main() {
	var r runner.Runner

	l := runner.Log{
		Line: make(chan *runner.Line, Count),
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

	close(l.Line)

	fmt.Println("done.")
}

func initDag() Dag {
	var dag Dag

	for index := range tasks {
		d := Vertex{
			Name:     tasks[index].Name,
			File:     tasks[index].File,
			Params:   tasks[index].Params,
			Commands: tasks[index].Commands,
			Width:    tasks[index].Width,
			Language: tasks[index].Language,
		}
		dag.Vertex = append(dag.Vertex, d)

		for _, dep := range tasks[index].Depends {
			e := Edge{
				From: dep,
				To:   tasks[index].Name,
			}
			dag.Edge = append(dag.Edge, e)
		}
	}

	return dag
}

func runDag(run runner.Runner, dag Dag, log runner.Log) error {
	for i := range dag.Vertex {
		run.AddVertex(dag.Vertex[i].Name, runHelper, dag.Vertex[i].File, dag.Vertex[i].Params, dag.Vertex[i].Commands,
			dag.Vertex[i].Width, dag.Vertex[i].Language)
	}

	for _, edge := range dag.Edge {
		run.AddEdge(edge.From, edge.To)
	}

	return run.Run(log)
}

func runHelper(_ string, _ runner.File, params []runner.Param, cmds []string, _ int64, _ runner.Language, log runner.Log) error {
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

func routine(scanner *bufio.Scanner, log runner.Log) {
	done := make(chan bool)

	go func(scanner *bufio.Scanner, log runner.Log, done chan bool) {
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

func printer(log runner.Log, done chan<- bool) {
	for range tasks {
		line := <-log.Line
		fmt.Println(line)
	}

	done <- true
}
