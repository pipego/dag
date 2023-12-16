package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

const (
	Timeout = 10 * time.Second
	Width   = 500
)

func TestZero(t *testing.T) {
	var r Runner

	log := Log{
		Line: make(chan *Line),
	}

	res := make(chan error)

	go func() { res <- r.Run(log) }()

	_, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	select {
	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}
	case <-time.After(Timeout):
		t.Error("timeout")
	}
}

func TestOne(t *testing.T) {
	var r Runner

	err := errors.New("error")
	r.AddVertex("one", func(string, File, []Param, []string, int64, Log) error { return err },
		File{}, []Param{}, []string{}, Width)

	log := Log{
		Line: make(chan *Line),
	}

	res := make(chan error)

	go func() { res <- r.Run(log) }()

	_, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	select {
	case err := <-res:
		if want, have := err, err; !errors.Is(want, have) {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(Timeout):
		t.Error("timeout")
	}
}

func TestManyNoDeps(t *testing.T) {
	var r Runner

	err := errors.New("error")
	r.AddVertex("one", func(string, File, []Param, []string, int64, Log) error { return err },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("two", func(string, File, []Param, []string, int64, Log) error { return err },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("three", func(string, File, []Param, []string, int64, Log) error { return err },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("four", func(string, File, []Param, []string, int64, Log) error { return err },
		File{}, []Param{}, []string{}, Width)

	log := Log{
		Line: make(chan *Line),
	}

	res := make(chan error)

	go func() { res <- r.Run(log) }()

	_, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	select {
	case err := <-res:
		if want, have := err, err; !errors.Is(want, have) {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(Timeout):
		t.Error("timeout")
	}
}

func TestManyWithCycle(t *testing.T) {
	var r Runner

	r.AddVertex("one", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("two", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("three", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("four", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("three", "one")

	log := Log{
		Line: make(chan *Line),
	}

	res := make(chan error)

	go func() { res <- r.Run(log) }()

	_, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	select {
	case err := <-res:
		if want, have := errCycleDetected, err; !errors.Is(want, have) {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(Timeout):
		t.Error("timeout")
	}
}

func TestInvalidToVertex(t *testing.T) {
	var r Runner

	r.AddVertex("one", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("two", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("three", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("four", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("three", "definitely-not-a-valid-vertex")

	log := Log{
		Line: make(chan *Line),
	}

	res := make(chan error)

	go func() { res <- r.Run(log) }()

	_, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	select {
	case err := <-res:
		if want, have := errMissingVertex, err; !errors.Is(want, have) {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(Timeout):
		t.Error("timeout")
	}
}

func TestInvalidFromVertex(t *testing.T) {
	var r Runner

	r.AddVertex("one", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("two", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("three", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)
	r.AddVertex("four", func(string, File, []Param, []string, int64, Log) error { return nil },
		File{}, []Param{}, []string{}, Width)

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("definitely-not-a-valid-vertex", "three")

	log := Log{
		Line: make(chan *Line),
	}

	res := make(chan error)

	go func() { res <- r.Run(log) }()

	_, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	select {
	case err := <-res:
		if want, have := errMissingVertex, err; !errors.Is(want, have) {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(Timeout):
		t.Error("timeout")
	}
}

func TestManyWithDepsSuccess(t *testing.T) {
	var r Runner

	res := make(chan string, 7)
	r.AddVertex("one", func(string, File, []Param, []string, int64, Log) error {
		res <- "one"
		return nil
	}, File{}, []Param{}, []string{}, Width)
	r.AddVertex("two", func(string, File, []Param, []string, int64, Log) error {
		res <- "two"
		return nil
	}, File{}, []Param{}, []string{}, Width)
	r.AddVertex("three", func(string, File, []Param, []string, int64, Log) error {
		res <- "three"
		return nil
	}, File{}, []Param{}, []string{}, Width)
	r.AddVertex("four", func(string, File, []Param, []string, int64, Log) error {
		res <- "four"
		return nil
	}, File{}, []Param{}, []string{}, Width)
	r.AddVertex("five", func(string, File, []Param, []string, int64, Log) error {
		res <- "five"
		return nil
	}, File{}, []Param{}, []string{}, Width)
	r.AddVertex("six", func(string, File, []Param, []string, int64, Log) error {
		res <- "six"
		return nil
	}, File{}, []Param{}, []string{}, Width)
	r.AddVertex("seven", func(string, File, []Param, []string, int64, Log) error {
		res <- "seven"
		return nil
	}, File{}, []Param{}, []string{}, Width)

	r.AddEdge("one", "two")
	r.AddEdge("one", "three")
	r.AddEdge("two", "four")
	r.AddEdge("two", "seven")
	r.AddEdge("five", "six")

	log := Log{
		Line: make(chan *Line),
	}

	err := make(chan error)

	go func() { err <- r.Run(log) }()

	_, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	select {
	case err := <-err:
		if want, have := error(nil), err; !errors.Is(want, have) {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(Timeout):
		t.Error("timeout")
	}

	results := make([]string, 7)
	timeout := time.After(Timeout)

	for i := range results {
		select {
		case results[i] = <-res:
		case <-timeout:
			t.Error("timeout")
		}
	}

	checkOrder("one", "two", results, t)
	checkOrder("one", "three", results, t)
	checkOrder("two", "four", results, t)
	checkOrder("two", "seven", results, t)
	checkOrder("five", "six", results, t)
}

func checkOrder(from, to string, results []string, t *testing.T) {
	var fromIndex, toIndex int

	for i := range results {
		if results[i] == from {
			fromIndex = i
		}
		if results[i] == to {
			toIndex = i
		}
	}

	if fromIndex > toIndex {
		t.Errorf("from vertex: %s came after to vertex: %s", from, to)
	}
}
