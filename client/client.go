package main

import (
	"bufio"
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
			delete(userInput[1])
		case "ls":
			if len(userInput) != 2 {
				fmt.Println("Wrong pattern! Enter 'ls sdfsfilename' to search machines.")
			}
		case "store":
			if len(userInput) != 1 {
				fmt.Println("Wrong pattern! Enter 'store' to search files.")
			}
		case "exit":
			return
		default:
			fmt.Println("Wrong input! Please try again:")
			fmt.Println("Enter 'put localfilename sdfsfilename' to upload file.")
			fmt.Println("Enter 'get sdfsfilename localfilename' to fetch file.")
			fmt.Println("Enter 'delete sdfsfilename' to delete file.")
			fmt.Println("Enter 'ls sdfsfilename' to list all addresses the file currently stored.")
			fmt.Println("Enter 'store' list all files stored at this machine.")
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

func SendFileTo(address string, localfilename string, sdfsfilename string) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Println("Client: Connected to server, start sending the file")
	sendFile(connection, localfilename, sdfsfilename)
}

func put(localfilename string, sdfsfilename string, addresses []string) {
	//should be one more paramater for function: addresses []string
	// addresses := []string{"localhost:27001"}
	for _, address := range addresses {
		SendFileTo(address+":27001", localfilename, sdfsfilename)
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
	fmt.Println(nodes)
	return nodes, nil
}

// func get(sdfsFileName string, localFileName string) error {
// nodes, err := getFileNodes(sdfsFileName)
// if err != nil {
// 	log.Fatal("error when get...", err)
// 	return err
// }
// 	return nil
// }

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

func delete(sdfsFileName string) error {
	nodes, err := getFileNodes(sdfsFileName)
	if err != nil {
		log.Fatal("error when get...", err)
		return err
	}
	err = deleteRequest(nodes, sdfsFileName)
	if err != nil {
		fmt.Println("Delet File Successfully")
	}
	return err
}

func deleteRequest(nodes []string, sdfsFileName string) error {

	for _, node := range nodes {
		var n int
		client, e := rpc.DialHTTP("tcp", node+":1105")
		if e != nil {
			log.Fatal("Error when dial")
			return e
		}
		err := client.Call("IP.DeleteFiles", sdfsFileName, &n)
		if err != nil {
			log.Fatal("Delete file fatal:")
			return e
		}
	}
	return nil
}

func main() {
	//input "put dummyfile.txt receivedfile.txt" to test put
	//input "delete receivedfile.txt" to test put
	initi()

	// showLocalStoredFiles()
	// getFileNodes("dummy")
	// delete("dummyfile.txt")
}
