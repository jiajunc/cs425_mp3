package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

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
			if len(userInput) < 3 {
				fmt.Println("Wrong pattern! Enter 'put localfilename sdfsfilename' to upload file.")
			}
			//target = findTarget() //ask for master target
			//put(localfilename string, sdfsfilename string, memberlist *[]MemberID, consistency int)
		case "get":
			if len(userInput) < 3 {
				fmt.Println("Wrong pattern! Enter 'get sdfsfilename localfilename' to fetch file.")
			}
			//get()
		case "delete":
			if len(userInput) < 2 {
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

func SendFileTo(address string, localfilename string, sdfsfilename string) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Println("Client: Connected to server, start sending the file")
	sendFile(connection, localfilename, sdfsfilename)
}

// func put(localfilename string, sdfsfilename string) {
// 	//put(localfilename string, sdfsfilename string, memberlist *[]MemberID, consistency int)
// 	/*
// 	   add two more parameters, targetaddress, memberlist, consistency=4
// 	*/
// 	// modeldata target & memberlist
// 	consistency := 2
// 	target := "localhost:27000"
// 	var memberList []MemberID
// 	p := new(MemberID)
// 	p.LocalIP = "localhost:27001"
// 	p.JoinedTime = time.Now()
// 	p1 := new(MemberID)
// 	p1.LocalIP = "localhost:27002"
// 	p1.JoinedTime = time.Now()
// 	p2 := new(MemberID)
// 	p2.LocalIP = "localhost:27003"
// 	p2.JoinedTime = time.Now()
// 	*memberList = append(*memberList, p)
// 	*memberList = append(*memberList, p1)
// 	*memberList = append(*memberList, p2)
// 	// find
// 	for i, Member := range *memberList {
// 		if (&Member.LocalIP).Equal(&(message.Member)) {
// 			*memberList = append((*memberList)[:i], (*memberList)[i+1:]...)
// 			fmt.Println("A member voluntarily left the group:")
// 			fmt.Println(message.Member)
// 			fmt.Println()

// 			logMsg, _ := json.Marshal(message.Member)
// 			mp2util.WriteLog("Delete: Member voluntarily left the group:" + string(logMsg))

// 			for i := 0; i < 2; i++ {
// 				sendFileToClient()
// 			}

// 		}
// 	}
	// if getlocaladdress() == leaderAddress:
	// address = findaddress(sdfsfilename);
	// if address in memberlist:
	// _,e = sendfile(localfilename, sdfsfilname, []address)
	// if e: resend??
	// else connect master:
	// reconnect master if connection fail;
	// if failed over 3 times:
	// reelect leader and reconnecr leader;
	// ask leader for sending address.[]address
	// if address in memberlist:
	// _,e = sendfile(localfilename, sdfsfilname, []address)
	// if e: resend??
}

func main() {
	// initi()
	// isLeader := false
	SendFileTo("localhost:27001", "dummyfile.txt", "receivedfile.txt")
}
