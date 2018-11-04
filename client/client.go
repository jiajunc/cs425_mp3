package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var masterAddress = "127.0.0.1"

const BUFFERSIZE = 1024

type MemberID struct {
	LocalIP    string
	JoinedTime time.Time
}

//cli api
func initi() {
	/* start at the beginning, wait for user command to get/put/delete file*/
	fmt.Println("Enter command to put, get, update or delete file")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := strings.Split(scanner.Text(), " ")
		userCommand := userInput[0]

		switch userCommand {
		case "put":
			if len(userInput) != 3 {
				fmt.Println("Wrong pattern! Enter 'put localfilename sdfsfilename' to upload file.")
			}
			addresses := getIP(masterAddress)
			fmt.Println("excuting Put method")
			put(userInput[1], userInput[2], addresses)
		case "get":
			if len(userInput) != 3 {
				fmt.Println("Wrong pattern! Enter 'get sdfsfilename localfilename' to fetch file.")
			}
			//get()
		case "delete":
			if len(userInput) != 2 {
				fmt.Println("Wrong pattern! Enter 'delete sdfsfilename' to delete file.")
			}
			//delete()
		case "exit":
			return
		default:
			fmt.Println("Wrong input! Please try again:")
			fmt.Println("Enter 'put localfilename sdfsfilename' to upload file.")
			fmt.Println("Enter 'get sdfsfilename localfilename' to fetch file.")
			fmt.Println("Enter 'delete sdfsfilename' to delete file.")
		}
	}
}

func sendFile(connection net.Conn, localfilename string, sdfsfilename string) {
	defer connection.Close()
	file, err := os.Open(localfilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	localfileName := fillString(fileInfo.Name(), 64)
	receivedfileName := fillString(sdfsfilename, 65)
	fmt.Println("Client: Sending filename and filesize!")
	connection.Write([]byte(fileSize))
	connection.Write([]byte(localfileName))
	connection.Write([]byte(receivedfileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Client: Start sending file!")
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	fmt.Println("Client: File has been sent, closing connection!")
	return
}

func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}

/*
	This function is not okay. The parameters name are confused.
		It will append a timestamp to the file name which indicates the version.
*/
func SendFileTo(address string, localfilename string, sdfsfilename string) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	t := time.Now()
	tUnix := int(t.Unix())
	sdfsfilename += "-" + strconv.Itoa(tUnix)
	fmt.Println("Client: Connected to server, start sending the file")
	sendFile(connection, localfilename, sdfsfilename)
}

/*
	Insert or update a file on sdfs.

*/
func put(localfilename string, sdfsfilename string) {
	//should be one more paramater for function: addresses []string
	// addresses := []string{"localhost:27001"}
	nodes, err := getFileNodes(sdfsfilename)
	// Insert a new file
	if err == nil {
		for _, address := range nodes {
			SendFileTo(address+":27001", localfilename, sdfsfilename)
		}
	} else { // update a file
		files, e := getNodeFiles(nodes[0])
		if e != nil {
			log.Fatal(e)
		}

	}

}

func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

/*
	This function will return following 3 IPs of the given parameter-address.
*/
func getIP(masterAddress string) []string {
	var localIP string
	localIP = GetLocalIP().String()
	var list []string
	client, e := rpc.DialHTTP("tcp", masterAddress+":1105")
	if e != nil {
		log.Fatal("Error when dial")
	}
	err := client.Call("IP.ReplyIPAddress", localIP, &list)
	if err != nil {
		log.Fatal("Reply from master error", err)
	}
	fmt.Println(list)
	return list
}

/*
	Command "ls"
	This function will return non-nil error when there is no such file in sdfs.
*/
func getFileNodes(fileName string) ([]string, error) {
	var nodes []string
	client, e := rpc.DialHTTP("tcp", "localhost:1105")
	if e != nil {
		log.Fatal("Error when dial")
		return nodes, e
	}
	err := client.Call("IP.ReplyFilesNodes", fileName, &nodes)
	if err != nil {
		log.Fatal("Reply from master error", err)
		return nodes, err
	}
	// fmt.Println(nodes)
	return nodes, nil
}

/*
	This function will return all files stored on a machine.
*/
func getNodeFiles(nodeAddress string) ([]string, error) {
	var files []string
	client, e := rpc.DialHTTP("tcp", "localhost:1105")
	if e != nil {
		log.Fatal("Error when dial")
		return nil, e
	}
	e = client.Call("IP.ReplyNodeFiles", nodeAddress, &files)
	if e != nil {
		log.Fatal("Reply from server error", e)
		return nil, e
	}
	return files, nil
}

/*
	Get the latest version of "sdfsFileName", saving as "localFilename".

	Note: Have not implemented "latest"
*/

func get(sdfsFileName string, localFileName string) error {
	// nodes, err := getFileNodes(sdfsFileName)
	_, err := getFileNodes(sdfsFileName)
	if err != nil {
		log.Fatal("error when get...", err)
		return err
	}
	// localIP := GetLocalIP().String()
	client, e := rpc.DialHTTP("tcp", "localhost:1105")
	if e != nil {
		log.Fatal("Error when dial")
		return e
	}
	var args []string
	args = append(args, sdfsFileName)
	args = append(args, localFileName)
	args = append(args, "127.0.0.1")
	var reply bool
	err = client.Call("IP.ReplyFile", args, &reply)
	if err != nil {
		log.Fatal("Reply file get error...", err)
		return err
	}
	if reply != true {
		log.Fatal("Send file failure")
		return errors.New("Send file failure")
	}
	return nil
}

func showLocalStoredFiles() []string {
	fileInfo, err := ioutil.ReadDir("../server")
	if err != nil {
		log.Fatal(err)
	}
	var files []string
	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	fmt.Println(files)
	return files
}

func main() {
	// initi()
	// type "put dummyfile.txt receivedfile.txt" for test
	// getFileNodes("dummy")
	// showLocalStoredFiles()
	get("dummyFile.txt", "dummy")
}
