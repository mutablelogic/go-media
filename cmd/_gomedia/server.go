package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	// Packages

	"github.com/mutablelogic/go-media/pkg/ffmpeg/httphandler"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServerCommands struct {
	RunServer RunServer `cmd:"" name:"run" help:"Run server."`
}

type RunServer struct {
	// TLS server options
	TLS struct {
		CertFile string `name:"cert" help:"TLS certificate file"`
		KeyFile  string `name:"key" help:"TLS key file"`
	} `embed:"" prefix:"tls."`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

func (cmd *RunServer) Run(ctx *Globals) error {
	// Parse server URL
	url, err := url.Parse(ctx.Endpoint)
	if err != nil {
		return err
	}
	if url.Scheme == "" {
		url.Scheme = "http"
	}

	// Create a TLS config
	var tlsconfig *tls.Config
	if cmd.TLS.CertFile != "" || cmd.TLS.KeyFile != "" {
		tlsconfig, err = httpserver.TLSConfig(url.Hostname(), true, cmd.TLS.CertFile, cmd.TLS.KeyFile)
		if err != nil {
			return err
		}
		if url.Scheme == "" {
			url.Scheme = "https"
		}
	}

	// Set host if not set
	if url.Host == "" {
		url.Host, err = freePort()
		if err != nil {
			return err
		}
	}

	// Register HTTP handlers
	router := http.NewServeMux()
	httphandler.RegisterHandlers(router, url.Path, ctx.manager)

	// Add a "not found" handler
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpresponse.Error(w, httpresponse.ErrNotFound, r.URL.String())
	})

	// Create a HTTP server
	server, err := httpserver.New(listenAddr(url), router, tlsconfig)
	if err != nil {
		return err
	}

	// Run the server
	//fmt.Println(version.ExecName(), version.Version())
	fmt.Println("Listening on", url.String())
	return server.Run(ctx.ctx)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func listenAddr(u *url.URL) string {
	if u.Port() == "" {
		if u.Scheme == "https" {
			return u.Host + ":443"
		} else {
			return u.Host + ":80"
		}
	}
	return u.Host
}

func freePort() (string, error) {
	// Create a listener on port 0 to get a free port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer listener.Close()
	return listener.Addr().String(), nil
}
