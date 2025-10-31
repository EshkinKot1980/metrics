package main

import (
	"fmt"
	"log"
	"os"
)

type config struct {
	panic bool
	falat bool
	exit  bool
}

func main() {
	cfg := &config{}
	if cfg.panic {
		panic("pacin in main") // want "unacceptable panic call"
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cgf *config) error {
	if cgf.falat {
		log.Fatal("fatal in main.run") // want "unacceptable log.Fatal call"
	}

	if cgf.exit {
		os.Exit(1) // want  "unacceptable os.Exit call"
	}

	return fmt.Errorf("run error")
}
