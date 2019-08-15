package miab

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
	"strings"
	"testing"
)

var testAliasDomain1 = AliasDomain{
	Domain: "example.org",
	Aliases: Aliases{
		Alias{
			Address:          "test@example.org",
			DisplayAddress:   "test@example.org",
			ForwardsTo:       []string{"mail@example.org", "info@example.org"},
			PermittedSenders: nil,
			Required:         false,
		},
		Alias{
			Address:          "abuse@example.org",
			DisplayAddress:   "abuse@example.org",
			ForwardsTo:       []string{"mail@example.org"},
			PermittedSenders: nil,
			Required:         true,
		}},
}

var testAliasDomain2 = AliasDomain{
	Domain: "example.com",
	Aliases: Aliases{
		Alias{
			Address:          "abuse@example.com",
			DisplayAddress:   "abuse@example.com",
			ForwardsTo:       []string{"mail@example.com"},
			PermittedSenders: nil,
			Required:         true,
		}},
}

var testAliasDomains = AliasDomains{
	testAliasDomain1,
	testAliasDomain2,
}

func TestAliasDomain_Print(t *testing.T) {
	fmt.Print("\nplain:\t")
	testAliasDomain1.Print(PLAIN)
	fmt.Print("\njson:\t")
	testAliasDomain1.Print(JSON)
	fmt.Print("\nyaml:\t")
	testAliasDomain1.Print(YAML)
	fmt.Print("\ncsv :\t")
	testAliasDomain1.Print(CSV)
}

func TestAliasDomain_String(t *testing.T) {

	want := `example.org:
	test@example.org -> mail@example.org, info@example.org
	abuse@example.org -> mail@example.org`
	got := testAliasDomain1.String()

	if got != want {
		t.Errorf("wrong format,\nwant:\n***%s***\n\ngot:\n***%s***", want, got)
	}
}

func TestAliasDomain_ToString(t *testing.T) {

	var a AliasDomain

	err := json.Unmarshal([]byte(testAliasDomain1.ToString(JSON)), &a)
	if err != nil || a.Domain != "example.org" || a.Aliases[0].Address != "test@example.org" ||
		a.Aliases[0].ForwardsTo[0] != "mail@example.org" || a.Aliases[0].ForwardsTo[1] != "info@example.org" ||
		a.Aliases[0].DisplayAddress != "test@example.org" || a.Aliases[0].Required == true ||
		a.Aliases[1].Address != "abuse@example.org" || a.Aliases[1].ForwardsTo[0] != "mail@example.org" ||
		a.Aliases[1].DisplayAddress != "abuse@example.org" || a.Aliases[1].Required == false {
		t.Error("Unable to unmarshal generated json", err)
	}

	err = yaml.Unmarshal([]byte(testAliasDomain1.ToString(YAML)), &a)
	if err != nil || a.Domain != "example.org" || a.Aliases[0].Address != "test@example.org" ||
		a.Aliases[0].ForwardsTo[0] != "mail@example.org" || a.Aliases[0].ForwardsTo[1] != "info@example.org" ||
		a.Aliases[0].DisplayAddress != "test@example.org" || a.Aliases[0].Required == true ||
		a.Aliases[1].Address != "abuse@example.org" || a.Aliases[1].ForwardsTo[0] != "mail@example.org" ||
		a.Aliases[1].DisplayAddress != "abuse@example.org" || a.Aliases[1].Required == false {
		t.Error("Unable to unmarshal generated yaml", err)
	}

	want := strings.Builder{}
	want.WriteString(CsvAliasHead)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "test@example.org", "test@example.org", "mail@example.org;info@example.org", "", false`)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "abuse@example.org", "abuse@example.org", "mail@example.org", "", true`)
	want.WriteByte('\n')

	got := testAliasDomain1.ToString(CSV)
	if got != want.String() {
		t.Errorf("wrong format, want: \n+++%s+++\n\ngot:\n+++%s+++", want.String(), got)
	}
}

func TestAliasDomains_Print(t *testing.T) {
	fmt.Print("\nplain:\t")
	testAliasDomains.Print(PLAIN)
	fmt.Print("\njson:\t")
	testAliasDomains.Print(JSON)
	fmt.Print("\nyaml:\t")
	testAliasDomains.Print(YAML)
	fmt.Print("\ncsv :\t")
	testAliasDomains.Print(CSV)
}

func TestAliasDomains_String(t *testing.T) {
	want := `example.org:
	test@example.org -> mail@example.org, info@example.org
	abuse@example.org -> mail@example.org
example.com:
	abuse@example.com -> mail@example.com`

	got := testAliasDomains.String()

	if got != want {
		t.Errorf("wrong format,\nwant:\n***%s***\n\ngot:\n***%s***", want, got)
	}
}

func TestAliasDomains_ToString(t *testing.T) {

	var a AliasDomains
	err := json.Unmarshal([]byte(testAliasDomains.ToString(JSON)), &a)
	if err != nil || a[0].Domain != "example.org" || a[0].Aliases[0].Address != "test@example.org" ||
		a[0].Aliases[0].ForwardsTo[0] != "mail@example.org" || a[0].Aliases[0].ForwardsTo[1] != "info@example.org" ||
		a[0].Aliases[0].DisplayAddress != "test@example.org" || a[0].Aliases[0].Required == true ||
		a[0].Aliases[1].Address != "abuse@example.org" || a[0].Aliases[1].ForwardsTo[0] != "mail@example.org" ||
		a[0].Aliases[1].DisplayAddress != "abuse@example.org" || a[0].Aliases[1].Required == false ||
		a[1].Domain != "example.com" || a[1].Aliases[0].Address != "abuse@example.com" || a[1].Aliases[0].ForwardsTo[0] != "mail@example.com" ||
		a[1].Aliases[0].DisplayAddress != "abuse@example.com" || a[1].Aliases[0].Required == false {
		t.Error("Unable to unmarshal generated json", err)
	}

	err = yaml.Unmarshal([]byte(testAliasDomains.ToString(YAML)), &a)
	if err != nil || a[0].Domain != "example.org" || a[0].Aliases[0].Address != "test@example.org" ||
		a[0].Aliases[0].ForwardsTo[0] != "mail@example.org" || a[0].Aliases[0].ForwardsTo[1] != "info@example.org" ||
		a[0].Aliases[0].DisplayAddress != "test@example.org" || a[0].Aliases[0].Required == true ||
		a[0].Aliases[1].Address != "abuse@example.org" || a[0].Aliases[1].ForwardsTo[0] != "mail@example.org" ||
		a[0].Aliases[1].DisplayAddress != "abuse@example.org" || a[0].Aliases[1].Required == false ||
		a[1].Domain != "example.com" || a[1].Aliases[0].Address != "abuse@example.com" || a[1].Aliases[0].ForwardsTo[0] != "mail@example.com" ||
		a[1].Aliases[0].DisplayAddress != "abuse@example.com" || a[1].Aliases[0].Required == false {
		t.Error("Unable to unmarshal generated yaml", err)
	}

	want := strings.Builder{}
	want.WriteString(CsvAliasHead)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "test@example.org", "test@example.org", "mail@example.org;info@example.org", "", false`)
	want.WriteByte('\n')
	want.WriteString(`"example.org", "abuse@example.org", "abuse@example.org", "mail@example.org", "", true`)
	want.WriteByte('\n')
	want.WriteString(`"example.com", "abuse@example.com", "abuse@example.com", "mail@example.com", "", true`)
	want.WriteByte('\n')

	got := testAliasDomains.ToString(CSV)
	if got != want.String() {
		t.Errorf("wrong format, want: \n+++%s+++\n\ngot:\n+++%s+++", want.String(), got)
	}
}

func TestGetAliases(t *testing.T) {

	testCases := []struct {
		serverStatus int
		want         AliasDomains
		wantError    bool
	}{
		{200, testAliasDomains, false},
		{503, nil, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("GetUsers %d", tc.serverStatus), func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodGet, tc.serverStatus, testAliasDomains.ToString(JSON), NONE, false, "")
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)
			got, err := GetAliases(c)

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

func TestAddAlias(t *testing.T) {

	testCases := []struct {
		email        string
		forwards     string
		serverStatus int
		wantError    bool
	}{
		{"user@example.org", "user2@example.org", 200, false},
		{"user@example.org", "user2@example.org", 503, true},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodPost, tc.serverStatus, "", NONE, false, fmt.Sprintf("address=%s&forwards_to=%s", tc.email, tc.forwards))
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)
			err := AddAlias(c, tc.email, tc.forwards)

			if tc.wantError && err == nil {
				t.Errorf("failed, want error, got: nil")

			} else if !tc.wantError && err != nil {
				t.Errorf("failed, got error: %v", err)
			}
		})
	}
}

func TestRemoveAlias(t *testing.T) {
	testCases := []struct {
		email        string
		serverStatus int
		wantError    bool
	}{
		{"user@example.org",  200, false},
		{"user@example.org", 503, true},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			ts := getDnsTestServer(t, http.MethodPost, tc.serverStatus, "", NONE, false, fmt.Sprintf("address=%s", tc.email))
			defer ts.Close()
			c, _ := NewConfig("test", "secret", ts.URL)
			err := RemoveAlias(c, tc.email)

			if tc.wantError && err == nil {
				t.Errorf("failed, want error, got: nil")

			} else if !tc.wantError && err != nil {
				t.Errorf("failed, got error: %v", err)
			}
		})
	}
}
