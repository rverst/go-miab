package miab

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
	"strings"
	"testing"
)

var testMailDomain1 = MailDomain{
	Domain: "example.org",
	Users: Users{
		User{
			Email:      "admin@example.org",
			Privileges: []interface{}{"admin"},
			Status:     Active,
			Mailbox:    "",
		},
		User{
			Email:      "user1@example.org",
			Privileges: nil,
			Status:     Active,
			Mailbox:    "",
		},
		User{
			Email:      "user2@example.org",
			Privileges: nil,
			Status:     Archived,
			Mailbox:    "/home/miab/mail/example.org/user2",
		},
	},
}

var testMailDomain2 = MailDomain{
	Domain: "example.com",
	Users: Users{
		User{
			Email:      "admin@example.com",
			Privileges: []interface{}{"admin"},
			Status:     Active,
			Mailbox:    "",
		},
	},
}

var testMailDomains = MailDomains{
	testMailDomain1,
	testMailDomain2,
}

func TestMailDomain_String(t *testing.T) {
	want := `example.org:
	admin@example.org
	user1@example.org`
	got := testMailDomain1.String()

	if got != want {
		t.Errorf("wrong format,\nwant:\n***%s***\n\ngot:\n***%s***", want, got)
	}
}

func TestMailDomain_ToString(t *testing.T) {

	var m MailDomain
	err := json.Unmarshal([]byte(testMailDomain1.ToString(JSON)), &m)
	if err != nil || m.Domain != "example.org" || m.Users[0].Email != "admin@example.org" || m.Users[0].Status != Active ||
		m.Users[0].Mailbox != "" || m.Users[1].Email != "user1@example.org" || m.Users[1].Status != Active ||
		m.Users[1].Mailbox != "" || m.Users[2].Email != "user2@example.org" || m.Users[2].Status != Archived ||
		m.Users[2].Mailbox != "/home/miab/mail/example.org/user2" {
		t.Error("Unable to unmarshal generated json", err)
	}

	err = yaml.Unmarshal([]byte(testMailDomain1.ToString(YAML)), &m)
	if err != nil || m.Domain != "example.org" || m.Users[0].Email != "admin@example.org" || m.Users[0].Status != Active ||
		m.Users[0].Mailbox != "" || m.Users[1].Email != "user1@example.org" || m.Users[1].Status != Active ||
		m.Users[1].Mailbox != "" || m.Users[2].Email != "user2@example.org" || m.Users[2].Status != Archived ||
		m.Users[2].Mailbox != "/home/miab/mail/example.org/user2" {
		t.Error("Unable to unmarshal generated yaml", err)
	}

	want := strings.Builder{}
	want.WriteString(csvUserHead)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "admin@example.org", "admin", "active", ""`)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "user1@example.org", "", "active", ""`)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "user2@example.org", "", "inactive", "/home/miab/mail/example.org/user2"`)
	want.WriteByte('\n')

	got := testMailDomain1.ToString(CSV)
	if got != want.String() {
		t.Errorf("wrong format, want: \n+++%s+++\n\ngot:\n+++%s+++", want.String(), got)
	}
}

func TestMailDomains_String(t *testing.T) {
	want := `example.org:
	admin@example.org
	user1@example.org
example.com:
	admin@example.com`
	got := testMailDomains.String()

	if got != want {
		t.Errorf("wrong format,\nwant:\n***%s***\n\ngot:\n***%s***", want, got)
	}
}

func TestMailDomains_ToString(t *testing.T) {

	var m MailDomains
	err := json.Unmarshal([]byte(testMailDomains.ToString(JSON)), &m)
	if err != nil || m[0].Domain != "example.org" || m[0].Users[0].Email != "admin@example.org" || m[0].Users[0].Status != Active ||
		m[0].Users[0].Mailbox != "" || m[0].Users[1].Email != "user1@example.org" || m[0].Users[1].Status != Active ||
		m[0].Users[1].Mailbox != "" || m[0].Users[2].Email != "user2@example.org" || m[0].Users[2].Status != Archived ||
		m[0].Users[2].Mailbox != "/home/miab/mail/example.org/user2" ||
		m[1].Domain != "example.com" || m[1].Users[0].Email != "admin@example.com" || m[1].Users[0].Status != Active ||
		m[1].Users[0].Mailbox != "" {
		t.Error("Unable to unmarshal generated json", err)
	}

	err = yaml.Unmarshal([]byte(testMailDomains.ToString(YAML)), &m)
	if err != nil || m[0].Domain != "example.org" || m[0].Users[0].Email != "admin@example.org" || m[0].Users[0].Status != Active ||
		m[0].Users[0].Mailbox != "" || m[0].Users[1].Email != "user1@example.org" || m[0].Users[1].Status != Active ||
		m[0].Users[1].Mailbox != "" || m[0].Users[2].Email != "user2@example.org" || m[0].Users[2].Status != Archived ||
		m[0].Users[2].Mailbox != "/home/miab/mail/example.org/user2" ||
		m[1].Domain != "example.com" || m[1].Users[0].Email != "admin@example.com" || m[1].Users[0].Status != Active ||
		m[1].Users[0].Mailbox != "" {
		t.Error("Unable to unmarshal generated yaml", err)
	}

	want := strings.Builder{}
	want.WriteString(csvUserHead)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "admin@example.org", "admin", "active", ""`)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "user1@example.org", "", "active", ""`)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "user2@example.org", "", "inactive", "/home/miab/mail/example.org/user2"`)
	want.WriteByte('\n')
	want.WriteString(`"example.com", "admin@example.com", "admin", "active", ""`)
	want.WriteByte('\n')

	got := testMailDomains.ToString(CSV)
	if got != want.String() {
		t.Errorf("wrong format, want: \n+++%s+++\n\ngot:\n+++%s+++", want.String(), got)
	}
}

func TestGetUsers(t *testing.T) {

	testCases := []struct {
		serverStatus int
		want         MailDomains
		wantError    bool
	}{
		{200, testMailDomains, false},
		{503, nil, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("GetUsers %d", tc.serverStatus), func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodGet, tc.serverStatus, testMailDomains.ToString(JSON), NONE, false, "")
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)
			got, err := GetUsers(c)

			if tc.wantError && err == nil {
				t.Errorf("failed, want error, got: nil")

			} else if !tc.wantError && err != nil {
				t.Errorf("failed, got error: %v", err)
			}

			if tc.want == nil && got != nil {
				t.Errorf("failed, want nil, got: %v", got)
			} else if tc.want != nil && got == nil {
				t.Errorf("failed, want %v, got: nil", tc.want)
			} else if got != nil && (tc.want.ToString(YAML) != got.ToString(YAML)) {
				t.Errorf("failed, want %v\ngot: %v", tc.want, got)
			}
		})
	}
}

func TestAddUser(t *testing.T) {

	testCases := []struct {
		email        string
		pass         string
		serverStatus int
		wantError    bool
	}{
		{"user@example.org", "supersecret", 200, false},
		{"user@example.org", "supersecret", 503, true},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodPost, tc.serverStatus, "", NONE, false, fmt.Sprintf("email=%s&password=%s", tc.email, tc.pass))
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)
			err := AddUser(c, tc.email, tc.pass)

			if tc.wantError && err == nil {
				t.Errorf("failed, want error, got: nil")

			} else if !tc.wantError && err != nil {
				t.Errorf("failed, got error: %v", err)
			}
		})
	}
}

func TestRemoveUser(t *testing.T) {
	testCases := []struct {
		email        string
		serverStatus int
		wantError    bool
	}{
		{"user@example.org", 200, false},
		{"user@example.org", 503, true},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodPost, tc.serverStatus, "", NONE, false, fmt.Sprintf("email=%s", tc.email))
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)
			err := DeleteUser(c, tc.email)

			if tc.wantError && err == nil {
				t.Errorf("failed, want error, got: nil")

			} else if !tc.wantError && err != nil {
				t.Errorf("failed, got error: %v", err)
			}
		})
	}
}

func TestAddPrivileges(t *testing.T) {
	testCases := []struct {
		email        string
		serverStatus int
		wantError    bool
	}{
		{"user@example.org", 200, false},
		{"user@example.org", 503, true},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodPost, tc.serverStatus, "", NONE, false, fmt.Sprintf("email=%s&privilege=admin", tc.email))
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)
			err := AddPrivileges(c, tc.email)

			if tc.wantError && err == nil {
				t.Errorf("failed, want error, got: nil")

			} else if !tc.wantError && err != nil {
				t.Errorf("failed, got error: %v", err)
			}
		})
	}
}

func TestRemovePrivileges(t *testing.T) {
	testCases := []struct {
		email        string
		serverStatus int
		wantError    bool
	}{
		{"user@example.org", 200, false},
		{"user@example.org", 503, true},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodPost, tc.serverStatus, "", NONE, false, fmt.Sprintf("email=%s&privilege=admin", tc.email))
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)
			err := RemovePrivileges(c, tc.email)

			if tc.wantError && err == nil {
				t.Errorf("failed, want error, got: nil")

			} else if !tc.wantError && err != nil {
				t.Errorf("failed, got error: %v", err)
			}
		})
	}
}
