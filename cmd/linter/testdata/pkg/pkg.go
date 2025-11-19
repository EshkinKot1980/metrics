package pkg

import (
	"log"
	"os"
)

func FatalCheck(fatal, exit bool) {
	if fatal {
		log.Fatal("fatal in pkg.FatalCheck") // want "unacceptable log.Fatal call"
	}

	if exit {
		os.Exit(1) // want  "unacceptable os.Exit call"
	}
}
