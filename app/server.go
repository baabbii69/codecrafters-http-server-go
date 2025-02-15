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

const (
	StatusOK       = "200 OK"
	StatusCreated  = "201 Created"
	StatusNotFound = "404 Not Found"
	StatusError    = "500 Internal Server Error"
)

func main() {
	// parsing the --directory flag
	args := os.Args[1:]
	if len(args) >= 2 && args[0] == "--directory" {
		baseDirectory = args[1]
		fmt.Println("Using directory:", baseDirectory)

		// checking if the directory exists
		fileInfo, err := os.Stat(baseDirectory)
		if err != nil || fileInfo.IsDir() { // file does not exist or it is a directory
			os.Exit(1)
		}
	}
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
	contentLength := 0
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
			if key == "content-length" {
				fmt.Sscanf(value, "%d", &contentLength)
			}
		}
	}

	var response string
	if method == "GET" && path == "/" {
		response = handleResponse(version, StatusOK, "text/plain", "")
	} else if method == "GET" && strings.HasPrefix(path, "/echo/") {
		echoStr := strings.TrimPrefix(path, "/echo/")

		response = handleResponse(version, StatusOK, "text/plain", echoStr)
		// contentLength := len(echoStr)
		// response = fmt.Sprintf("%s 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", version, contentLength, echoStr)

	} else if method == "GET" && path == "/user-agent" {
		userAgent, exists := header["user-agent"]
		if !exists {
			response = handleResponse(version, StatusNotFound, "text/plain", "")
			return
		}

		response = handleResponse(version, StatusOK, "text/plain", userAgent)
		// contentLength := len(userAgent)
		// response = fmt.Sprintf("%s 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", version, contentLength, userAgent)
	} else if method == "GET" && strings.HasPrefix(path, "/files/") {
		if baseDirectory == "" {
			response = handleResponse(version, StatusNotFound, "text/plain", "")
			// _, _ = conn.Write([]byte(response))
			return
		}
		// extracting the file name after /files/
		fileName := strings.TrimPrefix(path, "/files/")
		filePath := filepath.Join(baseDirectory, fileName)

		// checking if the file exists
		fileInfo, err := os.Stat(filePath)
		if err != nil || fileInfo.IsDir() { // file does not exist or it is a directory
			response = handleResponse(version, StatusNotFound, "text/plain", "")
			// _, _ = conn.Write([]byte(response))
			return
		}

		// reading the file content
		file, err := os.ReadFile(filePath)
		if err != nil {
			response = handleResponse(version, StatusError, "text/plain", "")
			_, _ = conn.Write([]byte(response))
			return
		}
		response = handleResponse(version, StatusOK, "application/octet-stream", string(file))

	} else if method == "POST" && strings.HasPrefix(path, "/files/") {
		if baseDirectory == "" {
			response = handleResponse(version, StatusNotFound, "text/plain", "")
			// _, _ = conn.Write([]byte(response))
			return
		}
		fileName := strings.TrimPrefix(path, "/files/")
		filePath := filepath.Join(baseDirectory, fileName)

		body := make([]byte, contentLength)
		_, err := reader.Read(body)
		if err != nil {
			response = handleResponse(version, StatusError, "text/plain", "")
			return
		}
		err = os.WriteFile(filePath, body, 0644)
		if err != nil {
			response = handleResponse(version, StatusError, "text/plain", "")
			return
		}
		response = handleResponse(version, StatusCreated, "text/plain", "")
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
