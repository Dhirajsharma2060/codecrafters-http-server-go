package main

import (
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

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error while waiting for and accepting a connection: ", err.Error())
		os.Exit(1)
	}
	//defer is the keyword which is used to ensure that the connection is closed just before the function returns
	defer conn.Close()

	// This reads the request from the client and stores it in a buffer
	// The `make` function is used to create a byte slice of size 1024 bytes (1 KB).
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err.Error())
		os.Exit(1)
	}
	request := string(buf)
	lines := strings.Split(request, "\r\n")
	if len(lines) > 0 {
		parts := strings.Split(lines[0], " ")
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

			} else {
				response := "HTTP/1.1 404 Not Found\r\n\r\n"
				// If the path is not '/', a 404 response is sent
				conn.Write([]byte(response))
			}
		}
	}
	// You can add more logic here to check the request and set response accordingly

}
