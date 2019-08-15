package miab

import (
	"encoding/json"
	"errors"
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

type AliasDomains []AliasDomain

func (a AliasDomains) Print(format Formats) {
	print(a, format)
}

func (a AliasDomains) ToString(format Formats) string {
	s, err := toString(a, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

func (a AliasDomains) String() string {
	r := strings.Builder{}
	for i, x := range a {
		r.WriteString(x.String())
		if i < len(a) - 1 {
			r.WriteByte('\n')
		}
	}
	return r.String()
}


type Aliases []Alias

type AliasDomain struct {
	Domain  string  `json:"domain"`
	Aliases Aliases `json:"aliases"`
}

func (a AliasDomain) String() string {

	r := strings.Builder{}
	r.WriteString(a.Domain)
	r.WriteString(":\n")
	for i, x := range a.Aliases {
		r.WriteString(fmt.Sprintf("\t%s -> %s", x.Address, strings.Join(x.ForwardsTo, ", ")))
		if i < len(a.Aliases) - 1 {
			r.WriteByte('\n')
		}
	}
	return r.String()
}

func (a AliasDomain) Print(format Formats) {
	print(a, format)
}

func (a AliasDomain) ToString(format Formats) string {
	s, err := toString(a, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

type Alias struct {
	Address          string   `json:"address"`
	DisplayAddress   string   `json:"address_display"`
	ForwardsTo       []string `json:"forwards_to"`
	PermittedSenders []string `json:"permitted_senders"`
	Required         bool     `json:"required"`
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
			return errors.New(fmt.Sprintf("response error (%d): %s", res.StatusCode, bodyString))
		} else {
			return errors.New(fmt.Sprintf("response error (%d)", res.StatusCode))
		}
	}
	return nil
}

// GetUsers returns a list of existing mail users.
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
			return nil, errors.New(fmt.Sprintf("response error (%d): %s", res.StatusCode, bodyString))
		} else {
			return nil, errors.New(fmt.Sprintf("response error (%d)", res.StatusCode))
		}
	}

	var result AliasDomains
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Adds a new mail user. Note: Adding a mail user with an unknown domain adds this domain to the server.
// The parameter `forwardsTo` can be a comma separated list of addresses.
func AddAlias(c *Config, address, forwardsTo string) error {

	body := fmt.Sprintf("address=%s&forwards_to=%s", address, forwardsTo)
	return exeAlias(c, "add", body)
}

// Removes a new mail user.
func RemoveAlias(c *Config, address string) error {

	body := fmt.Sprintf("address=%s", address)
	return exeAlias(c, "remove", body)
}
