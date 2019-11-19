package main

import(
	// "fmt"
	"tdfs"
	// "runtime"
	// "sync"
)

const DN2_DIR string = "TinyDFS/DataNode2"
const DN2_LOCATION string = "http://localhost:11092"
const DN2_CAPACITY int = 100

func main() {
	var dn2 tdfs.DataNode
	dn2.DATANODE_DIR = DN2_DIR
	
	dn2.Reset()
	dn2.SetConfig(DN2_LOCATION, DN2_CAPACITY)

	dn2.Run()
}