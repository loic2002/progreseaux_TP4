package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)
func main() {
    if len(os.Args) != 3 {
        fmt.Println("Usage: go run client.go <local_address> <peer_address>")
        os.Exit(1)
    }

    localAddress := os.Args[1]
    peerAddress := os.Args[2]

    // Start listening for incoming connections.
    go listenForConnections(localAddress)

    // Connect to the specified peer.
    connectToPeer(peerAddress)

    // Keep the program running.
    select {}
}
func handleConnection(conn net.Conn) {
    defer conn.Close()

    remoteAddress := conn.RemoteAddr().String()
    fmt.Printf("Accepted connection from %s\n", remoteAddress)

    // Read and display messages from the remote peer.
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        message := scanner.Text()
        fmt.Printf("Received from %s: %s\n", remoteAddress, message)
    }
}

func connectToPeer(peerAddress string) {
    conn, err := net.Dial("tcp", peerAddress)
    if err != nil {
        fmt.Println("Failed to connect to peer:", err)
        return
    }
    defer conn.Close()

    fmt.Println("Connected to peer at", peerAddress)

    // Read and send messages to the peer.
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        message := scanner.Text()
        _, err := conn.Write([]byte(message + "\n"))
        if err != nil {
            fmt.Println("Failed to send message to peer:", err)
            return
        }
        // Read the response from the peer.
        scannerPeer := bufio.NewScanner(conn)
        scannerPeer.Scan()
        response := scannerPeer.Text()
        fmt.Printf("Received response from %s: %s\n", peerAddress, response)
    }
}
func listenForConnections(localAddress string) {
    listener, err := net.Listen("tcp", localAddress)
    if err != nil {
        fmt.Println("Failed to listen:", err)
        os.Exit(1)
    }
    defer listener.Close()

    fmt.Printf("Listening on %s\n", localAddress)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Failed to accept connection:", err)
            continue
        }

        go handleConnection(conn)
    }
}
