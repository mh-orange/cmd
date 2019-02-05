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
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestProcessAppendArgs(t *testing.T) {
	proc := &process{}
	if len(proc.args) > 0 {
		t.Errorf("expected no arguments")
	} else {
		proc.AppendArgs("-o")
		if len(proc.args) == 1 {
			if proc.args[0] != "-o" {
				t.Errorf("unexpected argument %q", proc.args[0])
			}
		} else {
			t.Errorf("expected len(proc.cmd.Args) to be 1")
		}
	}
}

func TestCommandStart(t *testing.T) {
	cmd := New("")
	wantStderr := "Re-elect Mayor Red Thomas. Progress is his middle name"
	wantStdout := "I'll be the most powerful man in Hill Valley and I'm gonna clean up this town!"

	stderr := bytes.NewBuffer(nil)
	stdout := bytes.NewBuffer(nil)

	cmd.SetPath(os.Args[0])
	proc := cmd.Process()
	proc.(*process).cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("STDERR=%s", wantStderr), fmt.Sprintf("STDOUT=%s", wantStdout)}
	proc.AppendArgs("-test.run=TestHelperProcess", "--")
	proc.Stderr(stderr)
	proc.Stdout(stdout)

	err := proc.Start()
	if err == nil {
		err = proc.Wait()
	}

	if err == nil {
		gotStderr := string(stderr.Bytes())
		if gotStderr != wantStderr {
			t.Errorf("want %q got %q", wantStderr, gotStderr)
		}

		gotStdout := string(stdout.Bytes())
		if gotStdout != wantStdout {
			t.Errorf("want %q got %q", wantStdout, gotStdout)
		}
	} else {
		t.Errorf("Unexpected error: %v", err)
	}

}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "%s", os.Getenv("STDOUT"))
	fmt.Fprintf(os.Stderr, "%s", os.Getenv("STDERR"))
	os.Exit(0)
}

func TestTestCommand(t *testing.T) {
	wantStdout := "Humback... people?"
	wantStderr := "Whales, Mr. Scott, Whales"
	cmd := &TestCmd{
		Stdout: []byte(wantStdout),
		Stderr: []byte(wantStderr),
	}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	proc := cmd.Process()
	proc.Stdout(stdout)
	proc.Stderr(stderr)
	err := proc.Start()
	if err == nil {
		err = proc.Wait()
		gotStdout := string(stdout.Bytes())
		if wantStdout != gotStdout {
			t.Errorf("STDOUT: want %q got %q", wantStdout, gotStdout)
		}

		gotStderr := string(stderr.Bytes())
		if wantStderr != gotStderr {
			t.Errorf("STDERR: want %q got %q", wantStderr, gotStderr)
		}
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	cmd.StartErr = io.EOF
	gotErr := cmd.Process().Start()
	if io.EOF != gotErr {
		t.Errorf("want %v got %v", io.EOF, gotErr)
	}
	cmd.StartErr = nil

	cmd.WaitErr = io.EOF
	proc = cmd.Process()
	gotErr = proc.Start()
	if gotErr == nil {
		gotErr = proc.Wait()
		if gotErr != io.EOF {
			t.Errorf("want %v got %v", io.EOF, gotErr)
		}
	} else {
		t.Errorf("Unexpected error: %v", gotErr)
	}
	cmd.WaitErr = nil
	cmd.KillErr = io.EOF
	gotErr = cmd.Process().Kill()
	if io.EOF != gotErr {
		t.Errorf("want %v got %v", io.EOF, gotErr)
	}

}
