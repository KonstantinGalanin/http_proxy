package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func modifyRequest(r *http.Request) {
	host := r.URL.Host
	if host == "" {
		host = r.Host
	}

	r.URL.Scheme = ""
	r.URL.Host = ""

	r.Host = host
	r.Header.Set("Host", host)

	r.Header.Del("Proxy-Connection")
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	modifyRequest(r)

	fmt.Println(r)

	target := r.URL.Host
	if target == "" {
		target = r.Host
	}

	if !strings.Contains(target, ":") {
		target += ":80"
	}

	conn, err := net.Dial("tcp", target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	err = r.Write(conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	fmt.Fprint(conn, "HTTP/1.1 200 Connection established\r\n\r\n")

	cert, err := tls.LoadX509KeyPair("mail.ru.crt", "mail.ru.key")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	tlsConn := tls.Server(conn, tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	remoteConn, err := tls.Dial("tcp", r.URL.Host, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer remoteConn.Close()

	go transfer(remoteConn, tlsConn)
	transfer(tlsConn, remoteConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func main() {
	server := &http.Server{
		Addr: ":8081",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleHTTPS(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
	}

	log.Fatal(server.ListenAndServe())

}
