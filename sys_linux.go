//go:build !windows

package main

import "os/exec"

func setWindowAttributes(cmd *exec.Cmd) {}

func setKillGroupAttributes(cmd *exec.Cmd) {}
