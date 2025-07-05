package dev

import (
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"github.com/udhos/jazigo/conf"
	"github.com/udhos/jazigo/temp"
)

type optionsJunos struct {
	breakConn bool
}

func TestJuniperJunOS1(t *testing.T) {

	// launch bogus test server
	addr := ":2001"
	s, listenErr := spawnServerJuniperJunOS(t, addr, optionsJunos{})
	if listenErr != nil {
		t.Errorf("could not spawn bogus JunOS server: %v", listenErr)
	}

	// run client test
	logger := &testLogger{t}
	tab := NewDeviceTable()
	opt := conf.NewOptions()
	opt.Set(&conf.AppConfig{MaxConcurrency: 3, MaxConfigFiles: 10})
	RegisterModels(logger, tab)
	CreateDevice(tab, logger, "junos", "lab1", "localhost"+addr, "telnet", "lab", "pass", "en", false, nil)

	repo := temp.MakeTempRepo()
	defer temp.CleanupTempRepo()

	requestCh := make(chan FetchRequest)
	errlogPrefix := filepath.Join(repo, "errlog_test.")
	go Spawner(tab, logger, requestCh, repo, errlogPrefix, opt, NewFilterTable(logger))
	good, bad, skip := Scan(tab, tab.ListDevices(), logger, opt.Get(), requestCh)
	if good != 1 || bad != 0 || skip != 0 {
		t.Errorf("good=%d bad=%d skip=%d", good, bad, skip)
	}

	close(requestCh) // shutdown Spawner - we might exit first though

	s.close() // shutdown server

	<-s.done // wait termination of accept loop goroutine
}

func TestJuniperJunOS2(t *testing.T) {

	// launch bogus test server
	addr := ":2002"
	s, listenErr := spawnServerJuniperJunOS(t, addr, optionsJunos{breakConn: true})
	if listenErr != nil {
		t.Errorf("could not spawn bogus JunOS server: %v", listenErr)
	}

	// run client test
	logger := &testLogger{t}
	tab := NewDeviceTable()
	opt := conf.NewOptions()
	opt.Set(&conf.AppConfig{MaxConcurrency: 3, MaxConfigFiles: 10})
	RegisterModels(logger, tab)
	CreateDevice(tab, logger, "junos", "lab1", "localhost"+addr, "telnet", "lab", "pass", "en", false, nil)

	repo := temp.MakeTempRepo()
	defer temp.CleanupTempRepo()

	requestCh := make(chan FetchRequest)
	errlogPrefix := filepath.Join(repo, "errlog_test.")
	go Spawner(tab, logger, requestCh, repo, errlogPrefix, opt, NewFilterTable(logger))
	good, bad, skip := Scan(tab, tab.ListDevices(), logger, opt.Get(), requestCh)
	if good != 0 || bad != 1 || skip != 0 {
		t.Errorf("good=%d bad=%d skip=%d", good, bad, skip)
	}

	close(requestCh) // shutdown Spawner - we might exit first though

	s.close() // shutdown server

	<-s.done // wait termination of accept loop goroutine
}

func spawnServerJuniperJunOS(t *testing.T, addr string, options optionsJunos) (*testServer, error) {

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &testServer{listener: ln, done: make(chan int)}

	go acceptLoopJuniperJunOS(t, s, handleConnectionJuniperJunOS, options)

	return s, nil
}

func acceptLoopJuniperJunOS(t *testing.T, s *testServer, handler func(*testing.T, net.Conn, optionsJunos), options optionsJunos) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			t.Logf("acceptLoopJuniperJunOS: accept failure, exiting: %v", err)
			break
		}
		go handler(t, conn, options)
	}

	close(s.done)
}

func handleConnectionJuniperJunOS(t *testing.T, c net.Conn, options optionsJunos) {
	defer c.Close()

	buf := make([]byte, 1000)

	// send username prompt
	if _, err := c.Write([]byte("hostname (ttyp0)\n\nlogin: ")); err != nil {
		t.Logf("handleConnectionJuniperJunOS: send username prompt error: %v", err)
		return
	}

	// consume username
	if _, err := c.Read(buf); err != nil {
		t.Logf("handleConnectionJuniperJunOS: read username error: %v", err)
		return
	}

	// send password prompt
	if _, err := c.Write([]byte("\nPassword: ")); err != nil {
		t.Logf("handleConnectionJuniperJunOS: send password prompt error: %v", err)
		return
	}

	// consume password
	if _, err := c.Read(buf); err != nil {
		t.Logf("handleConnectionJuniperJunOS: read password error: %v", err)
		return
	}

	if _, err := c.Write([]byte("\n--- JUNOS 11.2R1.2 built 2011-06-22 02:55:58 UTC")); err != nil {
		t.Logf("handleConnectionJuniperJunOS: send banner error: %v", err)
		return
	}

LOOP:
	for {
		// send command prompt
		if _, err := c.Write([]byte("\n{master:0}\nlab@host.domain> ")); err != nil {
			t.Logf("handleConnectionJuniperJunOS: send command prompt error: %v", err)
			return
		}

		// consume command
		if _, err := c.Read(buf); err != nil {
			if err == io.EOF {
				return // peer closed connection
			}
			t.Logf("handleConnectionJuniperJunOS: read command error: %v", err)
			return
		}

		str := string(buf)

		switch {
		case strings.HasPrefix(str, "q"): //quit
			break LOOP
		case strings.HasPrefix(str, "ex"): //exit
			break LOOP
		case strings.HasPrefix(str, "set cli"):
		case strings.HasPrefix(str, "show conf"):
			if options.breakConn {
				return // break connection (defer/close)
			}

			if _, err := c.Write([]byte("\nshow running-configuration")); err != nil {
				t.Logf("handleConnectionJuniperJunOS: send sh run error: %v", err)
				return
			}
		default:
			if _, err := c.Write([]byte("\nIgnoring unknown command")); err != nil {
				t.Logf("handleConnectionJuniperJunOS: send unknown command error: %v", err)
				return
			}
		}

	}

	// send bye
	if _, err := c.Write([]byte("\nbye\n")); err != nil {
		t.Logf("handleConnectionJuniperJunOS: send bye error: %v", err)
		return
	}

}
