package extyaml_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/hujun-open/extyaml"
)

type testStruct[T any] struct {
	expectMarshal    []byte
	Val              T
	bytesToUnmarshal []byte
	compareFunc      func(v1, v2 T) bool
}

func testFunc[T any](t *testing.T, tcase testStruct[T]) {
	buf, err := extyaml.MarshalExt(tcase)
	if err != nil {
		t.Fatal(err)
	}
	buf = bytes.TrimRight(buf, "\n\r")
	if !bytes.Equal(tcase.expectMarshal, buf) {
		t.Fatalf("marshal result %s is different from expect %s", buf, tcase.expectMarshal)
	}
	out := &testStruct[T]{}
	err = extyaml.UnmarshalExt(tcase.bytesToUnmarshal, out)
	if err != nil {
		t.Fatal(err)
	}
	if !tcase.compareFunc(tcase.Val, out.Val) {
		t.Fatalf("unmarshal result %v is different from expect %v", out.Val, tcase.Val)
	}

}

func TestMAC(t *testing.T) {

	caseList := []testStruct[net.HardwareAddr]{
		{
			Val:              net.HardwareAddr{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
			expectMarshal:    []byte("val: 11:22:33:44:55:66"),
			bytesToUnmarshal: []byte("val: 11:22:33:44:55:66"),
		},
		{
			Val:              net.HardwareAddr{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
			expectMarshal:    []byte("val: 11:22:33:44:55:66"),
			bytesToUnmarshal: []byte("val: 11-22-33-44-55-66"),
		},
	}
	cmpFunc := func(v1, v2 net.HardwareAddr) bool {
		return bytes.Equal(v1, v2)
	}

	for i, c := range caseList {
		c.compareFunc = cmpFunc
		t.Logf("test case %d", i)
		testFunc(t, c)
	}
}

func TestIPNet(t *testing.T) {
	cmpFunc := func(v1, v2 net.IPNet) bool {
		return v1.String() == v2.String()
	}
	caseList := []testStruct[net.IPNet]{
		{
			Val:              net.IPNet{IP: net.ParseIP("192.168.1.0"), Mask: net.CIDRMask(24, 32)},
			expectMarshal:    []byte("val: 192.168.1.0/24"),
			bytesToUnmarshal: []byte("val: 192.168.1.0/24"),
		},
		{
			Val:              net.IPNet{IP: net.ParseIP("2001:dead::"), Mask: net.CIDRMask(96, 128)},
			expectMarshal:    []byte("val: 2001:dead::/96"),
			bytesToUnmarshal: []byte("val: 2001:dead::/96"),
		},
	}
	for i, c := range caseList {
		c.compareFunc = cmpFunc
		t.Logf("test case %d", i)
		testFunc(t, c)
	}
}
