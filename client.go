package main

import (
	// "fmt"

	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const BUFFER_SIZE_C = 32 * 1024

func startClient(ip string, uplink string, downlink string, tolink string, token string) {

	listener, err := net.Listen("tcp", ip)
	if err != nil {
		log.Println("Fail to start client: ", err)
		return
	}
	defer listener.Close()

	log.Println("Client started successfully at " + ip)

	for {
		myid := uuid.New().String()[28:]

		conn, err := listener.Accept()
		if err != nil {
			log.Println("Fail to initiate the connection", err)
			continue
		}
		go handleClientUp(conn, uplink, myid, tolink, token)
		go handleClientDown(conn, downlink, myid, token)
	}
}

func handleClientUp(conn net.Conn, uplink string, myid string, tolink string, token string) {

	log.Println("Start up request: ", myid)

	// // Create a custom io.Reader that simulates streaming data
	// // This reader will provide data in chunks, and since its length is unknown beforehand,
	// // the Go HTTP client will automatically use chunked encoding.
	// bodyReader := &chunkedReader{
	// 	conn:     conn,
	// 	buffSize: BUFFER_SIZE_C,
	// }

	// Create a new POST request
	req, err := http.NewRequest("POST", uplink+"?id="+myid+"&to="+tolink, conn)
	if err != nil {
		log.Println("Error creating up request:", err)
		return
	}

	// Set ContentLength to -1 to explicitly indicate unknown length and force chunked encoding
	// (though often not strictly necessary if the bodyReader doesn't provide a length)
	req.ContentLength = -1
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Connection", "close")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending up request: ", err)
		return
	}
	defer resp.Body.Close()

	// Read and print the response
	// responseBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Println("Error reading response body: ", err)
	// }
	// log.Println("Server Response: ", string(responseBody))

	io.ReadAll(resp.Body)

	log.Println("End up request: ", myid)
}

func handleClientDown(conn net.Conn, downlink string, myid string, token string) {
	log.Println("Start download request: ", myid)
	time.Sleep(500 * time.Millisecond)
	// Create a new POST request
	req, err := http.NewRequest("GET", downlink+"?id="+myid, nil)
	if err != nil {
		log.Println("Error creating down request: ", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Connection", "close")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if err.Error() != "EOF" {
			log.Println("Error sending down request: ", err)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		buf := make([]byte, BUFFER_SIZE_C)

		for {
			n, err := resp.Body.Read(buf)

			if n > 0 {
				conn.Write(buf[:n])
			}

			if err != nil {
				if err.Error() != "EOF" {
					log.Println("Error forwarding down response err: ", err)
				}
				break
			}
		}
	} else {
		log.Println("Down Status Code: ", resp.Status)
	}

	conn.Close()

	log.Println("End down request: ", myid)
}

// // chunkedReader simulates an io.Reader that provides data in chunks
// type chunkedReader struct {
// 	conn     net.Conn
// 	buffSize int
// }

// func (cr *chunkedReader) Read(p []byte) (n int, err error) {
// 	bsize := len(p)
// 	if bsize > cr.buffSize {
// 		bsize = cr.buffSize
// 	}

// 	buff := make([]byte, bsize)

// 	n, err = cr.conn.Read(buff)

// 	fmt.Println("Received from TCP:", n)

// 	if n > 0 {
// 		copy(p, buff[:n])
// 	}

// 	if err != nil {
// 		fmt.Println("Fail to receive from TCP", err)
// 	}

// 	return n, err
// }
