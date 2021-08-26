/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package streamserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	ServerPort           = "port"
	ConnectionPath       = "path"
	ConnectionTlsEnabled = "tlsEnabled"
	ConnectionUploadCRT  = "uploadCRT"
	ConnectionTlsCRTPath = "tlsCRTPath"
	ConnectionTlsKeyPath = "tlsKeyPath"
	ConnectionTlsCRT     = "tlsCRT"
	ConnectionTlsKey     = "tlsKey"
)

var (
	instance *StreamServerFactory
	once     sync.Once
)

type StreamRequestListener interface {
	ProcessRequest(request string) error
}

type StreamServerFactory struct {
	sseServers map[string]*Server
	mux        sync.Mutex
}

func GetFactory() *StreamServerFactory {
	once.Do(func() {
		instance = &StreamServerFactory{sseServers: make(map[string]*Server)}
	})
	return instance
}

func (this *StreamServerFactory) GetServer(serverId string) *Server {
	return this.sseServers[serverId]
}

func (this *StreamServerFactory) CreateServer(
	serverId string,
	properties map[string]interface{},
	listener StreamRequestListener) (*Server, error) {

	this.mux.Lock()
	defer this.mux.Unlock()
	server := this.sseServers[serverId]

	if nil == server {
		host := "0.0.0.0"
		if nil != properties["host"] {
			host = properties["host"].(string)
		}

		port := properties[ServerPort].(string)

		path := "/"
		if nil != properties[ConnectionPath] {
			path = properties[ConnectionPath].(string)
		}

		tls := properties[ConnectionTlsEnabled].(bool)
		tlsUploaded := properties[ConnectionUploadCRT].(bool)
		crtLoc := ""
		keyLoc := ""
		if tls {
			crtLoc = properties[ConnectionTlsCRTPath].(string)
			keyLoc = properties[ConnectionTlsKeyPath].(string)
			if tlsUploaded {
				if nil != properties[ConnectionTlsCRT] {
					err := ioutil.WriteFile(crtLoc, properties[ConnectionTlsCRT].([]byte), 0666)
					if err != nil {
						log.Fatal(err)
						return nil, err
					}
				}

				if nil != properties[ConnectionTlsKey] {
					err := ioutil.WriteFile(keyLoc, properties[ConnectionTlsKey].([]byte), 0666)
					if err != nil {
						log.Fatal(err)
						return nil, err
					}
				}
			}
		}

		server = &Server{
			host:     host,
			port:     port,
			path:     path,
			tls:      tls,
			crtLoc:   crtLoc,
			keyLoc:   keyLoc,
			clients:  make(map[string](map[string]*Client)),
			listener: listener,
		}
		this.sseServers[serverId] = server

	}
	return server, nil
}

type Server struct {
	host         string
	port         string
	path         string
	tls          bool
	crtLoc       string
	keyLoc       string
	dataChannels map[string](chan chan []byte)
	clients      map[string](map[string]*Client)
	listener     StreamRequestListener
}

func (this *Server) Start() {
	fmt.Println("Start server, path : ", this.path, ", Server : ", this)
	http.Handle(this.path, this)
	if !this.tls {
		http.ListenAndServe(fmt.Sprintf("%s:%s", this.host, this.port), nil)
	} else {
		crt, err := ioutil.ReadFile(this.crtLoc)
		if err != nil {
			log.Fatal(err)
		}
		key, err := ioutil.ReadFile(this.keyLoc)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Got crt from file : ", crt)
		fmt.Println("Got key from file : ", key)
		http.ListenAndServeTLS(fmt.Sprintf("%s:%s", this.host, this.port), this.crtLoc, this.keyLoc, nil)
	}
}

func (this *Server) Stop() {
}

func (this *Server) RegisterClient(streamId string, client *Client) {
	clientsByQuery := this.clients[streamId]
	if nil == clientsByQuery {
		clientsByQuery = make(map[string]*Client)
		this.clients[streamId] = clientsByQuery
	}
	clientsByQuery[client.GetID()] = client
}

func (this *Server) UnRegisterClient(streamId string, client *Client) {
	clientsByQuery := this.clients[streamId]
	if nil != clientsByQuery {
		delete(clientsByQuery, client.GetID())
	}
}

func (this *Server) SendData(streamId string, data []byte) {
	/* Clients which subscribe to all streams */
	clients := this.clients[streamId]
	if nil != clients {
		for _, client := range clients {
			client.UpdateJPEG(data)
		}
	}
}

func (this *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	pos := strings.LastIndex(req.URL.Path, "/")
	path := req.URL.Path[pos+1:]
	id := req.RemoteAddr
	fmt.Println("Client request in, id : ", id, ", request URI : ", path)
	client := NewClient(id)
	this.RegisterClient(path, client)
	fmt.Println("After registered, clients : ", this.clients)
	client.Listening(res)
	fmt.Println("Client quit, id : ", id, ", request URI : ", path)
	this.UnRegisterClient(path, client)
	fmt.Println("After client unregistered, clients : ", this.clients)
}

type Client struct {
	id            string
	dataChannels  map[chan []byte]bool
	frame         []byte
	frameInterval time.Duration
	lock          sync.Mutex
}

func (this *Client) GetID() string {
	return this.id
}

const MJPEG_BOUNDARY = "MJPEGBOUNDARY"
const HEADER = "\r\n" +
	"--" + MJPEG_BOUNDARY + "\r\n" +
	"Content-Type: image/jpeg\r\n" +
	"Content-Length: %d\r\n" +
	"X-Timestamp: 0.000000\r\n" +
	"\r\n"

func (this *Client) Listening(
	res http.ResponseWriter,
) {
	log.Println("Client channel:", this.id, "connected")
	res.Header().Add("Content-Type", "multipart/x-mixed-replace;boundary="+MJPEG_BOUNDARY)

	c := make(chan []byte)
	this.lock.Lock()
	this.dataChannels[c] = true
	this.lock.Unlock()

	for {
		time.Sleep(this.frameInterval)
		b := <-c
		_, err := res.Write(b)
		if err != nil {
			break
		}
	}

	this.lock.Lock()
	delete(this.dataChannels, c)
	this.lock.Unlock()
	log.Println("Client channel:", this.id, "disconnected")
}

func (this *Client) UpdateJPEG(jpeg []byte) {
	header := fmt.Sprintf(HEADER, len(jpeg))
	if len(this.frame) < len(jpeg)+len(header) {
		this.frame = make([]byte, (len(jpeg)+len(header))*2)
	}

	copy(this.frame, header)
	copy(this.frame[len(header):], jpeg)

	this.lock.Lock()
	for c := range this.dataChannels {
		select {
		case c <- this.frame:
		default:
		}
	}
	this.lock.Unlock()
}

func NewClient(id string) *Client {
	client := &Client{
		id:            id,
		dataChannels:  make(map[chan []byte]bool),
		frame:         make([]byte, len(HEADER)),
		frameInterval: 50 * time.Millisecond,
	}
	return client
}
