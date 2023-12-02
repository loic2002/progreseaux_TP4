package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
    "io/ioutil"
)

const (
    configFolder = "Config"
    servers = "servers.lst"
    files = "files.lst"
    data = "./data/"
)

func getRoutingNextHop(nodeId string, chars string) (string){

    // Open file servers and read line by line
    file, err := os.Open(configFolder +"/"+ nodeId +"/"+ servers)
    if err != nil {
        fmt.Println("Error:", err)
        return ""
    }

    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        // Get the line
        line := scanner.Text()

        // Detect the version of the file v1 (only one line with ip '192.168.1.1:23') or V2 (multiple lines with the start by A-C 192.168.1.1:22 ')
        if strings.Contains(line, "-") {

            //Split the line by the space
            split := strings.Split(line, " ")

            if checkRange(chars, split[0]) {
                return split[1]
            }

        }else{
            return line
        }
    }
    return ""
}

func getCharRange(nodeId string) (string){
    
        // Open file servers and read line by line
        file, err := os.Open(configFolder +"/"+ nodeId +"/"+ files)
        if err != nil {
            fmt.Println("Error:", err)
            return ""
        }
    
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            // Get the line
            line := scanner.Text()
    
            return line
        }
        return ""
}

func WriteFile(nodeId string, data string) {
    // Open file servers and read line by line
    file, err := os.OpenFile(configFolder +"/"+ nodeId +"/"+ files, os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    defer file.Close()

    if _, err := file.WriteString(data); err != nil {
        fmt.Println("Error:", err)
        return
    }
}

func WriteServer(nodeId string, data string) {
    // Open file servers and read line by line
    file, err := os.OpenFile(configFolder +"/"+ nodeId +"/"+ servers, os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    defer file.Close()

    if _, err := file.WriteString(data ); err != nil {
        fmt.Println("Error:", err)
        return
    }
}

func main() {
    
    localAddress := ""
    chars := ""
    nodeId := ""

    if len(os.Args) != 3 && len(os.Args) != 5 {
        fmt.Println("Usage: go run server.go <local_address> <nodeId>")
        fmt.Println("Usage: go run server.go <local_address> <char_range:A-B> <peer_address> <nodeId>")
        os.Exit(1)
    }
    
    // Read the local peer's address from the command line.
    if len(os.Args) == 5 {
    
        localAddress = os.Args[1]
        chars = os.Args[2]
        nodeId = os.Args[4]
        
        createDataFolder(nodeId)
        createConfigFolder(nodeId)
        
        WriteServer(os.Args[4], os.Args[3])
        WriteFile(os.Args[4], os.Args[2])
    }

    if len(os.Args) == 3 {
        localAddress = os.Args[1]
        nodeId = os.Args[2]

        createDataFolder(nodeId)
        createConfigFolder(nodeId)

        chars = getCharRange(nodeId)
    }


    // Listen for incoming connections.
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

        go handleConnection(conn, chars, nodeId)
    }
    
}

func handleConnection(conn net.Conn, chars string, nodeId string) {
    defer conn.Close()

    remoteAddress := conn.RemoteAddr().String()
    fmt.Printf("Accepted connection from %s\n", remoteAddress)

    
    // Read and display messages from the remote peer.
    scanner := bufio.NewScanner(conn)
    message := ""
    isForward := false
    response := ""
    for scanner.Scan() {
        message += scanner.Text() + "\n"
        if strings.Contains(scanner.Text(),"xENDx") {
                split := strings.Split(message, " ")
                if checkRange(split[1],chars) && !isForward {
                    response = "Found it!"

                //  message contain add, del, get
                if strings.Contains(message, "add") {
                    fmt.Println("Add")
                    // Add the file to the server
                    file, err := os.OpenFile(data + nodeId +"/"+split[1], os.O_WRONLY|os.O_CREATE, 0644)
                    if err != nil {
                        fmt.Println("Error:", err)
                        return
                    }

                    defer file.Close()

                    if _, err := file.WriteString(strings.Join(split[2:], " ")); err != nil {
                        fmt.Println("Error:", err)
                        return
                    }

                } else if strings.Contains(message, "del") {
                    fmt.Println("Del")
                    // Delete the file from the server
                    
                    err := os.Remove(data + nodeId +"/"+split[1])
                    if err != nil {
                        fmt.Println("Error:", err)
                        return
                    }
                    


                } else if strings.Contains(message, "get") {
                    fmt.Println("Get")
                    // Get the file from the server
                    fileByte, err := ioutil.ReadFile(data + nodeId +"/"+split[1])
                    if err != nil {
                        fmt.Println("Error:", err)
                        return
                    }
                    response  += " " + string(fileByte)

                } else {
                    fmt.Println("Not a valid command")
                }

                fmt.Println("Send response to peer:", response)
                _, err := conn.Write([]byte(response + " xENDx " + "\n"))
                if err != nil {
                    fmt.Println("Failed to send message to peer:", err)
                    return
                }

            } else {
                fmt.Println("Forward")
                isForward = true

                responsePeer := connectToPeer(getRoutingNextHop(nodeId, chars), message)
                fmt.Println("Response from peer:", responsePeer)
                if strings.Contains(responsePeer, "Found it!") {
                    fmt.Println("Found it!")
                    _, err := conn.Write([]byte(responsePeer + " xENDx " + "\n"))
                    if err != nil {
                        fmt.Println("Failed to send message to peer:", err)
                        return
                    }
                }
            }
        }
        isForward = false
        
    }

}

func connectToPeer(peerAddress string, command string) (string) {
    conn, err := net.Dial("tcp", peerAddress)
    if err != nil {
        fmt.Println("Failed to connect to peer:", err)
        return ""
    }
    defer conn.Close()

    fmt.Println("Connected to peer at", peerAddress)

    // Read and send messages to the peer.
    message := command + " xENDx " + "\n"
    _, err = conn.Write([]byte(message))
    if err != nil {
        fmt.Println("Failed to send message to peer:", err)
        return ""
    }

    // Read the response from the peer.
    scanner := bufio.NewScanner(conn)
    response := ""
    for scanner.Scan() {
        response += scanner.Text() + "\n"
        if strings.Contains(scanner.Text(),"xENDx") {
            break
        }
    }

    fmt.Printf("Received response from %s: %s\n", peerAddress, response)

    return response
}

func createDataFolder(nodeId string){

    // Check if data folder exists
    if _, err := os.Stat(data); os.IsNotExist(err) {
        os.Mkdir(data, 0777)
        fmt.Println("Created data folder")
    }

    // check if folder exists
    if _, err := os.Stat(data + nodeId); os.IsNotExist(err) {
        os.Mkdir(data + nodeId, 0777)
        fmt.Println("Created folder for node", nodeId)
    }
}

// function for dectect the fist letter to string and check if is start by a range A-C
func checkRange(message string, letter string) (bool){
    // Letter = A-D, E-H, I-L, M-P, Q-T, U-X, Y-Z
    // Get ascii value of letter beffore the -
    asciiLower := int(strings.ToLower(letter)[0])
    // Get ascii value of letter after the -
    ascii2Lower := int(strings.ToLower(letter)[2])
    // Get ascii value of letter beffore the -
    asciiUp := int(strings.ToUpper(letter)[0])
    // Get ascii value of letter after the -
    ascii2Up := int(strings.ToUpper(letter)[2])
    // Get ascii value of the first letter of the message
    ascii3 := int(message[0])
    
    // Check if the first letter of the message is in the range
    if ascii3 >= asciiLower && ascii3 <= ascii2Lower || ascii3 >= asciiUp && ascii3 <= ascii2Up {
        fmt.Println("Found it!")
        return true
    } else {
        fmt.Println("Not found")
        return false
    }

}

func createConfigFolder(nodeId string){

    // check if config file exist
    if _, err := os.Stat(configFolder); os.IsNotExist(err) {
        os.Mkdir(configFolder, 0777)
        fmt.Println("Created config folder")
    }
    // Check if nodeId exist
    if _, err := os.Stat(configFolder +"/"+ nodeId); os.IsNotExist(err) {
        os.Mkdir(configFolder +"/"+ nodeId, 0777)
        fmt.Printf("Created %s/%s folder\n",configFolder,nodeId)
    }
    // Check if config servers exist
    if _, err := os.Stat(configFolder +"/"+ nodeId +"/"+ servers); os.IsNotExist(err) {
        os.Create(configFolder +"/"+ nodeId +"/"+ servers)
        fmt.Printf("Created %s/%s/%s file\n",configFolder,nodeId,servers)
    }
        if _, err := os.Stat(configFolder +"/"+ nodeId +"/"+ files); os.IsNotExist(err) {
        os.Create(configFolder +"/"+ nodeId +"/"+ files)
        fmt.Printf("Created %s/%s/%s file\n",configFolder,nodeId,files)
    }
}