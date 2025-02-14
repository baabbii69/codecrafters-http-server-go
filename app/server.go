package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// reading input
	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}
	// spliting the request by spaces
	parts := strings.Fields(request)
	if len(parts) < 3 {
		fmt.Println("Invalid reques line")
		return
	}

	method := parts[0]  // GET
	path := parts[1]    // index.html
	version := parts[2] // HTTP/1.1

	fmt.Printf("Request: %s %s %s\n", method, path, version)

	var response string
	if path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(path, "/echo/") {
		echoStr := strings.TrimPrefix(path, "/echo/")
		contentLength := len(echoStr)
		response = fmt.Sprintf("%s 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", version, contentLength, echoStr)

	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		return
	}
}




	// handling /echo/{abc} requests
	// if strings.HasPrefix(path, "/echo/") {
	// 	// Extracting the string after /echo/
	// 	echoStr := strings.TrimPrefix(path, "/echo/")

	// 	contentLength := len(echoStr)
	// 	response := fmt.Sprintf("%s 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", version, contentLength, echoStr)

	// 	_, err = conn.Write([]byte(response))
	// 	if err != nil {
	// 		fmt.Println("Error writing response: ", err.Error())
	// 		return
	// 	}
	// 	return
	// }