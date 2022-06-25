package runner

import (
	"errors"
	"testing"
	"time"
)

func TestZero(t *testing.T) {
	var r Runner

	res := make(chan error)
	go func() { res <- r.Run() }()

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

	err := errors.New("error")
	r.AddVertex("one", func([]string) error { return err }, []string{})

	res := make(chan error)
	go func() { res <- r.Run() }()

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

	err := errors.New("error")
	r.AddVertex("one", func([]string) error { return err }, []string{})
	r.AddVertex("two", func([]string) error { return nil }, []string{})
	r.AddVertex("three", func([]string) error { return nil }, []string{})
	r.AddVertex("fout", func([]string) error { return nil }, []string{})

	res := make(chan error)
	go func() { res <- r.Run() }()

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

	r.AddVertex("one", func([]string) error { return nil }, []string{})
	r.AddVertex("two", func([]string) error { return nil }, []string{})
	r.AddVertex("three", func([]string) error { return nil }, []string{})
	r.AddVertex("four", func([]string) error { return nil }, []string{})

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("three", "one")

	res := make(chan error)
	go func() { res <- r.Run() }()

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

	r.AddVertex("one", func([]string) error { return nil }, []string{})
	r.AddVertex("two", func([]string) error { return nil }, []string{})
	r.AddVertex("three", func([]string) error { return nil }, []string{})
	r.AddVertex("four", func([]string) error { return nil }, []string{})

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("three", "definitely-not-a-valid-vertex")

	res := make(chan error)
	go func() { res <- r.Run() }()

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

	r.AddVertex("one", func([]string) error { return nil }, []string{})
	r.AddVertex("two", func([]string) error { return nil }, []string{})
	r.AddVertex("three", func([]string) error { return nil }, []string{})
	r.AddVertex("four", func([]string) error { return nil }, []string{})

	r.AddEdge("one", "two")
	r.AddEdge("two", "three")
	r.AddEdge("three", "four")
	r.AddEdge("definitely-not-a-valid-vertex", "three")

	res := make(chan error)
	go func() { res <- r.Run() }()

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

	res := make(chan string, 7)
	r.AddVertex("one", func([]string) error {
		res <- "one"
		return nil
	}, []string{})
	r.AddVertex("two", func([]string) error {
		res <- "two"
		return nil
	}, []string{})
	r.AddVertex("three", func([]string) error {
		res <- "three"
		return nil
	}, []string{})
	r.AddVertex("four", func([]string) error {
		res <- "four"
		return nil
	}, []string{})
	r.AddVertex("five", func([]string) error {
		res <- "five"
		return nil
	}, []string{})
	r.AddVertex("six", func([]string) error {
		res <- "six"
		return nil
	}, []string{})
	r.AddVertex("seven", func([]string) error {
		res <- "seven"
		return nil
	}, []string{})

	r.AddEdge("one", "two")
	r.AddEdge("one", "three")
	r.AddEdge("two", "four")
	r.AddEdge("two", "seven")
	r.AddEdge("five", "six")

	err := make(chan error)
	go func() { err <- r.Run() }()

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
