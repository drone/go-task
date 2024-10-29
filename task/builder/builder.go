package builder

import (
	"context"
	"fmt"
	globallogger "github.com/harness/runner/logger/logger"
	"log/slog"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Builder struct {
	TaskYmlPath string // path to the task.yml file which contains instructions for execution
}

func New(taskYmlPath string) *Builder {
	return &Builder{
		TaskYmlPath: taskYmlPath,
	}
}

// Build parses the task.yml file and generates the executable binary for a task
func (b *Builder) Build(ctx context.Context) (string, error) {
	log := globallogger.FromContext(ctx)
	out, err := ParseFile(b.TaskYmlPath)
	if err != nil {
		return "", err
	}

	// install any dependencies for the task
	log.Info("installing dependencies")
	if err := b.installDeps(ctx, out.Spec.Deps); err != nil {
		return "", fmt.Errorf("failed to install dependencies: %w", err)
	}
	log.Info("finished installing dependencies")

	var binpath string

	// build go binary if specified
	if out.Spec.Run.Go != nil {
		module := out.Spec.Run.Go.Module
		if module != "" {
			binName := "task.exe"
			err = b.buildGoModule(ctx, module, binName)
			if err != nil {
				return "", fmt.Errorf("failed to build go module: %w", err)
			}
			binpath = filepath.Join(filepath.Dir(b.TaskYmlPath), binName)
		}
	} else if out.Spec.Run.Bash != nil {
		binpath = filepath.Join(filepath.Dir(b.TaskYmlPath), out.Spec.Run.Bash.Script)
	} else {
		return "", fmt.Errorf("no execution specified in task.yml file")
	}
	return binpath, nil
}

// installDeps installs any dependencies for the task
func (b *Builder) installDeps(ctx context.Context, deps Deps) error {
	goos := runtime.GOOS

	// install linux dependencies
	if goos == "linux" {
		return b.installAptDeps(ctx, deps.Apt)
	}

	// install darwin dependencies
	if goos == "darwin" {
		return b.installBrewDeps(ctx, deps.Brew)
	}

	return nil
}

func (b *Builder) installAptDeps(ctx context.Context, deps []AptDep) error {
	log := globallogger.FromContext(ctx)
	var err error
	if len(deps) > 0 {
		log.Info("apt-get update")

		cmd := b.cmdRunner("sudo", "apt-get", "update")
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	for _, dep := range deps {
		log.Info("apt-get install", slog.String("package", dep.Name))

		cmd := b.cmdRunner("sudo", "apt-get", "install", dep.Name)
		if err = cmd.Run(); err != nil {
			// TODO: perhaps errors can be logged as warnings instead of returning here,
			// but we can evaluate this in the future.
			return err
		}
	}

	return nil
}

// cmdRunner returns a new exec.Cmd with the given name and arguments
// It populates the working directory as the directory of the task.yml file.
func (b *Builder) cmdRunner(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Dir = filepath.Dir(b.TaskYmlPath)
	return cmd
}

func (b *Builder) buildGoModule(
	ctx context.Context,
	module string,
	binName string, // name of the target binary
) error {
	log := globallogger.FromContext(ctx)
	log.Info("go build", slog.String("module", module))

	// build the code
	cmd := b.cmdRunner("go", "build", "-o", binName, module)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (b *Builder) installBrewDeps(ctx context.Context, deps []BrewDep) error {
	log := globallogger.FromContext(ctx)
	for _, item := range deps {
		log.Info("brew install", slog.String("package", item.Name))

		cmd := b.cmdRunner("brew", "install", item.Name)
		if err := cmd.Run(); err != nil {
			// TODO: perhaps errors can be logged as a warning instead of returning here,
			// but we can evaluate this in the future.
			return err
		}
	}
	return nil
}
