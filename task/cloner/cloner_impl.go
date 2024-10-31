// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloner

import (
	"context"
	"fmt"
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

	// Build the git clone command with verbosity
	if params.Ref != "" {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth=1", "--branch="+params.Ref, "-v", params.Repo, params.Dir)
	} else {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth=1", "-v", params.Repo, params.Dir)
	}

	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr

	// Run the clone command and capture output in case of error
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repo: %v, output: %s", err, output)
	}

	// Check out the specific SHA if provided, also with verbose output
	if params.Sha != "" {
		cmd = exec.CommandContext(ctx, "git", "-C", params.Dir, "checkout", params.Sha, "-v")
		cmd.Stdout = c.stdout
		cmd.Stderr = c.stderr

		// Run the checkout command and capture output in case of error
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout SHA: %v, output: %s", err, output)
		}
	}

	return nil
}
