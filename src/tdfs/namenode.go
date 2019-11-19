package tdfs

import(
	// "time"
	"fmt"
	// "strconv"
	"github.com/gin-gonic/gin"
	"net/http"
	"io"
	"os"
	"bytes"
	"mime/multipart"
	"strings"
	"io/ioutil"
	"strconv"
	"encoding/json"
	"encoding/hex"
	"crypto/sha256"
)

// curl -X POST http://127.0.0.1:11090/putfile -F "putfile=@/Users/treasersmac/Programming/MilkPrairie/Gou/TinyDFS/SmallFile.txt" 
// -H "Content-Type: multipart/form-data"
func (namenode *NameNode) Run(){
	router := gin.Default()
	router.POST("/putfile", func(c *gin.Context) {
        // name := c.PostForm("name")
		// fmt.Println(name)
		filename := namenode.recvFrom(c) //
		fname := strings.Split(filename, ".")
		fmt.Println("putfile ...", fname)
		fmt.Println(fname)

		exist, err := PathExists(namenode.NAMENODE_DIR+"/"+fname[0])
		if err!=nil { 
			fmt.Println("XXX NameNode error at Get Dir", err.Error()) 
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		if !exist {
			err = os.MkdirAll(namenode.NAMENODE_DIR+"/"+fname[0], os.ModePerm)
			if err!=nil { 
				fmt.Println("XXX NameNode error at MkdirAll", err.Error()) 
				TDFSLogger.Fatal("XXX NameNode error: ", err)
			}
		}
		
		chunkLen, offsetLast := splitToFileAndStore(filename, namenode.NAMENODE_DIR+"/"+fname[0]+"/chunk-")

		var f File
		f.Info = "{name:"+fname[0]+"}"

		for i:=0; i<chunkLen; i++ {
			replicaLocationList := namenode.AllocateChunk()
			f.Chunks[i].ReplicaLocationList = replicaLocationList
			PutChunk(namenode.NAMENODE_DIR+"/"+fname[0]+"/chunk-"+strconv.Itoa(i), i, replicaLocationList)
		}
		f.Offset_LastChunk = offsetLast
		f.Length = offsetLast + chunkLen*SPLIT_UNIT
		// namenode.NameSpace[fname[0]] = f
		// fmt.Println("# - fname[0] = ", fname[0])

		// m := NameSpaceStruct{}
		ns := namenode.NameSpace
		ns[fname[0]] = f
		namenode.NameSpace = ns
		fmt.Printf("# FILE %s produced: chunkLen=%d, offsetLast=%d\n", filename, chunkLen, offsetLast)
		fmt.Println("## file.Info", ns[fname[0]].Info)
		for k:=0; k<chunkLen; k++{
			fmt.Println("## file.Chunks[", k ,"].Info", ns[fname[0]].Chunks[k].Info)
			fmt.Println("## file.Chunks[", k ,"].RLList", ns[fname[0]].Chunks[k].ReplicaLocationList)
		}
		fmt.Println("## file.Offset_LastChunk", ns[fname[0]].Offset_LastChunk)
		
		// 把NN中暂存客户端发来的数据删去：
		err = os.Remove(namenode.NAMENODE_DIR+"/"+filename)
		if err!=nil {
			fmt.Println("XXX NameNode error at remove tempfiles", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		
		for i:=0; i<chunkLen; i++{
			err = os.Remove(namenode.NAMENODE_DIR+"/"+fname[0]+"/chunk-"+strconv.Itoa(i))
			if err!=nil {
				fmt.Println("XXX NameNode error at remove chunk-", i, err.Error())
				TDFSLogger.Fatal("XXX NameNode error: ", err)
			}
		}

	})

	router.GET("/getfile/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		fmt.Println("$ getfile ...", filename)
		file := namenode.NameSpace[filename]
		// fmt.Println(file)
		/* 从DN中读取file所在的块数据，将块数据暂存在一个文件夹下 */
		for i:=0; i<file.Length/SPLIT_UNIT; i++{
			namenode.GetChunk(file, filename, i) 
		}
		/* 将文件夹下的块数据整合成一个文件 */
		// fdata := bytes.Join(filedata, nil)
		fdata := namenode.AssembleFile(file, filename)
		/* 将该文件发送给客户端 */
		c.String(http.StatusOK, string(fdata))
	})

	// router.DELETE GET
	router.DELETE("/delfile/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		file := namenode.NameSpace[filename]
		for i:=0; i<file.Length/SPLIT_UNIT; i++{
			namenode.DelChunk(file, filename, i)
		}
		c.String(http.StatusOK, "DelFile:"+filename+" SUCCESS\n")
	})

	router.GET("/test", func(c *gin.Context) {
		namenode.GetDNMeta() //namenode.ShowInfo()
	})

	router.Run(":"+strconv.Itoa(namenode.Port))
}

func (namenode *NameNode)AssembleFile(file File, filename string) ([]byte){
	fmt.Println("@ AssembleFile of ",filename)
	filedata := make([][]byte, file.Length/SPLIT_UNIT)
	for i:=0; i<file.Length/SPLIT_UNIT; i++{
		// fmt.Println("& Assmble chunk-",i)
		b := readFileByBytes(namenode.NAMENODE_DIR+"/"+filename+"/chunk-"+strconv.Itoa(i))
		// fmt.Println("& Assmble chunk b=", b)
		filedata[i] = make([]byte, 100)
		filedata[i] = b
		// fmt.Println("& Assmble chunk filedata[i]=", filedata[i])
		// fmt.Println("& Assmble chunk-",i)
	}
	fdata := bytes.Join(filedata, nil)
	FastWrite(namenode.NAMENODE_DIR+"/nn-"+filename, fdata)
	return fdata
}


func (namenode *NameNode) GetChunk(file File, filename string, num int){//ChunkUnit chunkbytes []byte
	
	fmt.Println("* getting chunk-",num, "of file:", filename)

	for i:=0; i<REDUNDANCE ; i++{
		replicalocation := file.Chunks[num].ReplicaLocationList[i].ServerLocation
		repilcanum := file.Chunks[num].ReplicaLocationList[i].ReplicaNum
		/* send Get chunkdata request */
		url := replicalocation+"/getchunk/"+strconv.Itoa(repilcanum)
		dataRes, err := http.Get(url)
		if err != nil {
			fmt.Println("XXX NameNode error at Get chunk of ", file.Info, ": ", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		defer dataRes.Body.Close()
		/* deal response of Get */
		chunkbytes, err := ioutil.ReadAll(dataRes.Body)
		if err != nil {
			fmt.Println("XXX NameNode error at ReadAll response of chunk", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		fmt.Println("** DataNode Response of Get chunk-",num,": ", string(chunkbytes))
		/* store chunkdata at nn local */
		FastWrite(namenode.NAMENODE_DIR+"/"+filename+"/chunk-"+strconv.Itoa(num), chunkbytes)
		
		/* send Get chunkhash request */
		hashRes, err := http.Get(replicalocation+"/getchunkhash/"+strconv.Itoa(repilcanum))
		if err != nil {
			fmt.Println("XXX NameNode error at Get chunkhash", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		defer hashRes.Body.Close()
		/* deal Get chunkhash request */
		chunkhash, err := ioutil.ReadAll(hashRes.Body)
		if err != nil {
			fmt.Println("XXX NameNode error at Read response of chunkhash", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		
		/* check hash */
		// chunkfile := OpenFile(namenode.NAMENODE_DIR+"/"+filename+"/chunk-"+strconv.Itoa(chunknum))
		hash := sha256.New()
		hash.Write(chunkbytes)
		// if _, err := io.Copy(hash, chunkfile); err != nil {fmt.Println("XXX NameNode error at sha256", err.Error())}
		hashStr := hex.EncodeToString(hash.Sum(nil))
		fmt.Println("*** chunk hash calculated: ", hashStr)
		fmt.Println("*** chunk hash get: ", string(chunkhash))

		if hashStr==string(chunkhash) {
			break
		}else{
			fmt.Println("X=X the first replica of chunk-", num, "'s hash(checksum) is WRONG, continue to request anothor replica...")
			continue
		}
	}
}

func (namenode *NameNode) DelChunk(file File, filename string, num int){//ChunkUnit chunkbytes []byte
	fmt.Println("** deleting chunk-",num, "of file:", filename)
	for i:=0; i<REDUNDANCE; i++{
		chunklocation := file.Chunks[num].ReplicaLocationList[i].ServerLocation
		chunknum := file.Chunks[num].ReplicaLocationList[i].ReplicaNum
		url := chunklocation+"/delchunk/"+strconv.Itoa(chunknum)

		// response, err := http.Get(url)
		c := &http.Client{}
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			fmt.Println("XXX NameNode error at Del chunk of ", file.Info, ": ", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		
		response, err := c.Do(req)
		if err != nil {
			fmt.Println("XXX NameNode error at Del chunk(Do):", err.Error())
			TDFSLogger.Fatal("XXX NameNode error at Del chunk(Do):",err)
		}
		defer response.Body.Close()

		/** Read response **/
		delRes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("XXX NameNode error at Read response", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		fmt.Println("*** DataNode Response of Delete chunk-",num,"replica-",i,": ",string(delRes))
		// return chunkbytes
	}
}


func (namenode *NameNode) AllocateChunk() (rlList [REDUNDANCE]ReplicaLocation){
	redundance := namenode.REDUNDANCE
	// fmt.Println("# redundance:",redundance)
	var max [REDUNDANCE]int
	for i:=0; i<redundance; i++{
		max[i] = 0
		for j:=0; j<namenode.DNNumber; j++{
			if(namenode.DataNodes[j].StorageAvail > namenode.DataNodes[max[i]].StorageAvail){
				max[i] = j
			}
		}
		rlList[i].ServerLocation = namenode.DataNodes[max[i]].Location
		rlList[i].ReplicaNum = namenode.DataNodes[max[i]].ChunkAvail[0]
		n := namenode.DataNodes[max[i]].StorageAvail
		// fmt.Println("*** n = ", n)
		// fmt.Println("namenode.DataNodes[max[i]].ChunkAvail[0]", namenode.DataNodes[max[i]].ChunkAvail[0])
		// fmt.Println("namenode.DataNodes[max[i]].ChunkAvail[n-1]", namenode.DataNodes[max[i]].ChunkAvail[n-1])
		namenode.DataNodes[max[i]].ChunkAvail[0] = namenode.DataNodes[max[i]].ChunkAvail[n-1]
		namenode.DataNodes[max[i]].ChunkAvail = namenode.DataNodes[max[i]].ChunkAvail[0:n-1]
		namenode.DataNodes[max[i]].StorageAvail--
	}

	// fmt.Println("max[0-2]: ",max)
	return rlList
}

func (namenode *NameNode) Reset(){

	// CleanFile("TinyDFS/DataNode1/chunk-"+strconv.Itoa(i))
	fmt.Println("# Reset...")

	err := os.RemoveAll(namenode.NAMENODE_DIR+"/")
	if err!=nil {
		fmt.Println("XXX NameNode error at RemoveAll dir", err.Error())
		TDFSLogger.Fatal("XXX NameNode error: ", err)
	}

	err = os.MkdirAll(namenode.NAMENODE_DIR, 0777)
	if err!=nil {
		fmt.Println("XXX NameNode error at MkdirAll", err.Error())
		TDFSLogger.Fatal("XXX NameNode error: ", err)
	}


}

func (namenode *NameNode) SetConfig(location string, dnnumber int,  redundance int, dnlocations []string){
	temp := strings.Split(location, ":")
	res, err := strconv.Atoi(temp[2])
	if err!=nil{
		fmt.Println("XXX NameNode error at Atoi parse Port", err.Error())
		TDFSLogger.Fatal("XXX NameNode error: ", err)
	}

	ns := NameSpaceStruct{}
	namenode.NameSpace = ns
	namenode.Port = res
	namenode.Location = location
	namenode.DNNumber = dnnumber
	namenode.DNLocations = dnlocations
	namenode.REDUNDANCE = redundance
	fmt.Println("************************************************************")
	fmt.Println("************************************************************")
	fmt.Printf("*** Successfully Set Config data for the namenode\n")
	namenode.ShowInfo()
	fmt.Println("************************************************************")
	fmt.Println("************************************************************")
}

func (namenode *NameNode) ShowInfo(){
	fmt.Println("************************************************************")
	fmt.Println("****************** showinf for NameNode ********************")
	fmt.Printf("Location: %s\n", namenode.Location)
	fmt.Printf("DATANODE_DIR: %s\n", namenode.NAMENODE_DIR)
	fmt.Printf("Port: %d\n", namenode.Port)
	fmt.Printf("DNNumber: %d\n", namenode.DNNumber)
	fmt.Printf("REDUNDANCE: %d\n", namenode.REDUNDANCE)
	fmt.Printf("DNLocations: %s\n", namenode.DNLocations)
	fmt.Printf("DataNodes: ")
	fmt.Println(namenode.DataNodes)
	fmt.Println("******************** end of showinfo ***********************")
	fmt.Println("************************************************************")
}

func (namenode *NameNode) GetDNMeta(){ // UpdateMeta
	for i:=0; i<len(namenode.DNLocations); i++ {
		response, err := http.Get(namenode.DNLocations[i]+"/getmeta")
		if err != nil {
			fmt.Println("XXX NameNode error at Get meta of ", namenode.DNLocations[i],": ", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		defer response.Body.Close()
		// bytes, err := ioutil.ReadAll(response.Body)
		// if err != nil {fmt.Println("NameNode error at read responsed meta of ", namenode.DNLocations[i],": ", err.Error())}
		// fmt.Print("datanode", i, " metadata: ")
		// fmt.Printf("%s\n", bytes)

		var dn DataNode
		err = json.NewDecoder(response.Body).Decode(&dn)
		if err != nil {
			fmt.Println("XXX NameNode error at decode response to json.", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		// fmt.Println(dn)
		// err = json.Unmarshal([]byte(str), &dn)
		namenode.DataNodes = append(namenode.DataNodes, dn)
	}
	namenode.ShowInfo()
}

func (namenode *NameNode) PutDNMeta(){
	// 把namenode.DataNodes传给各个DataNodes
	// 没必要了，因为给DN发送存储位置信息时现在DN会进行更新了。
}

func (namenode *NameNode) recvFrom(c *gin.Context) (string){
	file, header, err := c.Request.FormFile("putfile") // no multipart boundary param in Content-Type
	if err != nil {
		c.String(http.StatusBadRequest, "Bad request")
		fmt.Println("XXX NameNode error at Request FormFile", err.Error())
		TDFSLogger.Fatal("XXX NameNode error: ", err)
		return "null"
	}
	filename := header.Filename

	fmt.Println()
	fmt.Println(file, err, filename)
	fmt.Println()

	out, err := os.Create(namenode.NAMENODE_DIR+"/"+filename) //在服务器本地新建文件进行存储
	if err != nil {
		c.String(http.StatusBadRequest, "NameNode error at Create file")
		fmt.Println("XXX NameNode error at Create", err.Error())
		TDFSLogger.Fatal("XXX NameNode error: ", err)
		return "null"
	}

	io.Copy(out, file) //在服务器本地新建文件进行存储
	c.String(http.StatusCreated, "PutFile SUCCESS\n")
	out.Close()//defer 

	return filename
}

func PutChunk(dataChunkPath string, chunkNum int, replicaLocationList [REDUNDANCE]ReplicaLocation){

	replicaLen := len(replicaLocationList)

	for i:=0; i<replicaLen; i++{
		fmt.Printf("* Putting [Chunk %d] to TDFS [DataNode: %s] at [Replica-%d].\n", 
		chunkNum, replicaLocationList[0].ServerLocation, replicaLocationList[0].ReplicaNum) //  as %s , fName

		/** Create form file **/
		buf := new(bytes.Buffer)
		writer := multipart.NewWriter(buf)
		formFile, err := writer.CreateFormFile("putchunk", dataChunkPath)
		if err != nil {
			fmt.Println("XXX NameNode error at Create form file", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		
		// fmt.Println("*** Create Form File OK ")

		/** Open source file **/
		srcFile, err := os.Open(dataChunkPath)
		if err != nil {
			fmt.Println("XXX NameNode error at Open source file", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		defer srcFile.Close()

		// fmt.Println("*** Open source file OK ")
		
		/** Write to form file **/
		_, err = io.Copy(formFile, srcFile)
		if err != nil {
			fmt.Println("XXX NameNode error at Write to form file", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}

		// fmt.Println("*** Write to form file OK ")

		/** Set Params Before Post **/
		params := map[string]string{
			"ReplicaNum" : strconv.Itoa(replicaLocationList[i].ReplicaNum),  //chunkNum
		}
		for key, val := range params {
			err = writer.WriteField(key, val)
			if err != nil {
				fmt.Println("XXX NameNode error at Set Params", err.Error())
				TDFSLogger.Fatal("XXX NameNode error: ", err)
			}
		}
		contentType := writer.FormDataContentType()
		writer.Close() // 发送之前必须调用Close()以写入结尾行

		// fmt.Println("*** Set Params OK ")

		/** Post form file **/
		res, err := http.Post(replicaLocationList[i].ServerLocation+"/putchunk", 
								contentType, buf) // /"+strconv.Itoa(chunkNum)
		if err != nil {
			fmt.Println("XXX NameNode error at Post form file", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		defer res.Body.Close()

		fmt.Println("** Post form file OK ")

		/** Read response **/
		response, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("XXX NameNode error at Read response", err.Error())
			TDFSLogger.Fatal("XXX NameNode error: ", err)
		}
		fmt.Print("*** DataNoed Response: ",string(response))
	}
}

