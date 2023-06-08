package main

import (
	fspkg "fs/filesystem"
	"os"
)

const (
	FileSystemBinary = "filesystem.bin"
)

func main() {
	_, err := os.Stat(FileSystemBinary)
	if os.IsNotExist(err) {
		if err := fspkg.CreateFileSystem(); err != nil {
			panic(err)
		}
	}

	fs := fspkg.LoadFileSystem()
	defer fspkg.SaveFileSystem(fs)

}

/**
fspkg.CreateFile(&fs, "test.txt", []byte("Hello there"))
fspkg.CreateFile(&fs, "test1.txt", []byte("Hello there"))
fspkg.CreateFile(&fs, "test2.txt", []byte("Hello there"))
fspkg.ListFiles(&fs)
fspkg.DeleteFile(&fs, "test2.txt")
fspkg.ListFiles(&fs)
*/
