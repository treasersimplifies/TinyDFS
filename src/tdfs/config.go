package tdfs

/** Configurations for Pseudo Distributed Mode **/

/** Configurations for ALL Mode **/
const SPLIT_UNIT int = 100
const REDUNDANCE int = 2
const CHUNKTOTAL int = 100


// Chunk 一律表示逻辑概念，表示文件块
// Replica 表示文件块副本，是实际存储的
/** Data Structure **/
type ChunkUnit []byte // SPLIT_UNIT
// type ChunkReplicaOfFile map[int]FileChunk

// type FileChunk struct{
// 	Filename string
// 	ChunkNum int
// }

////// File to Chunk
type NameSpaceStruct map[string]File
type File struct{
	Info string // file info
	Length int
	Chunks [CHUNKTOTAL]FileChunk
	Offset_LastChunk int
}
type FileChunk struct{
	Info string // checksum
	ReplicaLocationList [REDUNDANCE]ReplicaLocation
}
type ReplicaLocation struct{
	ServerLocation string 
	ReplicaNum int
}


type Client struct{
	NameNodeAddr string
	Mode int
}

type Config struct{
	NameNodeAddr string
}

type NameNode struct{
	NameSpace NameSpaceStruct
	Location string
	Port int
	DNNumber int
	DNLocations []string
	DataNodes []DataNode
	NAMENODE_DIR string 
	REDUNDANCE int
}
type DataNode struct{
	Location string  	`json:"Location"` // http://IP:Port/
	Port int		 	`json:"Port"`
	StorageTotal int 	`json:"StorageTotal"`// a chunk as a unit
	StorageAvail int 	`json:"StorageAvail"`
	ChunkAvail []int 	`json:"ChunkAvail"`//空闲块表
	LastEdit int64		`json:"LastEdit"`
	DATANODE_DIR string `json:"DATANODE_DIR"`
}
type DNMeta struct{
	StorageTotal int `json:"StorageTotal"`
	StorageAvail int
	ChunkAvail []int 
	LastEdit int64
}



func (conf *Config) Set(addr string){
	conf.NameNodeAddr = addr
}
