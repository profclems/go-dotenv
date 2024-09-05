//go:build !windows

package dotenv

import (
	"os"

	"github.com/google/renameio"
)

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	return renameio.WriteFile(filename, data, perm)
}
