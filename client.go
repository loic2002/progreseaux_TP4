package main

import (
    "fmt"
    "net/http"
    "net"
    "io"
    "encoding/json"
    "os"
    "bufio"
    "bytes"
    "strings"

)
type setPayload struct {
	Key   string
	Value string
}
type setPayloadData struct {
    Data []string `json:"data"`

}
type setPayloadDat struct {
    Data string `json:"data"`

}

func main() {

    go listenForConnections("127.0.0.1:666")

    // Create http server port 80
    
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // send a request to server port localhost:8989
        resp, err := http.Get("http://localhost:8989/getall")
        defer resp.Body.Close()
       if err != nil {
           fmt.Println("Error:", err)
              return
         }
            body, err := io.ReadAll(resp.Body)
            if err != nil {
                fmt.Println("Error:", err)
                return
            }
            
            // decode json body and check if data == null
	        b := setPayloadData{}
	        err = json.Unmarshal(body, &b)

            if err != nil {
                fmt.Println("Error:", err)
                return
            }

            if b.Data == nil {
                fmt.Println("Data is null")
                return
            }


            // Create a list of data \n
            var listData string
            for _, data := range b.Data {
                listData += "<a href='/get/"+ data +"'>" + data + "</> <br>"
            }
            fmt.Fprintf(w, listData)
            
    })
    http.HandleFunc("/get/", func(w http.ResponseWriter, r *http.Request) {
        filename := r.URL.Path[len("/get/"):]

        data := connectToPeer("127.0.0.1:1000", "get " + filename)

        // remove Found it! from data
        data = data[9:]

        // remove the last \n from data
        data = strings.TrimSuffix(data, "\\n")

        // send data to client
        fmt.Fprintf(w, data)
    })

    http.HandleFunc("/add/", func(w http.ResponseWriter, r *http.Request) {

        filename := r.URL.Path[len("/add/"):]

        // Check if file exist in dht server
        resp, err := http.Get("http://localhost:8989/getall")
        defer resp.Body.Close()
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
        
        // decode json body and check if data == null
        b := setPayloadData{}
        err = json.Unmarshal(body, &b)

        if err != nil {
            fmt.Println("Error:", err)
            return
        }

        // check if file exist in dht server
        for _, data := range b.Data {
            if data == filename {
                fmt.Fprintf(w, "File already exist")
                return
            }
        }

        data := r.URL.Query().Get("data")

        // add to dht server curl -X POST 'localhost:8989/add' -d '{"value": "Bonjour2"}' -H 'content-type: application/json'
        resp, err = http.Post("http://localhost:8989/add", "application/json", bytes.NewBuffer([]byte(`{"value": "` + filename + `"}`)))
        defer resp.Body.Close()

        if err != nil {
            fmt.Println("Error:", err)
            return
        }

        // send to peer
        connectToPeer("127.0.0.1:1000", "add " + filename + " " + data)

    })

    http.ListenAndServe(":80", nil)

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

func connectToPeer(peerAddress string, commands string) (string) {
    conn, err := net.Dial("tcp", peerAddress)
    if err != nil {
        fmt.Println("Failed to connect to peer:", err)
        return ""
    }
    defer conn.Close()

    fmt.Println("Connected to peer at", peerAddress)

    // Read and send messages to the peer.
    _, erra := conn.Write([]byte(commands + "\n"))
    if erra != nil {
        fmt.Println("Failed to send message to peer:", err)
        return ""
    }
    // Read the response from the peer.
    scannerPeer := bufio.NewScanner(conn)
    scannerPeer.Scan()
    response := scannerPeer.Text()
    fmt.Println("Received response from %s: %s\n", peerAddress, response)

    return response
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