package main

import(
	// "fmt"
	"tdfs"
	// "runtime"
	// "sync"
)

const DN1_DIR string = "TinyDFS/DataNode1"
const DN1_LOCATION string = "http://localhost:11091"
const DN1_CAPACITY int = 100

func main() {
	var dn1 tdfs.DataNode
	dn1.DATANODE_DIR = DN1_DIR
	
	dn1.Reset()
	dn1.SetConfig(DN1_LOCATION, DN1_CAPACITY)

	dn1.Run()
}