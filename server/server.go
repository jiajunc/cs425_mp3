package main

import (
	// "bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const BUFFERSIZE = 1024

// isMaster := true

type MemberID struct {
	LocalIP    string
	JoinedTime time.Time
}

var memberList []MemberID

type IP string

var fileToNodes = struct {
	sync.RWMutex
	m map[string][]string
}{m: make(map[string][]string)}

var nodeToFiles = struct {
	sync.RWMutex
	m map[string][]string
}{m: make(map[string][]string)}

func TcpListening() {
	server, err := net.Listen("tcp", "localhost:27002")
	if err != nil {
		fmt.Println("Server: Error listetning: ", err)
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Server: Server init succeed!!")
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		fmt.Println("Server: Client connected")
		bufferLocalFileName := make([]byte, 64)
		bufferSdfsFileName := make([]byte, 65)
		bufferFileSize := make([]byte, 10)
		bufferRequest := make([]byte, 5)

		num, e := connection.Read(bufferFileSize)
		if e != nil {
			fmt.Println(num, e)
		}
		fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

		connection.Read(bufferLocalFileName)
		localFileName := strings.Trim(string(bufferLocalFileName), ":")
		fmt.Println("Server: local file:", localFileName)

		connection.Read(bufferSdfsFileName)
		sdfsFileName := strings.Trim(string(bufferSdfsFileName), ":")
		fmt.Println("Server: receiving file:", sdfsFileName)

		connection.Read(bufferRequest)
		request := strings.Trim(string(bufferRequest), ":")
		fmt.Println("Server: request state:", request)

		// newFile, err := os.Create(localFileName)
		newFile, err := os.Create(sdfsFileName)
		if err != nil {
			panic(err)
		}
		defer newFile.Close()
		var receivedBytes int64

		for {
			if (fileSize - receivedBytes) < BUFFERSIZE {
				io.CopyN(newFile, connection, (fileSize - receivedBytes))
				connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
				break
			}
			io.CopyN(newFile, connection, BUFFERSIZE)
			receivedBytes += BUFFERSIZE
		}

		fmt.Println("Received file completely!")
		// if request == '0'{
		// 	fmt.Println("Sending acks to master")
		// 	e := ackMaster(sdfsFileName, "local",memberList[0].LocalIP)
		// }
	}
}

func SendFileTo(address string, localfilename string, sdfsfilename string) error {
	connection, err := net.Dial("tcp", address+":27002")
	if err != nil {
		return err
	}
	defer connection.Close()
	fmt.Println("Client: Connected to server, start sending the file")
	e := sendFile(connection, localfilename, sdfsfilename)
	if e != nil {
		return e
	}
	return nil
}

func sendFile(connection net.Conn, localFileName string, sdfsFileName string) error {
	defer connection.Close()
	file, err := os.Open(sdfsFileName)
	if err != nil {
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	sdfsFileName = fillString(sdfsFileName, 64)
	localFileName = fillString(localFileName, 65)
	request := fillString("1", 5)
	fmt.Println("Client: Sending filename and filesize!")
	connection.Write([]byte(fileSize))
	connection.Write([]byte(localFileName))
	connection.Write([]byte(sdfsFileName))
	connection.Write([]byte(request))

	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Client: Start sending file!")
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	if err != nil {
		return err
	}
	fmt.Println("Client: File has been sent, closing connection!")
	return nil
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
	This server will send required file.
*/
func (t *IP) ReplyFile(args []string, ok *bool) error {
	sdfsFileName := args[0]
	localFileName := args[1]
	remoteIP := args[2]
	SendFileTo(remoteIP, localFileName, sdfsFileName)
	*ok = true
	return nil
}

/*
	If this is a "master" server,
		return the following 3 IPs of the given node-ip.
*/
func (t *IP) ReplyIPAddress(ip string, returnList *[]string) error {
	idx := 0
	for _, v := range memberList {
		if v.LocalIP == ip {
			break
		}
		idx++
	}
	for i := 1; i < 4; i++ {
		*returnList = append(*returnList, memberList[(idx+i)%len(memberList)].LocalIP)
	}
	return nil
}

/*
	If this is a "master" server,
		return all nodes that store the queryed file.
*/
func (t *IP) ReplyFilesNodes(fileName string, returnList *[]string) error {
	fileToNodes.RLock()
	v, ok := fileToNodes.m[fileName]
	if ok == false {
		fmt.Println("There is no such file...")
		return errors.New("There is no such file")
	}
	*returnList = v
	fileToNodes.RUnlock()
	return nil
}

/*
	If this is a "master" server,
		return files stored on the queryed node.
*/
func (t *IP) ReplyNodeFiles(nodeAddress string, files *[]string) error {
	nodeToFiles.RLock()
	v, ok := nodeToFiles.m[nodeAddress]
	if ok == false {
		fmt.Println("There is no such node")
		return errors.New("There is no such node")
	}
	*files = v
	nodeToFiles.RUnlock()
	return nil
}

func RespondIPListening() {
	IP_reply := new(IP)
	rpc.Register(IP_reply)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", "localhost:1105")
	if e != nil {
		log.Fatal("Listen error", e)
	}
	fmt.Println("1105 succeed")
	go http.Serve(l, nil)
	for {

	}
}

func (t *IP) DeleteFiles(fileName string, numb *int) error {
	err := os.Remove(fileName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

/*
   This function will be used to tell the master when a normal node received a file
*/
func ackMaster(sdfsFileName string, localIP string, masterAddress string) error {
	var n int
	var args []string
	args = append(args, sdfsFileName)
	args = append(args, localIP)
	client, e := rpc.DialHTTP("tcp", masterAddress+":1105")
	if e != nil {
		log.Fatal("Error when connect to master")
	}
	err := client.Call("IP.ReceivedAck", args, &n)
	if err != nil {
		log.Fatal("Can't send ack to master", err)
	}
	return nil
}

/*
   This function will be used to update the master store information
*/
func (t *IP) ReceivedAck(args []string, num *int) error {
	fileToNodes.RLock()
	nodeToFiles.RLock()
	fileToNodes.m[args[0]] = append(fileToNodes.m[args[0]], args[1])
	nodeToFiles.m[args[1]] = append(fileToNodes.m[args[1]], args[0])
	//func send repToMaster() to other master
	fileToNodes.RUnlock()
	nodeToFiles.RUnlock()
	fmt.Println("master received node successfully as bellow:")
	fmt.Println(fileToNodes)
	fmt.Println(nodeToFiles)
	return nil
}

/*
   This function will be used to send file to other master replica servers once master store new info.
*/
// func repToMaster() error{

// }
func main() {
	go TcpListening()
	var m1 = MemberID{LocalIP: "127.0.0.1"}
	// fileToNodes.m["dummy"] = append(fileToNodes.m["dummy"], "test")
	// fileToNodes.m["dummyfile.txt"] = append(fileToNodes.m["dummyfile.txt"], "localhost")
	// fileToNodes.m["receivedfile.txt"] = append(fileToNodes.m["receivedfile.txt"], "localhost")
	memberList = append(memberList, m1)
	RespondIPListening()
	ackMaster("testfile", "testlocalIP", "localhost")
}
