// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/cloner"
	"github.com/drone/go-task/task/logger"
)

// Downloader downloads a task artifact.
type Downloader interface {
	Download(context.Context, *task.Artifact) error
}

// New returns a downloader using the default
// system cache directory
func New(cloner cloner.Cloner) Downloader {
	return &downloader{cloner: cloner}
}

type downloader struct {
	cloner cloner.Cloner
}

func (d *downloader) Download(ctx context.Context, artifact *task.Artifact) error {
	if IsRepository(artifact.Source) {
		return d.clone(ctx, artifact)
	} else {
		return d.download(ctx, artifact)
	}
}

func (d *downloader) clone(ctx context.Context, artifact *task.Artifact) error {
	log := logger.FromContext(ctx)

	// get the target clone directory in the cache
	dir := ExpandCache(artifact.Destination)

	// exit if the artifact already exists
	if _, err := os.Stat(dir); err == nil {
		log.With("target", dir).
			Debug("cache hit")
		return nil
	} else {
		log.With("target", dir).
			Debug("cache miss")
	}

	// extract the clone url and ref
	url, ref := SplitRef(artifact.Source)

	log.With("source", url).
		With("revision", ref).
		With("target", dir).
		Debug("clone artifact")

	// clone the repository
	err := d.cloner.Clone(ctx, cloner.Params{
		Repo: url,
		Ref:  ref,
		Dir:  dir,
	})
	if err != nil {
		return err
	}

	scripts := artifact.Scripts
	if scripts == nil {
		return nil
	}
	after := scripts.After
	if after == nil {
		return nil
	}

	// execute post download commands
	path := ExpandCache(after.Path)
	args := ExpandCacheSlice(after.Args)
	dir = ExpandCache(after.Dir)

	// lookup the path if needed
	if p, err := exec.LookPath(path); err == nil {
		path = p
	}

	log.With("dir", dir).
		With("path", path).
		With("args", args).
		Debug("execute after clone script")

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	cmd.Env = after.Envs
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (d *downloader) download(_ context.Context, artifact *task.Artifact) error {
	// get the target download location
	dest := ExpandCache(artifact.Destination)

	// exit if the artifact already exists
	if _, err := os.Stat(dest); err == nil {
		return nil
	}

	// create the directory where the target is downloaded.
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	// download the artifact
	res, err := http.Get(artifact.Source)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if code := res.StatusCode; code > 299 {
		return fmt.Errorf("download error with status code %d", code)
	}

	// read the file into memory
	// TODO improve by writing directly to file
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// write the file to disk
	return os.WriteFile(dest, out, 0777)
}
