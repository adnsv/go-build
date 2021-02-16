package mingw

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func BashCommand(root string, msystem string, env []string, commands []string) *exec.Cmd {
	bash := filepath.Join(root, "usr", "bin", "bash.exe")
	ret := exec.Command(bash, "-l", "-c", strings.Join(commands, " && "))
	ret.Env = append(env, "MSYSTEM="+msystem)
	return ret
}
