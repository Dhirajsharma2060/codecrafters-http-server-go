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
		go handleConnection(conn)
		//defer is the keyword which is used to ensure that the connection is closed just before the function returns
		// defer conn.Close()
	}

}

func handleConnection(conn net.Conn) {
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

		if len(parts) >= 2 {
			// GET / HTTP/1.1\r\n
			// Host: localhost\r\n
			// \r\n
			// here we then split the first line of line [0] by spaces this will open the resquest and spearets the GET , / and HTTP/1.1 protocol version

			path := parts[1]
			//if path is == "/" then we set the path to './index.html'
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
				bodyContent, err := os.ReadFile(strings.TrimPrefix(path, "/files/"))
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
		}
	}

}
