package main

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"github.com/KonstantinGalanin/http_proxy/internal/proxy"
	"github.com/KonstantinGalanin/http_proxy/internal/repository"
)

type ProxyServer struct {
	repository *repository.PostgresRepo
}

var (
	dbHost = os.Getenv("DATABASE_HOST")
	dbPort = os.Getenv("DATABASE_PORT")
	dbUser = os.Getenv("DATABASE_USER")
	dbPass = os.Getenv("DATABASE_PASSWORD")
	dbName = os.Getenv("DATABASE_NAME")

	serverPort = os.Getenv("SERVER_PORT")
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

func (p *ProxyServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	modifyRequest(r)

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


	bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = w.Write(bodyBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	parsedResp := &proxy.ParsedResponse{
        StatusCode:    resp.StatusCode,
        Status:        resp.Status,
        Headers:       resp.Header,
        Body:          string(bodyBytes),
        ContentLength: resp.ContentLength,
        Compressed:    resp.Header.Get("Content-Encoding") == "gzip",
    }
    
    err = p.repository.SaveResponse(parsedResp)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
}

func (p *ProxyServer) handleHTTPS(w http.ResponseWriter, r *http.Request) {
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
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	repository := repository.New(db)
	proxyServer := &ProxyServer{
		repository: repository,
	}

	server := &http.Server{
		Addr: ":8081",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme := "http"
			if r.Method == http.MethodConnect {
				proxyServer.handleHTTPS(w, r)
				scheme = "https"
			} else {
				proxyServer.handleHTTP(w, r)
			}

			parsedReq, err := proxy.ParseRequest(r, scheme)
			if err != nil {
				log.Println("failed parse request: %w", err)
			}
			err = repository.SaveRequest(parsedReq)
			if err != nil {
				log.Println("failed save request: %w", err)
			}
		}),
	}

	log.Fatal(server.ListenAndServe())

}
