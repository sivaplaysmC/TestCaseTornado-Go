package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

var goTunnelLogger = log.New(os.Stdout, "[GoTunnel] ", log.Ltime|log.Ldate|log.Lshortfile)
var msgLogger = log.New(os.Stdout, "[New Incoming Message] ", log.Ltime)

const subdomain string = "hidden-testcases-here"

type TunnelInfo struct {
	Subdomain    string `json:"id,omitempty"`
	Port         int    `json:"port,omitempty"`
	MaxConnCount int    `json:"max_conn_count,omitempty"`
	Url          string `json:"url,omitempty"`
}

func main() {

	handler := http.NewServeMux()
	count := 0
	handler.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi There Nigga %v\n", count)
		msgLogger.Printf("Got a %v request from nigga %v\n", r.Method, count)
		count++
	})

	handler.HandleFunc("/api/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseMultipartForm(1024 * 8)
			name := r.FormValue("name")
			file, _, _ := r.FormFile("file")

			bytes, _ := io.ReadAll(file)
			msgLogger.Println("File from: ", name, "\nSTART FILE-------------\n", string(bytes), "\n-------------END FILE")
		}
	})

	endPoint, err := net.Listen("tcp", "localhost:8080")
	fatalErr(err)
	server := http.Server{
		Handler: handler,
	}
	go server.Serve(endPoint)

	resp, err := http.Get("http://localtunnel.me/" + subdomain)
	checkErr(err)
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	checkErr(err)
	info := TunnelInfo{}
	json.Unmarshal(bytes, &info)
	checkErr(err)
	goTunnelLogger.Println(info.Url)

	for {
		remoteConn, err := net.Dial("tcp4", fmt.Sprintf("localtunnel.me:%d", info.Port))
		fatalErr(err)
		localConn, err := net.Dial("tcp4", fmt.Sprintf("localhost:%d", 8080))
		fatalErr(err)

		go func() {
			io.Copy(remoteConn, localConn)
		}()
		go func() {
			io.Copy(localConn, remoteConn)
		}()
	}

}

func checkErr(error error) {
	if error != nil {
		goTunnelLogger.Println(error.Error())
	}
}

func fatalErr(error error) {
	if error != nil {
		goTunnelLogger.Println(error.Error())
		os.Exit(1)
	}
}
