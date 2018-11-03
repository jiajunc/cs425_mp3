package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const BUFFERSIZE = 1024

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

func initi() {
	fmt.Println("Enter command to put, get, update or delete file")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		userInput := strings.Split(scanner.Text(), " ")
		fmt.Println(userInput)
		userCommand := userInput[0]
		fmt.Println(userCommand)

		switch userCommand {
		case "put":
			fmt.Println("running put command")
		case "get":
			fmt.Println("rinning get command")
		case "delete":
			fmt.Println("running delete command")
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

func main() {
	TcpListening()
}
