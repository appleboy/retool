package main

import (
	"os"
	"os/exec"
	"path"
	"strings"
)

type tool struct {
	Repository string // eg "github.com/tools/godep"
	Commit     string // eg "3020345802e4bff23902cfc1d19e90a79fae714e"
}

func (t tool) path() string {
	return path.Join(tooldir, "src", t.Repository)
}

func (t tool) executable() string {
	return path.Base(t.Repository)
}

func setEnvVar(cmd *exec.Cmd, key, val string) {
	var env []string
	if cmd.Env != nil {
		env = cmd.Env
	} else {
		env = os.Environ()
	}

	envSet := false
	for i, envVar := range env {
		if strings.HasPrefix(envVar, key+"=") {
			env[i] = key + "=" + val
			envSet = true
		}
	}
	if !envSet {
		env = append(cmd.Env, key+"="+val)
	}

	cmd.Env = env
}

func setGopath(cmd *exec.Cmd) {
	setEnvVar(cmd, "GOPATH", tooldir)
}

func get(t tool) error {
	log("downloading " + t.Repository)
	cmd := exec.Command("go", "get", "-d", t.Repository)
	setGopath(cmd)
	_, err := cmd.Output()
	return err
}

func setVersion(t tool) error {
	log("setting version for " + t.Repository)
	cmd := exec.Command("git", "fetch")
	cmd.Dir = t.path()
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("git", "checkout", t.Commit)
	cmd.Dir = t.path()
	_, err = cmd.Output()
	return err
}

func installBin(t tool) error {
	log("installing " + t.Repository)
	cmd := exec.Command("go", "install", t.Repository)
	setGopath(cmd)
	_, err := cmd.Output()
	return err
}

func cleanGit(t tool) error {
	log("cleaning " + t.Repository)
	cmd := exec.Command("rm", "-r", "-f", ".git")
	cmd.Dir = t.path()
	_, err := cmd.Output()
	return err
}

func install(t tool) error {
	err := get(t)
	if err != nil {
		fatalExec("go get -d "+t.Repository, err)
	}

	err = setVersion(t)
	if err != nil {
		fatalExec("git checkout "+t.Commit, err)
	}

	err = installBin(t)
	if err != nil {
		fatalExec("go install "+t.Repository, err)
	}

	err = cleanGit(t)
	if err != nil {
		fatalExec("rm -rf .git of "+t.Repository, err)
	}

	return nil
}