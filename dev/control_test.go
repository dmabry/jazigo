package dev

import (
	"bytes"
	"testing"
)

type controlLogger struct {
	*testing.T
}

func (t *controlLogger) Printf(format string, v ...interface{}) {
	t.Logf("controlLogger: "+format, v...)
}

func TestControl1(t *testing.T) {
	logger := &controlLogger{t}
	debug := false

	empty := []byte{}
	crlf := []byte{CR, LF}
	four := []byte("1234")
	five := []byte("12345")
	oneLF := []byte{LF}
	oneBS := []byte{BS}
	oneCR := []byte{CR}
	fiveBS := append([]byte("12345"), BS)
	fiveCR := append([]byte("12345"), CR)
	BSfive := append([]byte{BS}, []byte("12345")...)
	CRfive := append([]byte{CR}, []byte("12345")...)
	middleBS := []byte{'1', '2', '3', BS, '4', '5'}
	middleCR := []byte{'1', '2', '3', CR, '4', '5'}

	control(t, debug, logger, "empty", empty, empty, empty, empty)
	control(t, debug, logger, "bufCRLF", crlf, empty, crlf, empty)
	control(t, debug, logger, "suffixCRLF", empty, crlf, empty, crlf)
	control(t, debug, logger, "bothCRLF", crlf, crlf, crlf, crlf)
	control(t, debug, logger, "no-control", five, five, five, five)
	control(t, debug, logger, "LF", oneLF, oneLF, oneLF, oneLF)
	control(t, debug, logger, "BS", oneBS, oneBS, empty, empty)
	control(t, debug, logger, "CR", oneCR, oneCR, empty, empty)

	control(t, debug, logger, "suffix-BS1", empty, oneBS, empty, empty)
	control(t, debug, logger, "suffix-BS2", five, oneBS, four, empty)
	control(t, debug, logger, "suffix-BSfive1", empty, BSfive, empty, five)
	control(t, debug, logger, "suffix-BSfive2", five, BSfive, four, five)
	control(t, debug, logger, "suffix-fiveBS1", empty, fiveBS, empty, four)
	control(t, debug, logger, "suffix-fiveBS2", five, fiveBS, five, four)
	control(t, debug, logger, "suffix-middleBS1", empty, middleBS, empty, []byte("1245"))
	control(t, debug, logger, "suffix-middleBS2", five, middleBS, five, []byte("1245"))

	control(t, debug, logger, "suffix-CR1", empty, oneCR, empty, empty)
	control(t, debug, logger, "suffix-CR2", five, oneCR, empty, empty)
	control(t, debug, logger, "suffix-CRfive1", empty, CRfive, empty, five)
	control(t, debug, logger, "suffix-CRfive2", five, CRfive, empty, five)
	control(t, debug, logger, "suffix-fiveCR1", empty, fiveCR, empty, empty)
	control(t, debug, logger, "suffix-fiveCR2", five, fiveCR, empty, empty)
	control(t, debug, logger, "suffix-middleCR1", empty, middleCR, empty, []byte("45"))
	control(t, debug, logger, "suffix-middleCR2", five, middleCR, empty, []byte("45"))
}

func TestPrefixM(t *testing.T) {
	expectPrefixM(t, []byte("1m"), 2, true)
	expectPrefixM(t, []byte("12m"), 3, true)
	expectPrefixM(t, []byte("12mx"), 3, true)
	expectPrefixM(t, []byte(""), 0, false)
	expectPrefixM(t, []byte("1"), 0, false)
	expectPrefixM(t, []byte("m"), 0, false)
	expectPrefixM(t, []byte("12"), 0, false)
	expectPrefixM(t, []byte("12a"), 0, false)
	expectPrefixM(t, []byte("a"), 0, false)
	expectPrefixM(t, []byte("a1"), 0, false)
	expectPrefixM(t, []byte("x12m"), 0, false)
}

func expectPrefixM(t *testing.T, input []byte, wantSize int, wantFound bool) {
	size, found := prefixNumberM(input)
	if size != wantSize || found != wantFound {
		t.Errorf("expectPrefixM: input=%v wantSize=%d wantFound=%v gotSize=%d gotFound=%v", input, wantSize, wantFound, size, found)
	}
}

func control(t *testing.T, debug bool, logger hasPrintf, label string, inputBuf, inputSuffix, expectedBuf, expectedSuffix []byte) {
	buf := clone(inputBuf)
	suffix := clone(inputSuffix)

	gotBuf, gotSuffix := removeControlChars(logger, debug, buf, suffix)

	if !bytes.Equal(gotBuf, expectedBuf) {
		t.Errorf("%s: buf mismatch: got=%q wanted=%q", label, gotBuf, expectedBuf)
	}

	if !bytes.Equal(gotSuffix, expectedSuffix) {
		t.Errorf("%s: suffix mismatch: got=%q wanted=%q", label, gotSuffix, expectedSuffix)
	}
}

func clone(a []byte) []byte {
	b := make([]byte, len(a))
	copy(b, a)
	return b
}
