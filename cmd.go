// Copyright 2019 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// Process is an instance of a command.  A process is not initially
// running and must be started by means of the Start function
type Process interface {
	// AppendArgs will add arguments to the end of the argument list.  This allows
	// adding instance specific arguments to the process
	AppendArgs(args ...string)

	// Start will attempt to start the process.  Start always returns immediately and
	// does not wait for the process to complete
	Start() error

	// Kill attempts to kill the underlying OS process.  This may or may not be implemented
	// on all operating systems
	Kill() error

	// Wait will wait for the underlying process to complete.  Wait will not return until
	// the OS process has either finished on its own or has been killed
	Wait() error

	// Stdin sets the process's standard input to the given reader
	Stdin(io.Reader)

	// Stdout adds the Writer to Stdout.  All Writers added to Stdout will receive all
	// data that the process writes to Stdout.  This is implemented with an underlying
	// multi-writer
	Stdout(io.Writer)

	// Stderr adds the Writer to Stderr.  All Writers added to Stderr will receive all
	// data that the process writes to Stderr.  This is implemented with an underlying
	// multi-writer
	Stderr(io.Writer)
}

type process struct {
	cmd    *exec.Cmd
	stderr multiWriter
	stdout multiWriter
	args   []string
}

func (proc *process) AppendArgs(args ...string) {
	proc.args = append(proc.args, args...)
}

func (proc *process) String() string {
	args := []string{}
	for _, arg := range append(proc.cmd.Args, proc.args...) {
		if strings.IndexAny(arg, " \t\n\r") >= 0 {
			arg = fmt.Sprintf("%q", arg)
		}
		args = append(args, arg)
	}
	return strings.Join(args, " ")
}

func (proc *process) Start() error {
	proc.cmd.Args = append(proc.cmd.Args, proc.args...)
	stderr, err := proc.cmd.StderrPipe()
	if err == nil {
		go proc.stderr.copy(stderr)

		var stdout io.Reader
		stdout, err = proc.cmd.StdoutPipe()
		if err == nil {
			go proc.stdout.copy(stdout)
		}
	}

	if err == nil {
		err = proc.cmd.Start()
	}
	return err
}

func (proc *process) Kill() error {
	return proc.cmd.Process.Kill()
}

func (proc *process) Wait() error {
	return proc.cmd.Wait()
}

func (proc *process) Stdin(reader io.Reader) {
	proc.cmd.Stdin = reader
}

func (proc *process) Stderr(writer io.Writer) {
	proc.stderr.add(writer)
}

func (proc *process) Stdout(writer io.Writer) {
	proc.stdout.add(writer)
}

// Command represents a command to be run in the future.  A running command
// is represented by the Process object
type Command interface {
	// Path returns the path to the command.  This is equivalent to os.Args[0]
	Path() string

	// SetPath allows setting the path to the command
	SetPath(path string)

	// Process creates a new process with the command and its arguments.  The
	// returned Process will not yet be started
	Process() Process
}

type cmd struct {
	path string
	args []string
}

// New will create a new command for the given path and argument list.
// Arguments passed in to New will be assigned to every process that
// is created for this command.  Each Process can be customized by calling
// the AppendArguments function
func New(path string, args ...string) Command {
	return &cmd{path, args}
}

func (cmd *cmd) Path() string        { return cmd.path }
func (cmd *cmd) SetPath(path string) { cmd.path = path }

func (cmd *cmd) Process() Process {
	return &process{
		cmd: exec.Command(cmd.path, cmd.args...),
	}
}

type testProcess struct {
	stdin       []byte
	stdinReader io.Reader

	stdout       []byte
	stdoutWriter multiWriter

	stderr       []byte
	stderrWriter multiWriter

	wg sync.WaitGroup

	startErr error
	waitErr  error
	killErr  error
}

func (tp *testProcess) AppendArgs(args ...string) {}

func (tp *testProcess) Start() error {
	if tp.startErr == nil {
		tp.wg.Add(2)
		go func() {
			if len(tp.stdout) > 0 {
				tp.stdoutWriter.Write(tp.stdout)
			}
			tp.stdoutWriter.Close()
			tp.wg.Done()
		}()

		go func() {
			if len(tp.stderr) > 0 {
				tp.stderrWriter.Write(tp.stderr)
			}
			tp.stderrWriter.Close()
			tp.wg.Done()
		}()
	}
	return tp.startErr
}

func (tp *testProcess) Kill() error { return tp.killErr }

func (tp *testProcess) Wait() error {
	tp.wg.Wait()
	return tp.waitErr
}

func (tp *testProcess) Stdin(reader io.Reader)  { tp.stdinReader = reader }
func (tp *testProcess) Stdout(writer io.Writer) { tp.stdoutWriter.add(writer) }
func (tp *testProcess) Stderr(writer io.Writer) { tp.stderrWriter.add(writer) }

// TestCmd is useful for mocking commands without actually executing
// anything
type TestCmd struct {
	// Stdout is the string written to all Stdout receivers
	Stdout []byte

	// Stderr is the string written to all Stderr receivers
	Stderr []byte

	// StartErr is returned by the Process' Start function
	StartErr error

	// WaitErr is returned by the Process' Wait function
	WaitErr error

	// KillErr is returned by the Process' Kill function
	KillErr error
}

// Process creates a test process that will behave according to
// the Stdout/Sderr and StartErr/WaitErr/KillErr attributes of the
// TestCmd
func (tc *TestCmd) Process() Process {
	return &testProcess{
		stdout:   tc.Stdout,
		stderr:   tc.Stderr,
		startErr: tc.StartErr,
		waitErr:  tc.WaitErr,
		killErr:  tc.KillErr,
	}
}

// Path returns an empty string, it is not relevant for TestCmds
func (*TestCmd) Path() string { return "" }

// SetPath does nothing for TestCmds
func (*TestCmd) SetPath(string) {}
