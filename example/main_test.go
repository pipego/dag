package main

import (
	"testing"

	"go.uber.org/goleak"
)

func TestRunMain(t *testing.T) {
	defer goleak.VerifyNone(t)
	main()
}
