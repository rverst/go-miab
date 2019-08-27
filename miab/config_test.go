package miab

import (
	"fmt"
	"testing"
)

func TestNewConfig(t *testing.T) {

	testCasesUrlPass := []struct {
		url        string
		wantScheme string
		wantDomain string
	}{
		{"http://example.org", "http", "example.org"},
		{"http://example.org/", "http", "example.org"},
		{"https://example.org", "https", "example.org"},
		{"https://example.org/", "https", "example.org"},
		{"http://*.example.org", "http", "*.example.org"},
		{"http://*.example.org/", "http", "*.example.org"},
		{"https://*.example.org", "https", "*.example.org"},
		{"https://*.example.org/", "https", "*.example.org"},
		{"http://sub.example.org", "http", "sub.example.org"},
		{"http://sub.example.org/", "http", "sub.example.org"},
		{"https://sub.example.org", "https", "sub.example.org"},
		{"https://sub.example.org/", "https", "sub.example.org"},
		{"http://*.sub.example.org", "http", "*.sub.example.org"},
		{"http://*.sub.example.org/", "http", "*.sub.example.org"},
		{"https://*.sub.example.org", "https", "*.sub.example.org"},
		{"https://*.sub.example.org/", "https", "*.sub.example.org"},
		{"http://sub.sub.example.org", "http", "sub.sub.example.org"},
		{"http://sub.sub.example.org/", "http", "sub.sub.example.org"},
		{"https://sub.sub.example.org", "https", "sub.sub.example.org"},
		{"https://sub.sub.example.org/", "https", "sub.sub.example.org"},
	}

	testCasesUrlFail := []struct {
		url       string
		wantError error
	}{
		{"ftp://example.org", errInvUrl},
		{"http:/example", errInvUrl},
		{"httd://example", errInvUrl},
	}

	testCasesUserPass := []struct {
		user string
		pass string
		want Config
	}{
		{"testUser", "secretPassw0rd", Config{"testUser", "secretPassw0rd", "http", "example.org"}},
		{"t", "s", Config{"t", "s", "http", "example.org"}},
		{"1234567890", "1234567890", Config{"1234567890", "1234567890", "http", "example.org"}},
	}

	for _, tc := range testCasesUrlPass {
		t.Run(tc.url, func(t *testing.T) {
			dns, err := NewConfig("user", "password", tc.url)
			if err != nil {
				t.Errorf("err != nil: %v", err)
				return
			}

			if dns == nil {
				t.Error("dns == nil")
			}

			if tc.wantScheme != dns.scheme {
				t.Errorf("expected: %s, got %s", tc.wantScheme, dns.scheme)
			}

			if tc.wantDomain != dns.domain {
				t.Errorf("expected: %s, got %s", tc.wantDomain, dns.domain)
			}
		})
	}

	for _, tc := range testCasesUserPass {
		t.Run(fmt.Sprintf("%s:%s", tc.user, tc.pass), func(t *testing.T) {
			dns, err := NewConfig(tc.user, tc.pass, "http://example.org")
			if err != nil {
				t.Errorf("err != nil: %v", err)
				return
			}

			if dns == nil {
				t.Error("dns == nil")
			}

			if tc.want != *dns {
				t.Errorf("expected: %v, got %v", tc.want, dns)
			}
		})
	}

	_, err := NewConfig("user", "", "http://example.org")
	if err != errNoPass {
		t.Error("expected error: 'errNoPass'")
	}

	for _, tc := range testCasesUrlFail {
		t.Run(tc.url, func(t *testing.T) {
			_, err := NewConfig("user", "password", tc.url)

			if err == nil {
				t.Error("error expected")
				return
			}

			if err != tc.wantError {
				t.Errorf("expected error: %v, got %v", tc.wantError, err)
			}
		})
	}
}

func TestNewConfigNoUser(t *testing.T) {
	dns, err := NewConfig("", "", "")

	if err == nil {
		t.Errorf("err == nil, expected %v", errNoUser)
	}

	if dns != nil {
		t.Error("dns != nil")
	}
}

func TestNewConfigNoPass(t *testing.T) {
	dns, err := NewConfig("", "", "")

	if err == nil {
		t.Errorf("err == nil, expected %v", errNoPass)
	}

	if dns != nil {
		t.Error("dns != nil")
	}
}

func TestNewConfigNoUrl(t *testing.T) {
	dns, err := NewConfig("", "", "")

	if err == nil {
		t.Errorf("err == nil, expected %v", errInvUrl)
	}

	if dns != nil {
		t.Error("dns != nil")
	}
}

func TestConfig_url(t *testing.T) {
	testCases := []struct {
		cfg  Config
		want string
	}{
		{Config{"t", "s", "http", "example.org"}, "http://example.org"},
		{Config{"t", "s", "https", "example.org"}, "https://example.org"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s - %s", tc.cfg.scheme, tc.cfg.domain), func(t *testing.T) {
			if tc.want != tc.cfg.url() {
				t.Errorf("expected: %v, got %v", tc.want, tc.cfg.url())
			}
		})
	}
}
