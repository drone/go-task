// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/cloner"
	download "github.com/drone/go-task/task/downloader"
	"github.com/drone/go-task/task/drivers/cgi"
)

var (
	// path to the task file
	path = flag.String("path", "", "")

	// pretty print the task output
	pretty = flag.Bool("pretty", false, "")

	// runs with verbose output if true
	verbose = flag.Bool("verbose", false, "")

	// displays the help / usage if true
	help = flag.Bool("help", false, "")
)

func main() {

	// parse the input parameters
	flag.BoolVar(help, "h", false, "")
	flag.BoolVar(verbose, "v", false, "")
	flag.Usage = usage
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// set the default log level
	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}

	// set the default logger (used by handlers)
	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: level},
			),
		),
	)

	// user may specify the path as a non-flag variable
	if flag.NArg() > 0 {
		path = &flag.Args()[0]
	}

	// parse the task file
	data, err := os.ReadFile(*path)
	if err != nil {
		log.Fatalln(err)
	}

	// unmarshal the task file into a request
	req := new(task.Request)
	if err := json.Unmarshal(data, req); err != nil {
		log.Fatalln(err)
	}

	cache, err := os.UserCacheDir()
	if err != nil {
		log.Fatalln(err)
	}

	// create the task downloader which downloads and
	// caches tasks at ~/.cache/harness/task
	downloader := download.New(
		// use the built-in cloner which uses
		// os/exec to clone the repository.
		//
		// this avoids any external dependencies
		// in this module, however, a production
		// installation may want to provide a
		// custom implementation that uses a native
		// go git module to avoid os/exec.
		cloner.Default(),

		// top-level directory where the downloading should happen
		cache,
	)

	// create the task router
	router := task.NewRouter()
	router.RegisterFunc("sample/exec", execHandler) // sample bult-in handler
	router.RegisterFunc("sample/file", fileHandler) // sample bult-in handler
	router.NotFound(
		// default to cgi handler when no built-in
		// task handler is found.
		cgi.New(
			// use the default downloader which
			// caches tasks at ~/.cache/harness/task
			downloader,
		),
	)

	// handle the request
	res := router.Handle(context.Background(), req)
	if err := res.Error(); err != nil {
		log.Fatalln(err)
	}

	// if the response is an error, print the error
	// message and exit with failure.
	if err := res.Error(); err != nil {
		log.Fatalln(err)
	}

	if *pretty {
		// write the task details to stdout
		fmt.Fprintf(os.Stdout, "id:   %s", req.Task.ID)
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintf(os.Stdout, "type: %s", req.Task.Type)
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "")

		// decode the response body into a temporary
		// data structure.
		var temp any
		json.Unmarshal(res.Body(), &temp)

		// re-encode the response body as json with
		// indentation.
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(temp)
		return
	}

	// write the logs to stdout
	os.Stdout.Write(res.Body())
}

var usage = func() {
	println(`Usage: go-task [OPTION]... [PATH]

      --path           path to the task file
      --pretty         pretty print the task output
  -v, --verbose        execute the task with verbose output
  -h, --help           display this help and exit

Examples:
  go-task path/to/task.json
`)
}
