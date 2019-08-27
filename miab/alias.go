package miab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	aliasPath = `admin/mail/aliases`
)

// AliasDomains defines an array of AliasDomain.
type AliasDomains []AliasDomain

// String returns a string representation of the AliasDomains.
func (a AliasDomains) String() string {
	r := strings.Builder{}
	for i, x := range a {
		r.WriteString(x.String())
		if i < len(a)-1 {
			r.WriteByte('\n')
		}
	}
	return r.String()
}

// ToString returns a string of the AliasDomains in the provided Format.
func (a AliasDomains) ToString(format Format) string {
	s, err := toString(a, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

// Aliases defines an array of Alias
type Aliases []Alias

// AliasDomain defines a domain with its aliases
type AliasDomain struct {
	Domain  string  `json:"domain"`  // Domain is the domain name, e.g. example.org.
	Aliases Aliases `json:"aliases"` // Aliases is a list of Alias of the domain.
}

// String returns a string representation of the AliasDomains.
func (a AliasDomain) String() string {

	r := strings.Builder{}
	r.WriteString(a.Domain)
	r.WriteString(":\n")
	for i, x := range a.Aliases {
		r.WriteString(fmt.Sprintf("\t%s -> %s", x.Address, strings.Join(x.ForwardsTo, ", ")))
		if i < len(a.Aliases)-1 {
			r.WriteByte('\n')
		}
	}
	return r.String()
}

// ToString returns a string of the AliasDomains in the provided Format.
func (a AliasDomain) ToString(format Format) string {
	s, err := toString(a, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

// Alias defines an e-mail alias
type Alias struct {
	Address          string   `json:"address"`           // Address is the alias address.
	DisplayAddress   string   `json:"address_display"`   // DisplayAddress is the display address of the alias.
	ForwardsTo       []string `json:"forwards_to"`       // ForwardsTo is a comma separated list of e-mail addresses to which the alias should forward to.
	PermittedSenders []string `json:"permitted_senders"` // PermittedSenders is a comma separated list of e-mail addresses which users can send in the name of the alias, defaults to ForwardsTo.
	Required         bool     `json:"required"`          // Required describes if the alias is required by the Mail-in-a-Box Server (e.g. abuse@<domain> is required and can't be deleted).
}

func exeAlias(c *Config, path, body string) error {

	client := &http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/%s", c.url(), aliasPath, path), strings.NewReader(body))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.user, c.password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)
		if len(bodyString) > 0 {
			return fmt.Errorf("response error (%d): %s", res.StatusCode, bodyString)
		}
		return fmt.Errorf("response error (%d)", res.StatusCode)
	}
	return nil
}

// GetAliases returns a list of existing e-mail aliases.
func GetAliases(c *Config) (AliasDomains, error) {

	client := &http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s?format=json", c.url(), aliasPath), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		bodyString := string(bodyBytes)
		if len(bodyString) > 0 {
			return nil, fmt.Errorf("response error (%d): %s", res.StatusCode, bodyString)
		}
		return nil, fmt.Errorf("response error (%d)", res.StatusCode)
	}

	var result AliasDomains
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// AddAlias adds a new alias.
// The parameter `forwardsTo` can be a comma separated list of addresses.
func AddAlias(c *Config, address, forwardsTo string) error {

	body := fmt.Sprintf("address=%s&forwards_to=%s", address, forwardsTo)
	return exeAlias(c, "add", body)
}

// DeleteAlias removes an alias.
func DeleteAlias(c *Config, address string) error {

	body := fmt.Sprintf("address=%s", address)
	return exeAlias(c, "remove", body)
}
