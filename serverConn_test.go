package http2

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"io"
	"sync"
	"testing"
	"time"
)

func TestServerConn_closeIdleConn(t *testing.T) {
	s := &Server{
		s: &fasthttp.Server{
			Handler: func(ctx *fasthttp.RequestCtx) {
				io.WriteString(ctx, "Hello world")
			},
			ReadTimeout: time.Second * 5,
			IdleTimeout: time.Second * 1,
		},
		cnf: ServerConfig{
			Debug: false,
		},
	}

	s.cnf.defaults()

	ln := fasthttputil.NewInmemoryListener()

	go serve(s, ln)

	worker := 40
	loopCount := 100
	wg := &sync.WaitGroup{}
	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < loopCount; j++ {
				func() {
					c, err := ln.Dial()
					if err != nil {
						t.Fatalf("dial fail, %v", err)
					}
					nc := NewConn(c, ConnOpts{})
					if err := nc.doHandshake(); err != nil {
						t.Fatalf("handshake err, %v", err)
					}
					h1 := makeHeaders(3, nc.enc, true, true, map[string]string{
						string(StringAuthority): "localhost",
						string(StringMethod):    "GET",
						string(StringPath):      "/hello/world",
						string(StringScheme):    "https",
					})
					if err := nc.writeFrame(h1); err != nil {
						t.Fatalf("writeFrame err, %v", err)
					}
					time.Sleep(time.Millisecond * 999)
					if err := nc.Close(); err != nil && err.Error() != `connection closed` {
						t.Errorf("close err, %v, %T", err, err)
					}
				}()
			}
		}()
	}

	wg.Wait()
}
