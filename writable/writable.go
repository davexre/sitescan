package writable

import (
	"fmt"
	"os"
	"syscall"
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

	var stat syscall.Stat_t
	if err = syscall.Stat(path, &stat); err != nil {
		if debug {
			fmt.Println("Unable to get stat")
		}
		return
	}

	err = nil
	if uint32(os.Geteuid()) != stat.Uid {
		isWritable = false
		if debug {
			fmt.Println("User doesn't have permission to write to this directory")
		}
		return
	}

	isWritable = true
	return
}
