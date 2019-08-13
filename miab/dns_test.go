package miab

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func basicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
}

func getTestServer(t *testing.T, method, response string, rtype ResourceType, testUri bool, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Authorization") != basicAuthHeader("test", "secret") {
			t.Error("authentication failed")
		}

		if r.Method != method {
			t.Error("invalid http method")
		}

		if testUri && r.RequestURI != fmt.Sprintf("/%s", dnsPath("test.example.org", rtype)) {
			t.Error("invalid request uri")
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		if buf.String() != body {
			t.Errorf("invalid body, got: %s = expected: %s", buf.String(), body)
		}

		_, err := fmt.Fprintln(w, response)
		if err != nil {
			t.Fail()
		}
	}))
}

func TestResourceType_IsValid(t *testing.T) {
	testCases := []struct {
		rtype ResourceType
		want  bool
	}{
		{NONE, false},
		{A, true},
		{AAAA, true},
		{TXT, true},
		{CNAME, true},
		{MX, true},
		{SRV, true},
		{SSHFP, true},
		{CAA, true},
		{NS, true},
		{ResourceType(`X`), false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.rtype), func(t *testing.T) {
			if tc.rtype.IsValid() != tc.want {
				t.Errorf("%v.IsValid() wanted: %v, got: %v", tc.rtype, tc.want, tc.rtype.IsValid())
			}
		})
	}
}

func TestParseDnsResource(t *testing.T) {
	testCases := []struct {
		rtype string
		want  ResourceType
	}{
		{"foo", NONE},
		{"a", A},
		{"aaAA", AAAA},
		{"TXT", TXT},
		{"cName", CNAME},
		{"MX", MX},
		{"srv", SRV},
		{"ssHfP", SSHFP},
		{"caa", CAA},
		{"NS", NS},
		{"AAA", NONE},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.rtype), func(t *testing.T) {
			if res, err := ParseDnsResource(tc.rtype); res != tc.want {
				t.Error(err)
			}
		})
	}
}

func TestDnsPath(t *testing.T) {

	testCases := []struct {
		qname string
		rtype ResourceType
		want  string
	}{
		{"", NONE, "admin/dns/custom"},
		{"", AAAA, "admin/dns/custom"},
		{"test.example.org", NONE, "admin/dns/custom/test.example.org"},
		{"test.example.org", A, "admin/dns/custom/test.example.org/A"},
		{"test.example.org", AAAA, "admin/dns/custom/test.example.org/AAAA"},
		{"test.example.org", TXT, "admin/dns/custom/test.example.org/TXT"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("qname:%s - rtype:%s", tc.qname, tc.rtype), func(t *testing.T) {

			path := dnsPath(tc.qname, tc.rtype)

			if tc.want != path {
				t.Errorf("expected: %v, got %v", tc.want, path)
			}
		})
	}
}

func TestGetDns(t *testing.T) {

	ts := getTestServer(t, http.MethodGet, `[{"qname": "test.example.org",
    	"rtype": "A",
    	"value": "127.0.0.1"},
    	{"qname": "*.example.org",
    	"rtype": "A",
    	"value": "127.0.0.2"}]`, NONE, false, "")
	defer ts.Close()

	c, _ := NewConfig("test", "secret", ts.URL)

	r, err := GetDns(c, "", A)
	if err != nil {
		t.Fail()
	}

	if r == nil {
		t.Fail()
	}

	if len(r) != 2 {
		t.Fail()
	}

	for i := 0; i < len(r); i++ {
		if r[i].QName == "test.example.org" && r[i].Value != "127.0.0.1" {
			t.Fail()
		}
		if r[i].QName == "*.example.org" && r[i].Value != "127.0.0.2" {
			t.Fail()
		}
	}

}

func TestSetDns(t *testing.T) {

	ts := getTestServer(t, http.MethodPut, `updated DNS: 127.0.0.1`, A, true, "127.0.0.1")
	defer ts.Close()

	c, _ := NewConfig("test", "secret", ts.URL)

	b, err := SetDns(c, "test.example.org", A, "127.0.0.1")
	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}
}

func TestAddDns(t *testing.T) {

	ts := getTestServer(t, http.MethodPost, `updated DNS: 127.0.0.1`, A, true, "127.0.0.1")
	defer ts.Close()

	c, _ := NewConfig("test", "secret", ts.URL)

	b, err := AddDns(c, "test.example.org", A, "127.0.0.1")
	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}
}

func TestDeleteDns(t *testing.T) {

	ts := getTestServer(t, http.MethodDelete, `updated DNS: 127.0.0.1`, A, true, "127.0.0.1")
	defer ts.Close()

	c, _ := NewConfig("test", "secret", ts.URL)

	b, err := DeleteDns(c, "test.example.org", A, "127.0.0.1")
	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}
}

func TestUpdateDns(t *testing.T) {

	ts := getTestServer(t, http.MethodPut, `updated DNS: 127.0.0.1`, A, true, "")
	defer ts.Close()

	c, _ := NewConfig("test", "secret", ts.URL)
	b, err := SetOrAddAddressRecord(c, "tcp4", "test.example.org", "")

	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}

	ts = getTestServer(t, http.MethodPut, `updated DNS: 127.0.0.1`, A, true, "127.0.0.1")
	defer ts.Close()

	c, _ = NewConfig("test", "secret", ts.URL)
	b, err = SetOrAddAddressRecord(c, "tcp4", "test.example.org", "127.0.0.1")

	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}

	ts = getTestServer(t, http.MethodPut, `updated DNS: 127.0.0.1`, AAAA, true, "")
	defer ts.Close()

	c, _ = NewConfig("test", "secret", ts.URL)
	b, err = SetOrAddAddressRecord(c, "tcp6", "test.example.org", "")

	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}

	ts = getTestServer(t, http.MethodPut, `updated DNS: 127.0.0.1`, AAAA, true, "127.0.0.1")
	defer ts.Close()

	c, _ = NewConfig("test", "secret", ts.URL)
	b, err = SetOrAddAddressRecord(c, "tcp6", "test.example.org", "127.0.0.1")

	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}
}

func TestUpdateDns4(t *testing.T) {

	ts := getTestServer(t, http.MethodPut, `updated DNS: 127.0.0.1`, A, true, "")
	defer ts.Close()

	c, _ := NewConfig("test", "secret", ts.URL)
	b, err := UpdateDns4(c, "test.example.org", "")

	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}
}

func TestUpdateDns6(t *testing.T) {

	ts := getTestServer(t, http.MethodPut, `updated DNS: 127.0.0.1`, AAAA, true, "")
	defer ts.Close()

	c, _ := NewConfig("test", "secret", ts.URL)
	b, err := UpdateDns6(c, "test.example.org", "")

	if err != nil {
		t.Errorf("error from function: %v", err)
	}

	if !b {
		t.Error("expected true")
	}
}
