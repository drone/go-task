// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloner

import (
	"context"
	"io"
	"os/exec"
)

type cloner struct {
	stdout io.Writer
	stderr io.Writer
}

// Default returns the default cloner which relies on the
// os/exec package to clone a repository using the git
// binary installed on the host.
func Default() Cloner {
	return new(cloner)
}

func (c *cloner) Clone(ctx context.Context, params Params) error {
	var cmd *exec.Cmd

	if params.Ref != "" {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth=1", "--branch="+params.Ref, params.Repo, params.Dir)
	} else {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth=1", params.Repo, params.Dir)
	}

	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// check out the specific SHA if provided
	if params.Sha != "" {
		cmd = exec.CommandContext(ctx, "git", "-C", params.Dir, "checkout", params.Sha)
		cmd.Stdout = c.stdout
		cmd.Stderr = c.stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
