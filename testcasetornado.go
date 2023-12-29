package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var goTunnelLogger = log.New(os.Stdout, "[GoTunnel] ", log.Ltime|log.Ldate)
var msgLogger = log.New(os.Stdout, "[New Incoming Message] ", log.Ltime|log.Ldate|log.LstdFlags)

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
		fmt.Fprintf(w, "Hi There %v\n", count)
		msgLogger.Printf("Got a %v request from %v\n", r.Method, count)
		count++
	})

	handler.HandleFunc("/api/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var output strings.Builder

			name := r.FormValue("name")
			if name != "" {
				output.WriteString("From:: " + name + " ")
			}

			meta := r.FormValue("meta")
			if meta != "" {
				output.WriteString("MetaData:: " + meta + " ")
			}

			file, fileHeader, err := r.FormFile("test_case")
			if err == nil {
				defer file.Close()
				bytes, _ := io.ReadAll(file)
				output.WriteString("\nFile Name:: " + fileHeader.Filename)
				output.WriteString("\nSTART TESTCASE───────────────────────────────────────┐\n" +
					string(bytes) +
					"\n└───────────────────────────────────────END TESTCASE\n")
			}

			msgLogger.Print(output.String())

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
	goTunnelLogger.Println("Tunnel URL :: " + info.Url)

	// hack to tunnel localhost to a .loca.lt domain
	fun := func() {
		for {
			remoteConn, err := net.Dial("tcp4", fmt.Sprintf("localtunnel.me:%d", info.Port))
			fatalErr(err)
			localConn, err := net.Dial("tcp4", fmt.Sprintf("localhost:%d", 8080))
			fatalErr(err)

			go io.Copy(remoteConn, localConn)
			go io.Copy(localConn, remoteConn)
		}
	}

	for i := 0; i < 9; i++ {
		go fun()
	}

	fun()

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
