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

            // Create a list of data \n
            var listData string
            for _, data := range b.Data {
                listData += "<a href='/get/"+ data +"'>" + data + "</> <br>"
            }

            // Create a form to add data

            tmpl := `<html>
            <head>
            <title>Home</title>
            </head>
            <body>
            <form action="/add" method="post" enctype="multipart/form-data">
            <input type="file" name="myfile" id="myfile">
            <input type="submit" value="Upload">
            </form>
            ` + listData + `
            </body>
            </html>`
            fmt.Fprintf(w, tmpl)

    })
    http.HandleFunc("/get/", func(w http.ResponseWriter, r *http.Request) {
        filename := r.URL.Path[len("/get/"):]

        data := connectToPeer("127.0.0.1:1000", "get " + filename)

        // remove Found it! from data
        data = data[9:]

        // Remove first 
        data = data[1:]

        
        // Remove xENDx from data
        data = strings.ReplaceAll(data, "xENDx", "")
        
        // remove the last \n from data
        data = data[:len(data)-3]
        

        // Check if is jpg
        if strings.Contains(filename, "jpg") {
            w.Header().Set("Content-Type", "image/jpeg")
            w.Write([]byte(data))
            return
        }
        
        // when \n create new line
        data = strings.ReplaceAll(data, "\\n", "<br>")

        split := strings.Split(data, " ")
        // remove 2 first
        data = strings.Join(split[2:], " ")


        // 
        tmpl := `<html>
        <head>
        <title>`+ filename +`</title>
        </head>
        <body>
        <p>` + data + `</p>
        </body>
        </html>`
        fmt.Fprintf(w, tmpl)
    })
    
    http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {

        // get 
        r.ParseMultipartForm(32 << 20)
        file, handler, err := r.FormFile("myfile")
        if err != nil {
            fmt.Println(err)
            return
        }
        // get filename
        filename := handler.Filename

        // Check if file exist in RAFT
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

        // create a buffer to store the data
        var buf bytes.Buffer
        io.Copy(&buf, file)

        // convert buffer to string
        data := buf.String()


        fmt.Println("filename:", filename)
        fmt.Println("data:", data)


        //add to RAFT server curl -X POST 'localhost:8989/add' -d '{"value": "Bonjour2"}' -H 'content-type: application/json'
        resp, err = http.Post("http://localhost:8989/add", "application/json", bytes.NewBuffer([]byte(`{"value": "` + filename + `"}`)))
        defer resp.Body.Close()

        if err != nil {
            fmt.Println("Error:", err)
            return
        }

        // send data to server port 1000
        data = connectToPeer("127.0.0.1:1000", "add " + filename + " " + data)

        // send to RAFT

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
    _, erra := conn.Write([]byte(commands + " xENDx" + "\n"))
    if erra != nil {
        fmt.Println("Failed to send message to peer:", err)
        return ""
    }
    // Read the response from the peer.
    scannerPeer := bufio.NewScanner(conn)
    response := ""
    for scannerPeer.Scan() {
        response += scannerPeer.Text() + "\n"
        if strings.Contains(scannerPeer.Text(),"xENDx") {
            break
        }
    }

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