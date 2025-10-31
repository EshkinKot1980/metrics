package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestFatalCheckAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), FatalCheckAnalyzer, "./...")
}
