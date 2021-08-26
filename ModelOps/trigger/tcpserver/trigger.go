/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package tcpserver

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

const (
	cPort     = "Port"
	cPath     = "Path"
	CONN_HOST = "localhost"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)
const MIN = 1
const MAX = 100

func random() int {
	return rand.Intn(MAX-MIN) + MIN
}

//-============================================-//
//   Entry point register Trigger & factory
//-============================================-//

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&TcpServer{}, &Factory{})
}

//-===============================-//
//     Define Trigger Factory
//-===============================-//

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	settings := &Settings{}
	err := metadata.MapToStruct(config.Settings, settings, true)
	if err != nil {
		return nil, err
	}

	return &TcpServer{settings: settings}, nil
}

//-=========================-//
//      Define Trigger
//-=========================-//

var logger log.Logger

type TcpServer struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	mux      sync.Mutex

	settings *Settings
	handlers []trigger.Handler
}

// implements trigger.Initializable.Initialize
func (this *TcpServer) Initialize(ctx trigger.InitContext) error {
	this.handlers = ctx.GetHandlers()
	logger = ctx.Logger()

	return nil
}

// implements ext.Trigger.Start
func (this *TcpServer) Start() error {
	logger.Info("(Start) Processing handlers")
	for _, handler := range this.handlers {
		handlerSetting := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSetting, true)
		if err != nil {
			return err
		}

		go func() {
			// Listen for incoming connections.
			l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
			if err != nil {
				fmt.Println("Error listening:", err.Error())
				os.Exit(1)
			}
			// Close the listener when the application closes.
			defer l.Close()
			fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
			for {
				// Listen for an incoming connection.
				conn, err := l.Accept()
				if err != nil {
					fmt.Println("Error accepting: ", err.Error())
					os.Exit(1)
				}
				// Handle connections in a new goroutine.
				go this.handleRequest(conn)
			}
		}()
		logger.Info("(Start) Started path = ", handlerSetting.Path, ", port = ", this.settings.Port)
	}
	logger.Info("(Start) Now started")

	return nil
}

func (this *TcpServer) handleRequest(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			break
		}

		result := strconv.Itoa(random()) + "\n"
		c.Write([]byte(string(result)))
	}
	c.Close()
}

// implements ext.Trigger.Stop
func (this *TcpServer) Stop() error {
	return nil
}
