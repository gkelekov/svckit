package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/koding/websocketproxy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
)

type config struct {
	Services []string
	HTTP     struct {
		Port  int
		Proxy []struct {
			URL     string
			Backend string
		}
	}
	services map[string]*service
}

func (c *config) start() error {
	for _, key := range c.Services {
		service := c.services[key]
		if service == nil {
			warn("Service %s not found\n", key)
			continue
		}
		if err := service.Go(); err != nil {
			log.S("service", service.Name).Error(err)
			warn("Failed to start %s\n", service)
			return err
		}
	}
	info(">")
	return nil
}

func (c *config) stop() {
	for i := len(c.Services) - 1; i >= 0; i-- {
		service := c.services[c.Services[i]]
		service.stop()
	}
}

func (c *config) startHTTP() error {
	if c.HTTP.Port == 0 {
		return nil
	}
	for _, p := range c.HTTP.Proxy {
		u, err := url.Parse(p.Backend)
		if err != nil {
			log.Error(err)
			return err
		}
		if strings.HasPrefix(p.Backend, "http://") {
			http.Handle(p.URL, httputil.NewSingleHostReverseProxy(u))
			continue
		}
		if strings.HasPrefix(p.Backend, "ws://") {
			http.Handle(p.URL, websocketproxy.NewProxy(u))
			continue
		}
		fs := http.FileServer(http.Dir(env.ExpandPath(p.Backend)))
		http.Handle(p.URL, fs)
	}
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", c.HTTP.Port), nil)
	}()
	return nil
}
