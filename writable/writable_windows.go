package writable

import (
	"fmt"
	"os"
)

func IsWritable(path string, debug bool) (isWritable bool, err error) {
	isWritable = false
	info, err := os.Stat(path)
	if err != nil {
		if debug {
			fmt.Println("Path doesn't exist")
		}
		return
	}

	err = nil
	if !info.IsDir() {
		if debug {
			fmt.Println("Path isn't a directory")
		}
		return
	}

	// Check if the user bit is enabled in file permission
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		if debug {
			fmt.Println("Write permission bit is not set on this file for user")
		}
		return
	}

	isWritable = true
	return
}
