package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

const (
	TIMEOUT = 10 * time.Second
)

func TestZero(t *testing.T) {
	var r Runner

	_, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	log := Livelog{
		Error: make(chan error),
		Line:  make(chan *Line),
	}
	res := make(chan error)

	go func() { res <- r.Run(log) }()

	select {
	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestOne(t *testing.T) {
	var r Runner

	_, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	err := errors.New("error")
	r.AddVertex("one", func(string, []string, Livelog) error { return err }, []string{})

	log := Livelog{
		Error: make(chan error),
		Line:  make(chan *Line),
	}
	res := make(chan error)

	go func() { res <- r.Run(log) }()

	select {
	case err := <-res:
		if want, have := err, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestManyNoDeps(t *testing.T) {
	var r Runner

	_, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	err := errors.New("error")
	r.AddVertex("one", func(string, []string, Livelog) error { return err }, []string{})
	r.AddVertex("two", func(string, []string, Livelog) error { return err }, []string{})
	r.AddVertex("three", func(string, []string, Livelog) error { return err }, []string{})
	r.AddVertex("four", func(string, []string, Livelog) error { return err }, []string{})

	log := Livelog{
		Error: make(chan error),
		Line:  make(chan *Line),
	}
	res := make(chan error)

	go func() { res <- r.Run(log) }()

	select {
	case err := <-res:
		if want, have := err, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestManyWithCycle(t *testing.T) {
	var r Runner

	_, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	r.AddVertex("one", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("two", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("three", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("four", func(string, []string, Livelog) error { return nil }, []string{})

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("three", "one")

	log := Livelog{
		Error: make(chan error),
		Line:  make(chan *Line),
	}
	res := make(chan error)

	go func() { res <- r.Run(log) }()

	select {
	case err := <-res:
		if want, have := errCycleDetected, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestInvalidToVertex(t *testing.T) {
	var r Runner

	_, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	r.AddVertex("one", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("two", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("three", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("four", func(string, []string, Livelog) error { return nil }, []string{})

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("three", "definitely-not-a-valid-vertex")

	log := Livelog{
		Error: make(chan error),
		Line:  make(chan *Line),
	}
	res := make(chan error)

	go func() { res <- r.Run(log) }()

	select {
	case err := <-res:
		if want, have := errMissingVertex, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestInvalidFromVertex(t *testing.T) {
	var r Runner

	_, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	r.AddVertex("one", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("two", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("three", func(string, []string, Livelog) error { return nil }, []string{})
	r.AddVertex("four", func(string, []string, Livelog) error { return nil }, []string{})

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("definitely-not-a-valid-vertex", "three")

	log := Livelog{
		Error: make(chan error),
		Line:  make(chan *Line),
	}
	res := make(chan error)

	go func() { res <- r.Run(log) }()

	select {
	case err := <-res:
		if want, have := errMissingVertex, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestManyWithDepsSuccess(t *testing.T) {
	var r Runner

	_, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	res := make(chan string, 7)
	r.AddVertex("one", func(string, []string, Livelog) error {
		res <- "one"
		return nil
	}, []string{})
	r.AddVertex("two", func(string, []string, Livelog) error {
		res <- "two"
		return nil
	}, []string{})
	r.AddVertex("three", func(string, []string, Livelog) error {
		res <- "three"
		return nil
	}, []string{})
	r.AddVertex("four", func(string, []string, Livelog) error {
		res <- "four"
		return nil
	}, []string{})
	r.AddVertex("five", func(string, []string, Livelog) error {
		res <- "five"
		return nil
	}, []string{})
	r.AddVertex("six", func(string, []string, Livelog) error {
		res <- "six"
		return nil
	}, []string{})
	r.AddVertex("seven", func(string, []string, Livelog) error {
		res <- "seven"
		return nil
	}, []string{})

	r.AddEdge("one", "two")
	r.AddEdge("one", "three")
	r.AddEdge("two", "four")
	r.AddEdge("two", "seven")
	r.AddEdge("five", "six")

	log := Livelog{
		Error: make(chan error),
		Line:  make(chan *Line),
	}
	err := make(chan error)

	go func() { err <- r.Run(log) }()

	select {
	case err := <-err:
		if want, have := error(nil), err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}

	results := make([]string, 7)
	timeout := time.After(100 * time.Millisecond)

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
