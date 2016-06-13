package goss

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/urfave/cli"
	"github.com/fatih/color"
)

func Validate(c *cli.Context, startTime time.Time) {
	sys := system.New(c)

	// handle stdin
	var fh *os.File
	var err error
	var path string
	if !c.GlobalIsSet("gossfile") && hasStdin() {
		fh = os.Stdin
	} else {
		specFile := c.GlobalString("gossfile")
		path = filepath.Dir(specFile)
		fh, err = os.Open(specFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
	data, err := ioutil.ReadAll(fh)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	gossConfig := mergeJSONData(ReadJSONData(data), 0, path)

	out := make(chan []resource.TestResult)

	in := make(chan resource.Resource)

	go func() {
		for _, t := range gossConfig.Resources() {
			in <- t
		}
		close(in)
	}()

	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	gomaxprocs := runtime.GOMAXPROCS(-1)
	workerCount := gomaxprocs * 5
	if workerCount > 50 {
		workerCount = 50
	}
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range in {
				out <- f.Validate(sys)
			}

		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	//var outputer outputs.Outputer
	if c.Bool("no-color") {
		color.NoColor = true
	}

	outputer := outputs.GetOutputer(c.String("format"))

	exitCode := outputer.Output(out, startTime)
	os.Exit(exitCode)

}

func hasStdin() bool {
	if fi, err := os.Stdin.Stat(); err == nil {
		mode := fi.Mode()
		if (mode&os.ModeNamedPipe != 0) || mode.IsRegular() {
			return true
		}
	}
	return false
}
