package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

var baseDirectory string // to store --directory flag value

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

	// reading headers
	header := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading header: ", err.Error())
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // end of header
		}
		// splitting the header by ":" into key and value
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.ToLower(strings.TrimSpace(headerParts[0])) // making the key lower case causes it is case insensitive
			value := strings.TrimSpace(headerParts[1])
			header[key] = value
		}
	}

	var response string
	if path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(path, "/echo/") {
		echoStr := strings.TrimPrefix(path, "/echo/")

		response = handleResponse(version, "200 OK", "text/plain", echoStr)
		// contentLength := len(echoStr)
		// response = fmt.Sprintf("%s 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", version, contentLength, echoStr)

	} else if path == "/user-agent" {
		userAgent, exists := header["user-agent"]
		if !exists {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
			return
		}

		response = handleResponse(version, "200 OK", "text/plain", userAgent)
		// contentLength := len(userAgent)
		// response = fmt.Sprintf("%s 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", version, contentLength, userAgent)
	} else if strings.HasPrefix(path, "/files/") { // parsing the --directory flag
		args := os.Args[1:]
		if len(args) < 2 || args[0] != "--directory" {
			fmt.Println("Usage: ./your_program.sh --directory <absolute-path>")
			os.Exit(1)
		}
		baseDirectory = args[1]
		// extracting the file name after /files/
		fileName := strings.TrimPrefix(path, "/files/")
		filePath := filepath.Join(baseDirectory, fileName)

		// checking if the file exists
		fileInfo, err := os.Stat(filePath)
		if err != nil || fileInfo.IsDir() { // file does not exist or it is a directory
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
			return
		}

		// reading the file content
		file, err := os.ReadFile(filePath)
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
			return
		}
		response = handleResponse(version, "200 OK", "application/octet-stream", string(file))

	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
		return
	}
}

func handleResponse(version string, status string, contentType string, body string) string {
	return fmt.Sprintf("%s %s\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", version, status, contentType, len(body), body)
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
