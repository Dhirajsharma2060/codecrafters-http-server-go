package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = os.Exit

func main() {

	//Parse --directory flag
	directory := "."
	args := os.Args
	for i, arg := range args {
		if arg == "--directory" && i+1 < len(args) {
			directory = args[i+1]
			// If the --directory flag is provided, we set the directory variable to the next argument
			// This allows us to change the directory where the server will look for files
		}
	}
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	// The net.Listen function is used to create a TCP listener on port 4221
	fmt.Println("Listening on port 4221...")
	// The Accept method waits for and accepts a connection on the listener
	// The Accept method returns a net.Conn object which represents the connection
	// The net.Conn object can be used to read from and write to the connection
	// The for loop is used to continuously accept connections
	// The handleConnection function is called in a goroutine to handle each connection concurrently
	// The handleConnection function is defined below to handle the connection

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error while waiting for and accepting a connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, directory)
		//defer is the keyword which is used to ensure that the connection is closed just before the function returns
		// defer conn.Close()
	}

}

func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err.Error())
		os.Exit(1)
	}
	request := string(buf)
	lines := strings.Split(request, "\r\n")
	if len(lines) > 0 {
		parts := strings.Split(lines[0], " ")

		userAgent := ""
		for _, line := range lines {
			if strings.HasPrefix(line, "User-Agent: ") {
				// If the line starts with "User-Agent: ", we extract the user agent string
				userAgent = strings.TrimPrefix(line, "User-Agent: ")
				// TrimPrefix removes the "User-Agent: " part from the line
				// and leaves us with just the user agent string
				break
			}
		}

		if len(parts) >= 3 {
			// GET / HTTP/1.1\r\n
			// Host: localhost\r\n
			// \r\n
			// here we then split the first line of line [0] by spaces this will open the resquest and spearets the GET , / and HTTP/1.1 protocol version
			method := parts[0]
			path := parts[1]
			//if path is == "/" then we set the path to './index.html'
			if method == "GET" {

				if path == "/" {
					response := "HTTP/1.1 200 OK\r\n\r\n"
					conn.Write([]byte(response))

				} else if strings.HasPrefix(path, "/echo/") {
					// If the path starts with /echo/, we extract the message
					echoStr := path[len("/echo/"):]
					//GET /echo/abc HTTP/1.1\r\nHost: localhost\r\n\r\n
					//the HasPrefix function checks if the path starts with "/echo/"
					// then we extract the message by slicing the path from the length of "/echo/" to the end
					// in above example it will be abc
					// We then create a response with the message
					body := echoStr
					response := "HTTP/1.1 200 OK\r\n" +
						"Content-Type: text/plain\r\n" +
						fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
						"\r\n" +
						body
					conn.Write([]byte(response))
				} else if strings.HasPrefix(path, "/user-agent") {
					body := userAgent
					response := "HTTP/1.1 200 OK\r\n" +
						"Content-Type:text/plain\r\n" +
						fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
						"\r\n" +
						body
					conn.Write([]byte(response))
				} else if strings.HasPrefix(path, "/files/") {
					filepath := directory + "/" + strings.TrimPrefix(path, "/files/")
					bodyContent, err := os.ReadFile(filepath)
					if err != nil {
						// If the file does not exist, we send a 404 response
						response := "HTTP/1.1 404 Not Found\r\n\r\n"
						conn.Write([]byte(response))
						return
					}
					body := string(bodyContent)
					response := "HTTP/1.1 200 OK\r\n" +
						"Content-Type: application/octet-stream\r\n" +
						fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
						"\r\n" +
						body
					conn.Write([]byte(response))

				} else {
					response := "HTTP/1.1 404 Not Found\r\n\r\n"
					// If the path is not '/', a 404 response is sent
					conn.Write([]byte(response))
				}
			} else if method == "POST" {
				if strings.HasPrefix(path, "/files/") {

					filepath := directory + "/" + strings.TrimPrefix(path, "/files/")
					// If the path starts with /files/, we extract the file path
					// We use strings.TrimPrefix to remove the "/files/" part from the path
					// and concatenate it with the directory to get the full file path
					// We then read the request body to get the content to write to the file

					bodyStartIndex := -1
					for i, line := range lines {
						if line == "" {
							bodyStartIndex = i + 1
							// We find the index of the first empty line, which indicates the start of the body
							// The body starts after the headers, so we set bodyStartIndex to the next line
							break
						}
					}
					// We loop through the lines of the request to find the first empty line
					// If we find an empty line, we set bodyStartIndex to the next line
					// If we don't find an empty line, bodyStartIndex will remain -1
					// If bodyStartIndex is -1 or greater than or equal to the length of lines,
					// it means we didn't find an empty line, which indicates that the body is missing
					if bodyStartIndex == -1 || bodyStartIndex >= len(lines) {
						// If we didn't find an empty line, we return a 400 Bad Request response
						response := "HTTP/1.1 400 Bad Request\r\n\r\n"
						conn.Write([]byte(response))
						return
					}
					body := strings.Join(lines[bodyStartIndex:], "\r\n")
					// We join the lines after the empty line to get the body content
					err := os.WriteFile(filepath, []byte(body), 0644)
					if err != nil {
						// If there is an error writing the file, we send a 500 Internal Server Error response
						response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
						conn.Write([]byte(response))
						return
					}
					// If the file is written successfully, we send a 200 OK response
					response := "HTTP/1.1 201 OK\r\n\r\n"
					conn.Write([]byte(response))
				} else {
					// If the path does not start with /files/, we send a 404 Not Found response
					response := "HTTP/1.1 404 Not Found\r\n\r\n"
					conn.Write([]byte(response))
				}

			}
		}
	}

}
