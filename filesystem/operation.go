package filesystem

import (
	"fmt"
	"time"
)

func CreateFile(fs *FileSystem, name string, data []byte) {
	if name == "" {
		fmt.Println("File name cannot be empty.")
		return
	}

	for _, entry := range fs.Directory {
		if entry.Name == name {
			fmt.Printf("File already exists with name %v\n", name)
			return
		}
	}

	// TODO: add support for data tobe added to multiple Inodes
	if len(data) > 4096 {
		fmt.Println("file size is too huge and not supported right now")
		return
	}

	inodeIndex := findFreeInode(fs)
	if inodeIndex == -1 {
		return
	}

	// Update the file system by updating Inode and Directory entries
	fs.Inodes[inodeIndex] = Inode{
		Data:       data,
		Used:       true,
		Permission: DEFAULT_FILE_PERMISSION,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	fs.Directory = append(fs.Directory, DirectoryEntry{
		Name:       name,
		Inode:      inodeIndex,
		Permission: DEFAULT_FILE_PERMISSION,
	})
}

func DeleteFile(fs *FileSystem, name string) {
	if name == "" {
		return
	}

	index := -1
	for i, entry := range fs.Directory {
		if entry.Name == name {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	inodeIndexToBeRemoved := fs.Directory[index].Inode

	fs.Directory = append(fs.Directory[:index], fs.Directory[index+1:]...)
	fs.Inodes = append(fs.Inodes[:inodeIndexToBeRemoved], fs.Inodes[inodeIndexToBeRemoved+1:]...)
	fs.FreeList[inodeIndexToBeRemoved] = true
}

func ListFiles(fs *FileSystem) {

	fmt.Println()
	fmt.Println(":::::::::::MetaData::::::::::")
	fmt.Printf("BlockSize: %v, TotalBlocks:%v, ", fs.SuperBlock.BlockSize, fs.SuperBlock.TotalBlocks)
	fmt.Println()
	fmt.Println("Current Directory: ", fs.CurrentDir)
	fmt.Println()
	fmt.Println(":::::::::::ENTRIES::::::::::")
	for i, entry := range fs.Directory {
		fmt.Printf("%v) %v", i+1, entry)
		fmt.Println()
	}
}
