package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Parse --directory flag
	directory := "."
	args := os.Args
	for i, arg := range args {
		if arg == "--directory" && i+1 < len(args) {
			directory = args[i+1]
		}
	}

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Listening on port 4221...")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error while waiting for and accepting a connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, directory)
	}
}

func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()
	buf := make([]byte, 4096) // 🔥 Changed: Increased buffer size to handle larger requests

	for { // 🔥 Changed: Wrap in loop to support persistent connections
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading from connection:", err.Error())
			return
		}
		if n == 0 {
			continue
		}

		request := string(buf[:n])
		lines := strings.Split(request, "\r\n")

		if len(lines) == 0 {
			continue
		}

		parts := strings.Split(lines[0], " ")
		if len(parts) < 3 {
			continue
		}

		method := parts[0]
		path := parts[1]

		// Extract headers
		userAgent := ""
		acceptEncoding := ""
		connectionHeader := "" // 🔥 New: Track connection type

		for _, line := range lines {
			if strings.HasPrefix(line, "User-Agent: ") {
				userAgent = strings.TrimPrefix(line, "User-Agent: ")
			} else if strings.HasPrefix(line, "Accept-Encoding: ") {
				acceptEncoding = strings.TrimPrefix(line, "Accept-Encoding: ")
			} else if strings.HasPrefix(line, "Connection: ") { // 🔥 New
				connectionHeader = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Connection: ")))
			}
		}

		if method == "GET" {
			if path == "/" {
				response := "HTTP/1.1 200 OK\r\n\r\n"
				conn.Write([]byte(response))

			} else if strings.HasPrefix(path, "/echo/") {
				echoStr := path[len("/echo/"):]
				body := echoStr

				if strings.Contains(acceptEncoding, "gzip") {
					var b bytes.Buffer
					gz := gzip.NewWriter(&b)
					gz.Write([]byte(body))
					gz.Close()

					response := "HTTP/1.1 200 OK\r\n" +
						"Content-Type: text/plain\r\n" +
						"Content-Encoding: gzip\r\n" +
						fmt.Sprintf("Content-Length: %d\r\n", b.Len()) +
						"\r\n"
					conn.Write([]byte(response))
					conn.Write(b.Bytes())
				} else {
					headers := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n"
					if connectionHeader == "close" {
						headers += "Connection: close\r\n"
					}
					headers += fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
					response := headers + body
					conn.Write([]byte(response))
				}

			} else if strings.HasPrefix(path, "/user-agent") {
				body := userAgent
				response := "HTTP/1.1 200 OK\r\n" +
					"Content-Type: text/plain\r\n" +
					fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
					"\r\n" +
					body
				conn.Write([]byte(response))

			} else if strings.HasPrefix(path, "/files/") {
				filepath := directory + "/" + strings.TrimPrefix(path, "/files/")
				bodyContent, err := os.ReadFile(filepath)
				if err != nil {
					response := "HTTP/1.1 404 Not Found\r\n\r\n"
					conn.Write([]byte(response))
					continue
				}
				body := string(bodyContent)
				response := "HTTP/1.1 200 OK\r\n" +
					"Content-Type: application/octet-stream\r\n"
				if connectionHeader == "close" {
					response += "Connection: close\r\n"
				}
				response += fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body)) + body
				conn.Write([]byte(response))

			} else {
				response := "HTTP/1.1 404 Not Found\r\n\r\n"
				conn.Write([]byte(response))
			}

		} else if method == "POST" {
			if strings.HasPrefix(path, "/files/") {
				filepath := directory + "/" + strings.TrimPrefix(path, "/files/")

				headerEnd := strings.Index(request, "\r\n\r\n")
				if headerEnd == -1 {
					response := "HTTP/1.1 400 Bad Request\r\n\r\n"
					conn.Write([]byte(response))
					continue
				}

				bodyStart := headerEnd + 4
				body := buf[bodyStart:n]

				err := os.WriteFile(filepath, body, 0644)
				if err != nil {
					response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
					conn.Write([]byte(response))
					continue
				}

				response := "HTTP/1.1 201 Created\r\n\r\n"
				conn.Write([]byte(response))

			} else {
				response := "HTTP/1.1 404 Not Found\r\n\r\n"
				conn.Write([]byte(response))
			}

		} else {
			response := "HTTP/1.1 405 Method Not Allowed\r\n\r\n"
			conn.Write([]byte(response))
		}

		if connectionHeader == "close" { // 🔥 New: Respect Connection: close
			break
		}
	}
}
