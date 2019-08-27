package miab

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testRec1 = Record{
	QName: "example.org",
	RType: A,
	Value: "127.0.0.1",
}

var testRec2 = Record{
	QName: "example.org",
	RType: AAAA,
	Value: "::1",
}

var testRecs = Records{
	testRec1,
	testRec2,
}

func basicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
}

func getDnsTestServer(t *testing.T, method string, status int, response string, rtype ResourceType, testUri bool, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var err error = nil

		if r.Header.Get("Authorization") != basicAuthHeader("test", "secret") {
			t.Error("authentication failed")
			err = errors.New(fmt.Sprintf("authentication failed; want: %s - got: %s",
				basicAuthHeader("test", "secret"), r.Header.Get("Authorization")))
		}

		if r.Method != method {
			err = errors.New(fmt.Sprintf("invalid http method, want: %s, got: %s", method, r.Method))
		}

		if testUri && r.RequestURI != fmt.Sprintf("/%s", dnsPath("test.example.org", rtype)) {
			err = errors.New(fmt.Sprintf("invalid request uri, want: %s, got: %s",
				fmt.Sprintf("/%s", dnsPath("test.example.org", rtype)), r.RequestURI))
		}

		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		if buf.String() != body {
			err = errors.New(fmt.Sprintf("invalid body, want: %s, got: %s", body, buf.String()))
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(response))
		}
	}))
}

func TestRecord_String(t *testing.T) {

	want := "example.org	A	127.0.0.1"
	got := testRec1.String()

	if got != want {
		t.Errorf("wrong format, want: \n***%s***\ngot:\n***%s***", want, got)
	}
}

func TestRecord_ToString(t *testing.T) {

	var record Record
	err := json.Unmarshal([]byte(testRec1.ToString(JSON)), &record)
	if err != nil || record.QName != "example.org" || record.RType != "A" || record.Value != "127.0.0.1" {
		t.Error("Unable to unmarshal generated json", err)
	}

	err = yaml.Unmarshal([]byte(testRec1.ToString(YAML)), &record)
	if err != nil || record.QName != "example.org" || record.RType != "A" || record.Value != "127.0.0.1" {
		t.Error("Unable to unmarshal generated yaml", err)
	}

	want := fmt.Sprintf("%s\n\"example.org\", \"A\", \"127.0.0.1\"\n", csvDnsHead)
	c := testRec1.ToString(CSV)
	if c != want {
		t.Errorf("wrong format, want: \n%s\n\ngot:\n%s", want, c)
	}
}

func TestRecords_String(t *testing.T) {

	want := "example.org	A	127.0.0.1\nexample.org	AAAA	::1"
	s := testRecs.String()

	if s != want {
		t.Errorf("wrong format, want:\n***%s***\ngot:\n***%s***", want, s)
	}
}

func TestRecords_ToString(t *testing.T) {

	var records Records
	err := json.Unmarshal([]byte(testRecs.ToString(JSON)), &records)
	if err != nil || records[0].QName != "example.org" || records[0].RType != A ||
		records[0].Value != "127.0.0.1" || records[1].QName != "example.org" || records[1].RType != AAAA ||
		records[1].Value != "::1" {
		t.Error("Unable to unmarshal generated json", err)
	}

	err = yaml.Unmarshal([]byte(testRecs.ToString(YAML)), &records)
	if err != nil || records[0].QName != "example.org" || records[0].RType != A ||
		records[0].Value != "127.0.0.1" || records[1].QName != "example.org" || records[1].RType != AAAA ||
		records[1].Value != "::1" {
		t.Error("Unable to unmarshal generated json", err)
	}

	expectedCsv := fmt.Sprintf("%s\n\"example.org\", \"A\", \"127.0.0.1\"\n\"example.org\", \"AAAA\", \"::1\"\n", csvDnsHead)
	c := testRecs.ToString(CSV)
	if c != expectedCsv {
		t.Errorf("wrong format, expected: \n%s\n\ngot:\n%s", expectedCsv, c)
	}
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

	testCases := []struct {
		name         string
		qname        string
		statusCode   int
		responseBody string
		want         Records
		wantError    bool
	}{
		{"GetDns OK", "", 200, `
[{"qname": "test.example.org","rtype": "A","value": "127.0.0.1"},
{"qname": "*.example.org","rtype": "A","value": "127.0.0.2"}]`,
			Records{
				Record{QName: "test.example.org", RType: A, Value: "127.0.0.1"},
				Record{QName: "*.example.org", RType: A, Value: "127.0.0.2"},
			}, false},
		{"GetDns invalid qname", "fooBar", 404, "", nil, true},
		{"GetDns invalid server Status", "test.example.org", 503, "", nil, true},
		{"GetDns invalid server Status", "test.example.org", 503, "No Gateway", nil, true},
		{"GetDns invalid response", "", 200, `
"test.example.org", "A", "127.0.0.1",
"*.example.org", "A", "127.0.0.2"
`, nil, true},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodGet, tc.statusCode, tc.responseBody, NONE, false, "")
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)

			got, err := GetDns(c, tc.qname, A)

			if tc.wantError {
				if err == nil {
					t.Errorf("%s failed, want error, got: nil", tc.name)
				}
			} else {
				if err != nil {
					t.Errorf("%s failed, got error: %v", tc.name, err)
				}
			}

			if tc.want != nil {
				if len(tc.want) != len(got) {
					t.Errorf("%s failed, want: %v - got: %v", tc.name, len(tc.want), len(got))
				}

				for i := 0; i < len(got); i++ {
					if tc.want[i] != got[i] {
						t.Errorf("%s failed at index %d, want: %v - got: %v", tc.name, i, tc.want[i], got[i])
					}
				}
			}
		})
	}
}

func TestSetDns(t *testing.T) {
	testSetAddDel(t, "SetDns", http.MethodPut)
}

func TestAddDns(t *testing.T) {
	testSetAddDel(t, "AddDns", http.MethodPost)
}

func TestDeleteDns(t *testing.T) {
	testSetAddDel(t, "DeleteDns", http.MethodDelete)
}

func testSetAddDel(t *testing.T, name, httpMethod string) {
	testCases := []struct {
		name         string
		qname        string
		value        string
		rtype        ResourceType
		statusCode   int
		responseBody string
		want         bool
		wantError    bool
	}{
		{fmt.Sprintf("%s OK A", name), "test.example.org", "127.0.0.1", A, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s OK AAAA", name), "test.example.org", "::1", AAAA, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s OK TXT", name), "test.example.org", "FooBar", TXT, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s OK CNAME", name), "test.example.org", "foo.example.com.", CNAME, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s OK MX", name), "test.example.org", "example.org.", MX, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s OK SRV", name), "test.example.org", "_sip._tcp.example.org.", SRV, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s OK SSHFP", name), "test.example.org", "SSHFP 2 1 123456789abcdef67890123456789abcdef67890", SSHFP, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s OK CAA", name), "test.example.org", `0 issue "ca.example.net"`, CAA, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s OK NS", name), "test.example.org", "example.org", NS, 200, "updated DNS:", true, false},
		{fmt.Sprintf("%s invalid rtype NONE", name), "test.example.org", "127.0.0.1", NONE, 200, "", false, true},
		{fmt.Sprintf("%s invalid rtype B", name), "test.example.org", "127.0.0.1", NONE, 200, "", false, true},
		{fmt.Sprintf("%s invalid qname", name), "test%example_org", "127.0.0.1", A, 200, "", false, true},
		{fmt.Sprintf("%s invalid server Status", name), "test.example.org", "127.0.0.1", A, 503, "", false, true},
		{fmt.Sprintf("%s invalid server Status", name), "test.example.org", "127.0.0.1", A, 503, "No Gateway", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := getDnsTestServer(t, httpMethod, tc.statusCode, tc.responseBody, tc.rtype, true, tc.value)
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)

			var got bool
			var err error

			switch httpMethod {
			case http.MethodPut:
				got, err = SetDns(c, tc.qname, tc.rtype, tc.value)
			case http.MethodPost:
				got, err = AddDns(c, tc.qname, tc.rtype, tc.value)
			case http.MethodDelete:
				got, err = DeleteDns(c, tc.qname, tc.rtype, tc.value)
			}

			if tc.wantError && err == nil {
				t.Errorf("%s failed, want error, got: nil", tc.name)

			} else if !tc.wantError && err != nil {
				t.Errorf("%s failed, got error: %v", tc.name, err)
			}

			if tc.want != got {
				t.Errorf("%s failed, want: %v - got: %v", tc.name, tc.want, got)
			}
		})
	}
}

func TestSetOrAddAddressRecord_Set(t *testing.T) {
	testSetOrAddAddressRecord(t, "SetAddressRecord", http.MethodPut)
}

func TestSetOrAddAddressRecord_Add(t *testing.T) {
	testSetOrAddAddressRecord(t, "AddAddressRecord", http.MethodPost)
}

func testSetOrAddAddressRecord(t *testing.T, name, method string) {

	testCases := []struct {
		name         string
		qname        string
		value        string
		network      NetworkType
		statusCode   int
		responseBody string
		want         bool
		wantError    bool
	}{
		{
			name: fmt.Sprintf("%s OK A (tcp4)", name), qname: "test.example.org", value: "127.0.0.1",
			network: TCP4, statusCode: 200, responseBody: "updated DNS:", want: true, wantError: false,
		},
		{
			name: fmt.Sprintf("%s OK A (tcp6)", name), qname: "test.example.org", value: "::1",
			network: TCP6, statusCode: 200, responseBody: "updated DNS:", want: true, wantError: false,
		},
		{
			name: fmt.Sprintf("%s OK A (tcp4) no val", name), qname: "test.example.org", value: "",
			network: TCP4, statusCode: 200, responseBody: "updated DNS:", want: true, wantError: false,
		},
		{
			name: fmt.Sprintf("%s OK A (tcp4) no val", name), qname: "test.example.org", value: "",
			network: TCP4, statusCode: 200, responseBody: "updated DNS:", want: true, wantError: false,
		},
		{
			name: fmt.Sprintf("%s OK AAAA (tcp6) no val", name), qname: "test.example.org", value: "",
			network: TCP6, statusCode: 200, responseBody: "updated DNS:", want: true, wantError: false,
		},
		{
			name: fmt.Sprintf("%s OK AAAA (tcp4)", name), qname: "test.example.org", value: "127.0.0.1",
			network: TCP4, statusCode: 200, responseBody: "updated DNS:", want: true, wantError: false,
		},
		{
			name: fmt.Sprintf("%s OK AAAA (tcp6)", name), qname: "test.example.org", value: "::1",
			network: TCP6, statusCode: 200, responseBody: "updated DNS:", want: true, wantError: false,
		},
		{
			name: fmt.Sprintf("%s OK AAAA (tcp6) no val", name), qname: "test.example.org", value: "",
			network: TCP6, statusCode: 200, responseBody: "updated DNS:", want: true, wantError: false,
		},
		{
			name: fmt.Sprintf("%s NOK A (tcp4) server error", name), qname: "test.example.org", value: "127.0.0.1",
			network: TCP4, statusCode: 503, responseBody: "", want: false, wantError: true,
		},
		{
			name: fmt.Sprintf("%s NOK AAAA (tcp6) server error", name), qname: "test.example.org", value: "::1",
			network: TCP6, statusCode: 503, responseBody: "updated DNS:", want: false, wantError: true,
		},
		{
			name: fmt.Sprintf("%s NOK A (tcp4) server error no val", name), qname: "test.example.org", value: "",
			network: TCP4, statusCode: 503, want: false, wantError: true,
		},
		{
			name: fmt.Sprintf("%s NOK AAAA (tcp6) server error no val", name), qname: "test.example.org", value: "",
			network: TCP6, statusCode: 503, want: false, wantError: true,
		},
		{
			name: fmt.Sprintf("%s NOK invalid network", name), qname: "test.example.org", value: "",
			network: "udp", statusCode: 200, want: false, wantError: true,
		},
		{
			name: fmt.Sprintf("%s NOK invalid qname", name), qname: "test,example,org", value: "",
			network: "TCP4", statusCode: 200, want: false, wantError: true,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			var rtype = A
			if tc.network == TCP6 {
				rtype = AAAA
			}
			ts := getDnsTestServer(t, method, tc.statusCode, tc.responseBody, rtype, true, tc.value)
			defer ts.Close()

			c, _ := NewConfig("test", "secret", ts.URL)
			got, err := SetOrAddAddressRecord(c, tc.network, tc.qname, tc.value, name == "AddAddressRecord")

			if tc.wantError && err == nil {
				t.Errorf("%s failed, want error, got: nil", tc.name)

			} else if !tc.wantError && err != nil {
				t.Errorf("%s failed, got error: %v", tc.name, err)
			}

			if tc.want != got {
				t.Errorf("%s failed, want: %v - got: %v", tc.name, tc.want, got)
			}
		})
	}
}

func TestUpdateDns4(t *testing.T) {

	testCases := []struct {
		name      string
		qname     string
		value     string
		want      bool
		wantError bool
	}{
		{"UpdateDns4 OK val", "test.example.org", "127.0.0.1", true, false},
		{"UpdateDns4 OK non val", "test.example.org", "", true, false},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			ts := getDnsTestServer(t, http.MethodPut, 200, fmt.Sprintf("updated DNS: %s", tc.value), A, true, tc.value)
			defer ts.Close()

			c, _ := NewConfig("test", "secret", ts.URL)
			got, err := UpdateDns4(c, tc.qname, tc.value)

			if tc.wantError && err == nil {
				t.Errorf("%s failed, want error, got: nil", tc.name)

			} else if !tc.wantError && err != nil {
				t.Errorf("%s failed, got error: %v", tc.name, err)
			}

			if tc.want != got {
				t.Errorf("%s failed, want: %v - got: %v", tc.name, tc.want, got)
			}
		})
	}
}

func TestUpdateDns6(t *testing.T) {

	testCases := []struct {
		name      string
		qname     string
		value     string
		want      bool
		wantError bool
	}{
		{"UpdateDns6 OK val", "test.example.org", "::1", true, false},
		{"UpdateDns6 OK non val", "test.example.org", "", true, false},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			ts := getDnsTestServer(t, http.MethodPut, 200, fmt.Sprintf("updated DNS: %s", tc.value), AAAA, true, tc.value)
			defer ts.Close()

			c, _ := NewConfig("test", "secret", ts.URL)
			got, err := UpdateDns6(c, tc.qname, tc.value)

			if tc.wantError && err == nil {
				t.Errorf("%s failed, want error, got: nil", tc.name)

			} else if !tc.wantError && err != nil {
				t.Errorf("%s failed, got error: %v", tc.name, err)
			}

			if tc.want != got {
				t.Errorf("%s failed, want: %v - got: %v", tc.name, tc.want, got)
			}
		})
	}
}
