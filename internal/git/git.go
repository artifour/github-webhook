package git

import "os/exec"

func Pull(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull", "--no-edit")
	_, err := cmd.Output()

	return err
}
