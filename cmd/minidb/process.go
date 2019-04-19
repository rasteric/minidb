package main

import (
	"errors"

	ps "github.com/mitchellh/go-ps"
)

// FindProcessByName returns a running process PID by the name of the process,
// or 0 and an error if the process could not be found. A PID of 0 is always invalid.
func FindProcessByName(name string) (int, error) {
	processes, err := ps.Processes()
	if err != nil {
		return 0, err
	}
	for _, p := range processes {
		if p.Executable() == name {
			return p.Pid(), nil
		}
	}
	return 0, errors.New("process '%s' not found")
}
