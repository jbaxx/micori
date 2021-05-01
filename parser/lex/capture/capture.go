package capture

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const DefaultEditor = "vim"

type PreferredEditorResolver func() string

func GetPreferredEditorFromEnvironment() string {
	editor := os.Getenv("EDITOR")

	if editor == "" {
		return DefaultEditor
	}

	return editor
}

func resolvedEditorArguments(executable string, filename string) []string {
	args := []string{filename}

	if strings.Contains(executable, "Visual Studio Code.app") {
		args = append([]string{"--wait"}, args...)
	}

	return args
}

func OpenFileInEditor(filename string, resolveEditor PreferredEditorResolver) error {
	executable, err := exec.LookPath(resolveEditor())
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, resolvedEditorArguments(executable, filename)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func CaptureInputFromEditor(content []byte, resolveEditor PreferredEditorResolver) ([]byte, error) {
	file, err := ioutil.TempFile(os.TempDir(), "*.json")
	if err != nil {
		return []byte{}, err
	}

	filename := file.Name()
	defer os.Remove(filename)

	_, err = file.Write(content)
	if err != nil {
		return []byte{}, err
	}

	if err = file.Close(); err != nil {
		return []byte{}, err
	}

	if err = OpenFileInEditor(filename, resolveEditor); err != nil {
		return []byte{}, err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil

}
