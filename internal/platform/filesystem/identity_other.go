//go:build !linux

package filesystem

import (
	"os"

	"sdd.dev/sdd/internal/workspace"
)

// Native certification must replace this unknown identity with a proven
// platform implementation before mutation support is claimed.
func nativeIdentity(os.FileInfo) (workspace.NativeFileIdentity, *uint64) {
	return workspace.NativeFileIdentity{}, nil
}
