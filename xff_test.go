package xff

import (
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_none(t *testing.T) {
	res := Parse("")
	assert.Equal(t, "", res)
}

func TestParse_localhost(t *testing.T) {
	res := Parse("127.0.0.1")
	assert.Equal(t, "", res)
}

func TestParse_invalid(t *testing.T) {
	res := Parse("invalid")
	assert.Equal(t, "", res)
}

func TestParse_invalid_sioux(t *testing.T) {
	res := Parse("123#1#2#3")
	assert.Equal(t, "", res)
}

func TestParse_invalid_private_lookalike(t *testing.T) {
	res := Parse("102.3.2.1")
	assert.Equal(t, "102.3.2.1", res)
}

func TestParse_valid(t *testing.T) {
	res := Parse("68.45.152.220")
	assert.Equal(t, "68.45.152.220", res)
}

func TestParse_multi_first(t *testing.T) {
	res := Parse("12.13.14.15, 68.45.152.220")
	assert.Equal(t, "12.13.14.15", res)
}

func TestParse_multi_last(t *testing.T) {
	res := Parse("192.168.110.162, 190.57.149.90")
	assert.Equal(t, "190.57.149.90", res)
}

func TestParse_multi_with_invalid(t *testing.T) {
	res := Parse("192.168.110.162, invalid, 190.57.149.90")
	assert.Equal(t, "190.57.149.90", res)
}

func TestParse_multi_with_invalid2(t *testing.T) {
	res := Parse("192.168.110.162, 190.57.149.90, invalid")
	assert.Equal(t, "190.57.149.90", res)
}

func TestParse_multi_with_invalid_sioux(t *testing.T) {
	res := Parse("192.168.110.162, 190.57.149.90, 123#1#2#3")
	assert.Equal(t, "190.57.149.90", res)
}

func TestParse_ipv6_with_port(t *testing.T) {
	res := Parse("2604:2000:71a9:bf00:f178:a500:9a2d:670d")
	assert.Equal(t, "2604:2000:71a9:bf00:f178:a500:9a2d:670d", res)
}

func TestGetRemoteAddr_ipv4(t *testing.T) {
	r := &http.Request{
		RemoteAddr: "1.2.3.4:1234",
	}
	ra := GetRemoteAddr(r)
	assert.Equal(t, "1.2.3.4:1234", ra)
}

func TestGetRemoteAddr_ipv6(t *testing.T) {
	r := &http.Request{
		RemoteAddr: "[2001:db8:0:1:1:1:1:1]:1234",
	}
	ra := GetRemoteAddr(r)
	assert.Equal(t, "[2001:db8:0:1:1:1:1:1]:1234", ra)
}

func TestGetRemoteAddr_ipv4_with_xff(t *testing.T) {
	r := &http.Request{
		RemoteAddr: "1.2.3.4:1234",
		Header: http.Header{
			"X-Forwarded-For": []string{"100.0.0.1"},
		},
	}
	ra := GetRemoteAddr(r)
	assert.Equal(t, "100.0.0.1:1234", ra)
}

func TestGetRemoteAddr_ipv6_with_xff(t *testing.T) {
	r := &http.Request{
		RemoteAddr: "1.2.3.4:1234",
		Header: http.Header{
			"X-Forwarded-For": []string{"2001:db8:0:1:1:1:1:1"},
		},
	}
	ra := GetRemoteAddr(r)
	assert.Equal(t, "[2001:db8:0:1:1:1:1:1]:1234", ra)
}

func TestToMasks_empty(t *testing.T) {
	ips := []string{}
	masks, err := toMasks(ips)
	assert.Empty(t, masks)
	assert.Nil(t, err)
}

func TestToMasks(t *testing.T) {
	ips := []string{"127.0.0.1/32", "10.0.0.0/8"}
	masks, err := toMasks(ips)
	_, ipnet1, _ := net.ParseCIDR("127.0.0.1/32")
	_, ipnet2, _ := net.ParseCIDR("10.0.0.0/8")
	assert.Equal(t, []net.IPNet{*ipnet1, *ipnet2}, masks)
	assert.Nil(t, err)
}

func TestToMasks_error(t *testing.T) {
	ips := []string{"error"}
	masks, err := toMasks(ips)
	assert.Empty(t, masks)
	assert.Equal(t, &net.ParseError{Type: "CIDR address", Text: "error"}, err)
}

func TestAllowed_all(t *testing.T) {
	m, _ := New(Options{
		AllowedSubnets: []string{},
	})
	assert.True(t, m.allowed("127.0.0.1"))

	m, _ = Default()
	assert.True(t, m.allowed("127.0.0.1"))
}

func TestAllowed_yes(t *testing.T) {
	m, _ := New(Options{
		AllowedSubnets: []string{"127.0.0.0/16"},
	})
	assert.True(t, m.allowed("127.0.0.1"))

	m, _ = New(Options{
		AllowedSubnets: []string{"127.0.0.1/32"},
	})
	assert.True(t, m.allowed("127.0.0.1"))
}

func TestAllowed_no(t *testing.T) {
	m, _ := New(Options{
		AllowedSubnets: []string{"127.0.0.0/16"},
	})
	assert.False(t, m.allowed("127.1.0.1"))

	m, _ = New(Options{
		AllowedSubnets: []string{"127.0.0.1/32"},
	})
	assert.False(t, m.allowed("127.0.0.2"))
}
