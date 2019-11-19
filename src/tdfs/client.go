package tdfs

import(
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
	"mime/multipart"
	"bytes"
	"io"
)

func (client *Client) PutFile(fPath string){ //, fName string
	fmt.Println("****************************************")
	fmt.Printf("*** Putting ${GOPATH}/%s to TDFS [NameNode: %s] )\n", fPath, client.NameNodeAddr) //  as %s , fName

	client.uploadFileByMultipart(fPath)

	fmt.Println("****************************************")
}

func (client *Client) GetFile(fName string){ //, fName string
	fmt.Println("****************************************")
	fmt.Printf("*** Getting from TDFS [NameNode: %s] to ${GOPATH}/%s )\n", client.NameNodeAddr, fName) //  as %s , fName

	response, err := http.Get(client.NameNodeAddr + "/getfile/" + fName)
	if err != nil {
		fmt.Println("XXX Client error at Get file", err.Error())
		TDFSLogger.Fatal("XXX Client error at Get file",err)
	}

    defer response.Body.Close()

    bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("XXX Client error at read response data", err.Error())
		TDFSLogger.Fatal("XXX Client error at read response data",err)
	}

	err = ioutil.WriteFile("local-"+fName, bytes, 0666)
	if err != nil { 
		fmt.Println("XXX Client error at store file", err.Error())
		TDFSLogger.Fatal("XXX Client error at store file",err)
	}
	
	fmt.Print("*** NameNode Response: ",string(bytes))
	fmt.Println("****************************************")
}

func (client *Client) DelFile(fName string){
	fmt.Println("****************************************")
	fmt.Printf("*** Deleting from TDFS [NameNode: %s] of /%s )\n", client.NameNodeAddr, fName)

	// Create client
	c := &http.Client{}
	// Create request
	req, err := http.NewRequest("DELETE", client.NameNodeAddr + "/delfile/" + fName, nil)
	if err != nil {
		fmt.Println("XXX Client error at del file(NewRequest):", err.Error())
		TDFSLogger.Fatal("XXX Client error at del file(NewRequest):",err)
	}
	// Fetch Request
	response, err := c.Do(req)
	if err != nil {
		fmt.Println("XXX Client error at del file(Do):", err.Error())
		TDFSLogger.Fatal("XXX Client error at del file(Do):",err)
	}
	defer response.Body.Close()
	// Read Response Body
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("XXX Client error at read response data", err.Error())
		TDFSLogger.Fatal("XXX Client error at read response data",err)
	}

	// // Display Results
	// fmt.Println("response Status : ", resp.Status)
	// fmt.Println("response Headers : ", resp.Header)
	// fmt.Println("response Body : ", string(respBody))


	// // response, err := http.Delete(client.NameNodeAddr + "/delfile/" + fName) //Get
	// if err != nil {
	// 	fmt.Println("XXX Client error at del file", err.Error())
	// 	TDFSLogger.Fatal("XXX Client error at del file",err)
	// }

    // defer response.Body.Close()

    // bytes, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	fmt.Println("XXX Client error at read response data", err.Error())
	// 	TDFSLogger.Fatal("XXX Client error at read response data",err)
	// }

	fmt.Print("*** NameNode Response: ",string(bytes))
	fmt.Println("****************************************")
}

func (client *Client) Test(){
	_, err := http.Get(client.NameNodeAddr+"/test")
	if err != nil {
		fmt.Println("Client error at Get test", err.Error())
	}
}

func (client *Client) uploadFileByMultipart(fPath string){
	buf := new(bytes.Buffer)
    writer := multipart.NewWriter(buf)
    formFile, err := writer.CreateFormFile("putfile", fPath)
	if err != nil {
		fmt.Println("XXX Client error at Create form file", err.Error())
		TDFSLogger.Fatal("XXX Client error at Create form file",err)
	}
	
	srcFile, err := os.Open(fPath)
    if err != nil {
		fmt.Println("XXX Client error at Open source file", err.Error())
		TDFSLogger.Fatal("XXX Client error at Open source file",err)
    }
	defer srcFile.Close()
	
    _, err = io.Copy(formFile, srcFile)
    if err != nil {
		fmt.Println("XXX Client error at Write to form file", err.Error())
		TDFSLogger.Fatal("XXX Client error at Write to form file",err)
    }

	contentType := writer.FormDataContentType()
    writer.Close() // 发送之前必须调用Close()以写入结尾行
    res, err := http.Post(client.NameNodeAddr+"/putfile", contentType, buf)
    if err != nil {
		fmt.Println("XXX Client error at Post form file", err.Error())
		TDFSLogger.Fatal("XXX Client error at Post form file",err)
	}
	defer res.Body.Close()

    content, err := ioutil.ReadAll(res.Body)
    if err != nil {
		fmt.Println("XXX Client error at Read response", err.Error())
		TDFSLogger.Fatal("XXX Client error at Read response",err)
	}
	
    fmt.Println("*** NameNode Response: ",string(content))
}

func (client *Client) SetConfig(nnaddr string) {
	client.NameNodeAddr = nnaddr
}

/* Allocation or AskInfo
 * POST (fileName string, fileBytes int)
 * Wait ReplicaList []ReplicaLocation   */
func RequestInfo(fileName string, fileBytes int) ([]ReplicaLocation){
	/* POST and Wait */
	replicaLocationList := []ReplicaLocation{
		ReplicaLocation{"http://localhost:11091", 3,}, 
		ReplicaLocation{"http://localhost:11092", 5,},
	}
	return replicaLocationList
}

func uploadFileByBody(client *Client, fPath string){
	file, err := os.Open(fPath)
    if err != nil {
		fmt.Println("XXX Client Fatal error at Open uploadfile", err.Error())
		TDFSLogger.Fatal("XXX Client error at Open uploadfile",err)
    }
    defer file.Close()

    res, err := http.Post(client.NameNodeAddr+"/putfile", "multipart/form-data", file) //  
    if err != nil {
		fmt.Println("Client Fatal error at Post", err.Error())
		TDFSLogger.Fatal("XXX Client error at at Post",err)
    }
    defer res.Body.Close()

    content, err := ioutil.ReadAll(res.Body)
    if err != nil {
		fmt.Println("Client Fatal error at Read response", err.Error())
		TDFSLogger.Fatal("XXX Client error at Read response",err)
    }
    fmt.Println(string(content))
}