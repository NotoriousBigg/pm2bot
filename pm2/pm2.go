package pm2

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

type StartOptions struct {
	Script      string
	Name        string
	Interpreter string
	Args        []string
}

const DefaultTimeout = 15 * time.Second

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pm2", args...)

	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		if errBuf.Len() == 0 {
			return "", err
		}
		return strings.TrimSpace(errBuf.String()), err
	}

	return strings.TrimSpace(out.String()), nil
}

func ListJSON() (string, error) {
	return run("jlist")
}

func Describe(name string) (string, error) {
	return run("describe", name)
}

func Delete(name string) (string, error) {
	return run("delete", name)
}

func StartWithOptions(opt StartOptions) (string, error) {
	args := []string{"start", opt.Script}

	if opt.Name != "" {
		args = append(args, "--name", opt.Name)
	}
	if opt.Interpreter != "" {
		args = append(args, "--interpreter", opt.Interpreter)
	}
	if len(opt.Args) > 0 {
		args = append(args, "--")
		args = append(args, opt.Args...)
	}

	return run(args...)
}

func List() (string, error)               { return run("list") }
func Restart(name string) (string, error) { return run("restart", name) }
func Stop(name string) (string, error)    { return run("stop", name) }
func Save() (string, error)               { return run("save") }
