// Package editor provides a utility to read from $EDITOR.
package editor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

// ReadEditor opens the editor and returns the value.
func Edit(editor string, content string) ([]byte, error) {
	f, err := ioutil.TempFile("", "spass")
	if err != nil {
		return nil, fmt.Errorf("creating tmpfile: %s", err)
	}
	defer os.Remove(f.Name())

	err = f.Truncate(0)
	if err != nil {
		return nil, fmt.Errorf("clearing tmpfile: %s", err)
	}

	_, err = f.Write([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("writing to tmpfile: %s", err)
	}

	cmd := exec.Command("sh", "-c", editor+" "+f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("opening $EDITOR: %s", err)
	}

	b, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("reading tmpfile: %s", err)
	}

	return b, nil
}
