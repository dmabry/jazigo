package dev

import (
	"time"
)

const (
	cmdWill = 251 // Unused - Telnet command Will
	cmdWont = 252 // Unused - Telnet command Wont
	cmdDo   = 253 // Unused - Telnet command Do
	cmdDont = 254 // Unused - Telnet command Dont
	cmdIAC  = 255 // Unused - Telnet command Interpret As Command

	// Telnet options constants (unused)
	optEcho           = 1  // Unused - Telnet option Echo
	optSupressGoAhead = 3  // Unused - Telnet option Suppress Go Ahead
	optLinemode       = 34 // Unused - Telnet option Linemode
)

func shift(b []byte, size, offset int) int { // Unused - Byte array shifting function
	copy(b, b[offset:size])
	return size - offset
}

type telnetNegotiationOnly struct{} // Unused

var telnetNegOnly = telnetNegotiationOnly{}

func (e telnetNegotiationOnly) Error() string {
	return "telnetNegotiationOnlyError"
}

func telnetNegotiation(buf []byte, n int, t transp, logger hasPrintf, debug bool) (int, error) { // Unused - Telnet negotiation function

	timeout := 5 * time.Second // FIXME??
	hitNeg := false

	for {
		if n < 3 {
			break
		}
		if buf[0] != cmdIAC {
			break // not IAC
		}
		b1 := buf[1]
		switch b1 {
		case cmdDo, cmdDont:
			opt := buf[2]
			t.SetWriteDeadline(time.Now().Add(timeout)) // FIXME: handle error
			t.Write([]byte{cmdIAC, cmdWont, opt})       // IAC WONT opt - FIXME: handle error
			n = shift(buf, n, 3)
			hitNeg = true
			continue
		case cmdWill, cmdWont:
			opt := buf[2]
			t.SetWriteDeadline(time.Now().Add(timeout)) // FIXME: handle error
			t.Write([]byte{cmdIAC, cmdDont, opt})       // IAC DONT opt - FIXME: handle error
			n = shift(buf, n, 3)
			hitNeg = true
			continue
		}
		break
	}

	if n == 0 && hitNeg {
		return 0, telnetNegOnly
	}

	return n, nil
}
