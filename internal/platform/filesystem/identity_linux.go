//go:build linux

package filesystem

import (
	"os"
	"syscall"

	"sdd.dev/sdd/internal/workspace"
)

func nativeIdentity(info os.FileInfo) (workspace.NativeFileIdentity, *uint64) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return workspace.NativeFileIdentity{}, nil
	}
	links := stat.Nlink
	return workspace.NativeFileIdentity{Device: stat.Dev, File: stat.Ino}, &links
}
