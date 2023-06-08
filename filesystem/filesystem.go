package filesystem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const (
	BlockSize        = 4096
	FileSystemBinary = "../filesystem.bin"

	DEFAULT_FILE_PERMISSION = "rw-r--r--"
)

type SuperBlock struct { // meta-data of the filesystem
	TotalBlocks int
	BlockSize   int
}

type DirectoryEntry struct {
	Name       string // name of the directory i.e '/'
	Inode      int    // Inode index where directory begins on file system
	Permission string
}

type Inode struct {
	Data       []byte
	Used       bool
	Permission string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type FileSystem struct {
	SuperBlock SuperBlock
	FreeList   []bool
	Directory  []DirectoryEntry
	Inodes     []Inode
	CurrentDir string
}

func CreateFileSystem() error {
	fs := FileSystem{
		SuperBlock: SuperBlock{
			TotalBlocks: 1000,
			BlockSize:   BlockSize,
		},
		FreeList:   make([]bool, 1000),
		Directory:  make([]DirectoryEntry, 0),
		Inodes:     make([]Inode, 0),
		CurrentDir: "/",
	}

	// Initialize the free list
	for i := 0; i < fs.SuperBlock.TotalBlocks; i++ {
		fs.FreeList[i] = true
	}

	// Create root directory
	rootInode := fs.allocateInode(make([]byte, BlockSize), "rwxr-xr-x")
	fs.Directory = append(fs.Directory, DirectoryEntry{
		Name:       "/",
		Inode:      rootInode,
		Permission: "rwxr-xr-x",
	})

	if err := SaveFileSystem(fs); err != nil {
		return err
	}

	return nil
}

func LoadFileSystem() FileSystem {
	fs := FileSystem{}

	file, err := os.Open(FileSystemBinary)
	if err != nil {
		fmt.Println("Error opening file system:", err)
		return fs
	}
	defer file.Close()

	sb := make([]byte, 8)
	file.Read(sb)
	fs.SuperBlock.TotalBlocks = int(binary.LittleEndian.Uint32(sb[:4]))
	fs.SuperBlock.BlockSize = int(binary.LittleEndian.Uint32(sb[4:]))

	fl := make([]byte, fs.SuperBlock.TotalBlocks)
	file.Read(fl)
	fs.FreeList = make([]bool, fs.SuperBlock.TotalBlocks)
	for i := 0; i < fs.SuperBlock.TotalBlocks; i++ {
		if fl[i] != 0 {
			fs.FreeList[i] = true
		}
	}

	dirSizeBytes := make([]byte, 4)
	file.Read(dirSizeBytes)
	dirSize := binary.LittleEndian.Uint32(dirSizeBytes)
	fs.Directory = make([]DirectoryEntry, dirSize)
	for i := uint32(0); i < dirSize; i++ {
		entry := DirectoryEntry{}

		nameLengthBytes := make([]byte, 4)
		file.Read(nameLengthBytes)
		nameLength := binary.LittleEndian.Uint32(nameLengthBytes)

		nameBytes := make([]byte, nameLength)
		file.Read(nameBytes)
		entry.Name = string(nameBytes)

		inodesBytes := make([]byte, 4)
		file.Read(inodesBytes)
		entry.Inode = int(binary.LittleEndian.Uint32(inodesBytes))

		permissionBytes := make([]byte, 9)
		file.Read(permissionBytes)
		entry.Permission = string(permissionBytes)

		fs.Directory[i] = entry
	}

	inodeLenBytes := make([]byte, 4)
	file.Read(inodeLenBytes)
	inodeLen := binary.LittleEndian.Uint32(inodeLenBytes)
	for i := uint32(0); i < inodeLen; i++ {
		inode := Inode{}

		dataLengthBytes := make([]byte, 4)
		file.Read(dataLengthBytes)
		dataLength := binary.LittleEndian.Uint32(dataLengthBytes)

		dataBytes := make([]byte, dataLength)
		file.Read(dataBytes)
		inode.Data = dataBytes // Data is filled

		usedBytes := make([]byte, 1)
		file.Read(usedBytes)
		if usedBytes[0] != 0 { // if usedBytes[0] is not zero then used flag was set true
			inode.Used = true
		}

		permissionBytes := make([]byte, 9)
		file.Read(permissionBytes)
		inode.Permission = string(permissionBytes) // Permission is filled

		createdAtBytes := make([]byte, 8)
		file.Read(createdAtBytes)
		inode.CreatedAt = time.Unix(int64(binary.LittleEndian.Uint64(createdAtBytes)), 0)

		updatedAtBytes := make([]byte, 8)
		file.Read(updatedAtBytes)
		inode.UpdatedAt = time.Unix(int64(binary.LittleEndian.Uint64(updatedAtBytes)), 0)

		fs.Inodes = append(fs.Inodes, inode)
	}

	currDirLenBytes := make([]byte, 4)
	file.Read(currDirLenBytes)
	currDirStringLen := binary.LittleEndian.Uint32(currDirLenBytes)
	currDirBytes := make([]byte, currDirStringLen)
	file.Read(currDirBytes)
	fs.CurrentDir = string(currDirBytes)
	return fs
}

func SaveFileSystem(fs FileSystem) error {
	file, err := os.OpenFile(FileSystemBinary, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error creating file system: ", err)
		return err
	}

	defer file.Close()
	superblockBytes := encodeSuperBlock(fs.SuperBlock)
	file.Write(superblockBytes)

	freeListBytes := encodeFreeList(fs.FreeList)
	file.Write(freeListBytes)

	directoryBytes := encodeDirectory(fs.Directory)
	file.Write(directoryBytes)

	inodesBytes := encodeInodes(fs.Inodes)
	file.Write(inodesBytes)

	currDirBytes := encodeCurrDirecrory(fs.CurrentDir)
	file.Write(currDirBytes)

	return nil
}

func encodeSuperBlock(sb SuperBlock) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(sb.TotalBlocks))
	binary.Write(buf, binary.LittleEndian, uint32(sb.BlockSize))
	return buf.Bytes()
}

func encodeFreeList(fl []bool) []byte {
	buf := new(bytes.Buffer)
	for _, block := range fl {
		binary.Write(buf, binary.LittleEndian, block)
	}
	return buf.Bytes()
}

func encodeDirectory(directory []DirectoryEntry) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(directory)))
	for _, entry := range directory {
		binary.Write(buf, binary.LittleEndian, uint32(len(entry.Name)))
		buf.WriteString(entry.Name)
		binary.Write(buf, binary.LittleEndian, uint32(entry.Inode))
		buf.WriteString(entry.Permission)
	}
	return buf.Bytes()
}

func encodeInodes(inodes []Inode) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(inodes)))
	for _, inode := range inodes {
		dataLength := len(inode.Data)
		binary.Write(buf, binary.LittleEndian, uint32(dataLength))
		binary.Write(buf, binary.LittleEndian, inode.Data)
		binary.Write(buf, binary.LittleEndian, inode.Used)
		buf.WriteString(inode.Permission)
		binary.Write(buf, binary.LittleEndian, inode.CreatedAt.Unix())
		binary.Write(buf, binary.LittleEndian, inode.UpdatedAt.Unix())
	}
	return buf.Bytes()
}

func encodeCurrDirecrory(currDir string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(currDir)))
	buf.WriteString(currDir)
	return buf.Bytes()
}

func (fs *FileSystem) allocateInode(data []byte, permissions string) int {
	inode := Inode{
		Data:       data,
		Used:       true,
		Permission: permissions,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	index := findFreeInode(fs)
	if index != -1 {
		fs.Inodes[index] = inode
	}

	return index
}

func findFreeInode(fs *FileSystem) int {
	index := -1
	for i, fn := range fs.FreeList {
		if fn {
			index = i
			break
		}
	}

	if index == -1 {
		fmt.Println("No space left on FileSystem.")
		fmt.Println("Please try again later.")
		return index
	}

	InodesLen := len(fs.Inodes)
	if InodesLen-1 < index {
		fs.Inodes = append(fs.Inodes, make([]Inode, index+1-InodesLen)...)
	}

	fs.FreeList[index] = false

	return index
}
