package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteLog(log []byte) {
	path := filepath.Join(os.TempDir(), "server.log")
	if err := os.WriteFile(path, []byte(log), 0644); err != nil {
		fmt.Println("File Write error")
		return
	}
	fmt.Println("Log written successfully")
}
