package main

import (
	// "bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
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

var memberList []MemberID

type IP string

func TcpListening() {
	server, err := net.Listen("tcp", "localhost:27001")
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
		bufferReceivedFileName := make([]byte, 65)
		bufferFileSize := make([]byte, 10)

		connection.Read(bufferFileSize)
		fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

		connection.Read(bufferLocalFileName)
		localfileName := strings.Trim(string(bufferLocalFileName), ":")
		fmt.Println("Server: receiving file:")
		fmt.Println("Server: localname -- " + localfileName)

		connection.Read(bufferReceivedFileName)
		receivedfileName := strings.Trim(string(bufferReceivedFileName), ":")
		fmt.Println("Server: receiving file:")
		fmt.Println("Server: " + receivedfileName)

		newFile, err := os.Create(receivedfileName)

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
	}
}

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
}

func main() {
	go TcpListening()
	var m1 = MemberID{LocalIP: "127.0.0.1"}
	memberList = append(memberList, m1)
	RespondIPListening()
	for {

	}

}
