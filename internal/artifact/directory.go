// Package artifact creates new output directories while protecting source trees.
package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveNew(name string) (string, error) {
	absolute, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(absolute); err == nil {
		return "", fmt.Errorf("OUTPUT_EXISTS: %s", absolute)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	ancestor := filepath.Dir(absolute)
	suffix := []string{filepath.Base(absolute)}
	for {
		resolved, err := filepath.EvalSymlinks(ancestor)
		if err == nil {
			return filepath.Join(append([]string{resolved}, suffix...)...), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", err
		}
		suffix = append([]string{filepath.Base(ancestor)}, suffix...)
		ancestor = parent
	}
}

func CreateDirectory(name string, inputs []string) (string, error) {
	output, err := resolveNew(name)
	if err != nil {
		return "", err
	}
	if err := checkInputs(output, inputs); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0700); err != nil {
		return "", err
	}
	if err := os.Mkdir(output, 0700); err != nil {
		return "", err
	}
	return output, nil
}

func checkInputs(output string, inputs []string) error {
	for _, input := range inputs {
		resolved, err := filepath.EvalSymlinks(input)
		if err != nil {
			return err
		}
		resolved, err = filepath.Abs(resolved)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(filepath.Dir(resolved), output)
		if err != nil {
			return err
		}
		if relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative) {
			return fmt.Errorf("OUTPUT_INSIDE_INPUT_DIRECTORY: %s", output)
		}
	}
	return nil
}

func WriteFile(name string, data []byte, inputs []string) error {
	output, err := resolveNew(name)
	if err != nil {
		return err
	}
	if err := checkInputs(output, inputs); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
