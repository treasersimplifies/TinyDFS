package main

import(
	"fmt"
	"tdfs"
	"flag"
	// "runtime"
	// "sync"
)

func main() {
	
	/* 需要将文件放到$GOPATH下 */
	// go run TinyDFS/Client.go -putfile "SmallFile.txt"
	// go run TinyDFS/Client.go -getfile "SmallFile"
	// go run TinyDFS/Client.go -delfile "SmallFile"

	// go run TinyDFS/Client.go -putfile "AFile.txt"
	// go run TinyDFS/Client.go -getfile "AFile"
	// go run TinyDFS/Client.go -delfile "AFile"

	// go run TinyDFS/Client.go -putfile "BFile.txt"
	// go run TinyDFS/Client.go -getfile "BFile"
	// go run TinyDFS/Client.go -delfile "BFile"

	var client tdfs.Client
	client.SetConfig("http://localhost:11090")

	filenameOfGet := flag.String("getfile", "unknow", "the filename of the file you want to get") // SmallFile
	filenameOfPut := flag.String("putfile", "unknow", "the filename of the file you want to put") // SmallFile.txt
	filenameOfDel := flag.String("delfile", "unknow", "the filename of the file you want to del")

	flag.Parse()
	
	if *filenameOfPut!="unknow" {
		client.PutFile(*filenameOfPut)
		fmt.Println(" -PutFile for ", *filenameOfPut)
	}
	
	if *filenameOfGet!="unknow" {
		client.GetFile(*filenameOfGet)
		fmt.Println(" -Getfile for ", *filenameOfGet)
	}

	if *filenameOfDel!="unknow" {
		client.DelFile(*filenameOfDel)
		fmt.Println(" -Delfile for ", *filenameOfDel)
	}

	// fmt.Println(flag.Args())
	//, "smallfile.txt" /Users/treasersmac/Programming/MilkPrairie/Gou/TinyDFS/
	// client.Test()
}