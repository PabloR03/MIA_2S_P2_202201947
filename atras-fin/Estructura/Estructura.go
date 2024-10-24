package Estructura

import (
	"bytes"
	"fmt"
	"strings"
)

//  =================================Estructura MRB=================================

type MRB struct {
	MRBSize         int32
	MRBCreationDate [10]byte
	MRBSignature    int32
	MRBFit          [1]byte
	MRBPartitions   [4]Partition
}

func PrintMBR(buffer *bytes.Buffer, data MRB) {
	fmt.Fprintf(buffer, "\nFecha de Creación: %s, Ajuste: %s, Tamaño: %d, Identificador: %d\n",
		string(data.MRBCreationDate[:]), string(data.MRBFit[:]), data.MRBSize, data.MRBSignature)
	for i := 0; i < 4; i++ {
		PrintPartition(buffer, data.MRBPartitions[i])
	}
}

func PrintMBRnormal(data MRB) {
	println("\nFecha de Creación: %s, Ajuste: %s, Tamaño: %d, Identificador: %d\n",
		string(data.MRBCreationDate[:]), string(data.MRBFit[:]), data.MRBSize, data.MRBSignature)
	for i := 0; i < 4; i++ {
		PrintPartitionnormal(data.MRBPartitions[i])
	}
}

//  =================================Estructura Particion=================================

type Partition struct {
	PART_Status      [1]byte
	PART_Type        [1]byte
	PART_Fit         [1]byte
	PART_Start       int32
	PART_Size        int32
	PART_Name        [16]byte
	PART_Correlative int32
	PART_Id          [4]byte
	PART_Unit        [1]byte
	PART_Path        [100]byte
}

func PrintPartition(buffer *bytes.Buffer, data Partition) {
	fmt.Fprintf(buffer, "\nNombre: %s, Tipo: %s, Inicio: %d, Tamaño: %d, Estado: %s, ID: %s, Ajuste: %s, Correlativo: %d\n",
		strings.TrimRight(string(data.PART_Name[:]), "\x00"),
		strings.TrimRight(string(data.PART_Type[:]), "\x00"),
		data.PART_Start,
		data.PART_Size,
		strings.TrimRight(string(data.PART_Status[:]), "\x00"),
		strings.TrimRight(string(data.PART_Id[:]), "\x00"),
		strings.TrimRight(string(data.PART_Fit[:]), "\x00"),
		data.PART_Correlative)
}
func PrintPartitionnormal(data Partition) {
	println("\nNombre: %s, Tipo: %s, Inicio: %d, Tamaño: %d, Estado: %s, ID: %s, Ajuste: %s, Correlativo: %d\n",
		string(data.PART_Name[:]), string(data.PART_Type[:]), data.PART_Start, data.PART_Size, string(data.PART_Status[:]),
		string(data.PART_Id[:]), string(data.PART_Fit[:]), data.PART_Correlative)
}

//  =================================Estructura EBR=================================

type EBR struct {
	EBRMount [1]byte
	EBRFit   [1]byte
	EBRStart int32
	EBRSize  int32
	EBRNext  int32
	EBRName  [16]byte
}

func PrintEBR(buffer *bytes.Buffer, data EBR) {
	fmt.Fprintf(buffer, "\nMount: %s, Fit: %s, Start: %d, Size: %d, Next: %d, Name: %s\n",
		data.EBRMount, data.EBRFit, data.EBRStart, data.EBRSize, data.EBRNext, string(data.EBRName[:]))
}

func PrintEBRnormal(data EBR) {
	fmt.Printf("Mount: %s, Fit: %s, Start: %d, Size: %d, Next: %d, Name: %s\n",
		string(data.EBRMount[:]), string(data.EBRFit[:]), data.EBRStart, data.EBRSize, data.EBRNext, string(data.EBRName[:]))
}

// =================================Estuctura MountId=================================

type MountId struct {
	MIDpath   string
	MIDnumber int32
	MIDletter int32
}

// =================================Esctructuras para carpetas y Archivos=================================
// =================================Estuctura Superblock=================================

type SuperBlock struct {
	S_Filesystem_Type   int32
	S_Inodes_Count      int32
	S_Blocks_Count      int32
	S_Free_Blocks_Count int32
	S_Free_Inodes_Count int32
	S_Mtime             [17]byte
	S_Umtime            [17]byte
	S_Mnt_Count         int32
	S_Magic             int32
	S_Inode_Size        int32
	S_Block_Size        int32
	S_Fist_Ino          int32
	S_First_Blo         int32
	S_BM_Inode_Start    int32
	S_BM_Block_Start    int32
	S_Inode_Start       int32
	S_Block_Start       int32
}

func PrintSuperBlock(buffer *bytes.Buffer, data SuperBlock) {
	fmt.Fprint(buffer, "SUPERBLOQUE\n")
	fmt.Fprintf(buffer,
		"Filesystem Type: %d\n"+
			"Inodes Count: %d\n"+
			"Blocks Count: %d\n"+
			"Free Blocks Count: %d\n"+
			"Free Inodes Count: %d\n"+
			"Mtime: %s\n"+
			"Utime: %s\n"+
			"Mnt Count: %d\n"+
			"Magic: %d\n"+
			"Inode Size: %d\n"+
			"Block Size: %d\n"+
			"First Ino: %d\n"+
			"First Blo: %d\n"+
			"BM Inode Start: %d\n"+
			"BM Block Start: %d\n"+
			"Inode Start: %d\n"+
			"Block Start: %d\n",
		data.S_Filesystem_Type,
		data.S_Inodes_Count,
		data.S_Blocks_Count,
		data.S_Free_Blocks_Count,
		data.S_Free_Inodes_Count,
		data.S_Mtime[:],
		data.S_Umtime[:],
		data.S_Mnt_Count,
		data.S_Magic,
		data.S_Inode_Size,
		data.S_Block_Size,
		data.S_Fist_Ino,
		data.S_First_Blo,
		data.S_BM_Inode_Start,
		data.S_BM_Block_Start,
		data.S_Inode_Start,
		data.S_Block_Start,
	)
}
func PrintSuperBlocknormal(data SuperBlock) {
	println("\nFilesystem Type: %d, Inodes Count: %d, Blocks Count: %d, Free Blocks Count: %d, Free Inodes Count: %d, Mtime: %s, Utime: %s, Mnt Count: %d, Magic: %d, Inode Size: %d, Block Size: %d, Fist Ino: %d, First Blo: %d, BM Inode Start: %d, BM Block Start: %d, Inode Start: %d, Block Start: %d\n",
		data.S_Filesystem_Type, data.S_Inodes_Count, data.S_Blocks_Count, data.S_Free_Blocks_Count, data.S_Free_Inodes_Count,
		data.S_Mtime[:], data.S_Umtime[:], data.S_Mnt_Count, data.S_Magic, data.S_Inode_Size, data.S_Block_Size, data.S_Fist_Ino,
		data.S_First_Blo, data.S_BM_Inode_Start, data.S_BM_Block_Start, data.S_Inode_Start, data.S_Block_Start)
}

// =================================Estuctura Inode=================================

type Inode struct {
	I_Uid   int32
	I_Gid   int32
	I_Size  int32
	I_Atime [17]byte
	I_Ctime [17]byte
	I_Mtime [17]byte
	I_Block [15]int32
	I_Type  byte
	I_Perm  [3]byte
}

func PrintInode(buffer *bytes.Buffer, data Inode) {
	fmt.Fprintf(buffer, "\nINODO %d\nUID: %d \nGID: %d \nSIZE: %d \nACTUAL DATE: %s \nCREATION TIME: %s \nMODIFY TIME: %s \nBLOCKS:%d \nTYPE:%s \nPERM:%s \n",
		int(data.I_Gid),
		int(data.I_Uid),
		int(data.I_Gid),
		int(data.I_Size),
		data.I_Atime[:],
		data.I_Ctime[:],
		data.I_Mtime[:],
		data.I_Block[:],
		string(data.I_Type),
		string(data.I_Perm[:]),
	)
}

func PrintInodenormal(data Inode) {
	println("\nINODO %d\nUID: %d \nGID: %d \nSIZE: %d \nACTUAL DATE: %s \nCREATION TIME: %s \nMODIFY TIME: %s \nBLOCKS:%d \nTYPE:%s \nPERM:%s \n",
		int(data.I_Gid),
		int(data.I_Uid),
		int(data.I_Gid),
		int(data.I_Size),
		data.I_Atime[:],
		data.I_Ctime[:],
		data.I_Mtime[:],
		data.I_Block[:],
		string(data.I_Type),
		string(data.I_Perm[:]),
	)
}

// =================================Estuctura Fileblock=================================

type FileBlock struct {
	B_Content [64]byte
}

func PrintFileBlock(buffer *bytes.Buffer, data FileBlock) {
	fmt.Fprint(buffer, "File Block\n")
	fmt.Fprintf(buffer, "\nContent: %s\n", string(data.B_Content[:]))
	fmt.Println("=========================")
}

func PrintFileBlocknormal(data FileBlock) {
	println("File Block\n")
	println("\nContent: %s\n", string(data.B_Content[:]))
	println("=========================")

}

// ================================= BLOQUE =================================
// =================================Estuctura Folderblock=================================

type FolderBlock struct {
	B_Content [4]Content
}

func PrintFolderBlock(buffer *bytes.Buffer, data FolderBlock) {
	fmt.Fprint(buffer, "Folder Block\n")
	for i, content := range data.B_Content {
		fmt.Printf("Content %d: Name: %s, Inodo: %d\n", i, string(content.B_Name[:]), content.B_Inodo)
	}
	fmt.Println("=========================")
}

func PrintFolderBlocknormal(data FolderBlock) {
	println("Folder Block\n")
	for i, content := range data.B_Content {
		println("Content %d: Name: %s, Inodo: %d\n", i, string(content.B_Name[:]), content.B_Inodo)
	}
	println("=========================")
}

type Content struct {
	B_Name  [12]byte
	B_Inodo int32
}

type PointerBlock struct {
	B_Pointers [16]int32
}

func PrintPointerblock(buffer *bytes.Buffer, pointerblock PointerBlock) {
	fmt.Println("====== Pointerblock ======")
	for i, pointer := range pointerblock.B_Pointers {
		fmt.Fprintf(buffer, "\nPointer %d: %d\n", i, pointer)
	}
	fmt.Println("=========================")
}

func PrintPointerblocknormal(pointerblock PointerBlock) {
	println("====== Pointerblock ======")
	for i, pointer := range pointerblock.B_Pointers {
		println("\nPointer %d: %d\n", i, pointer)
	}
	println("=========================")
}

type Journaling struct {
	Size      int32
	Ultimo    int32
	Contenido [50]Content_J
}

type Content_J struct {
	Operation [10]byte
	Path      [100]byte
	Content   [100]byte
	Date      [17]byte
}
