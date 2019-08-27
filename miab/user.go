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

type Privilege string
type Status string

const (
	Admin    = Privilege("admin")
	Active   = Status("active")
	Archived = Status("inactive")
)

const (
	usersPath = `admin/mail/users`
)

type MailDomains []MailDomain

func (m MailDomains) String() string {
	r := strings.Builder{}
	for i, x := range m {
		r.WriteString(x.String())
		if i < len(m) - 1 {
			r.WriteString("\n")
		}
	}
	return r.String()
}

func (m MailDomains) ToString(format Formats) string {
	s, err := toString(m, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

type Users []User

type MailDomain struct {
	Domain string `json:"domain"`
	Users  Users  `json:"users"`
}

type User struct {
	Email      string      `json:"email"`
	Privileges interface{} `json:"privileges"` //due to a bug in miab < v0.42, we have to use an generic interface, because the datatype differs in archived users (string instead of array)
	Status     Status      `json:"status"`
	Mailbox    string      `json:"mailbox"`
}

func (m MailDomain) String() string {

	r := strings.Builder{}
	r.WriteString(m.Domain)
	r.WriteString(":\n")
	for i, u := range m.Users {
		if u.Status != Active {
			continue
		}
		if i > 0 {
			r.WriteByte('\n')
		}
		r.WriteString(fmt.Sprintf("\t%s", u.Email))
	}
	return r.String()
}

func (m MailDomain) ToString(format Formats) string {
	s, err := toString(m, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

func execUser(c *Config, path, body string) error {

	client := &http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/%s", c.url(), usersPath, path), strings.NewReader(body))
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
func GetUsers(c *Config) (MailDomains, error) {

	client := &http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s?format=json", c.url(), usersPath), nil)
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

	var result MailDomains
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Adds a new mail user. Note: Adding a mail user with an unknown domain adds this domain to the server.
func AddUser(c *Config, email, password string) error {

	body := fmt.Sprintf("email=%s&password=%s", email, password)
	return execUser(c, "add", body)
}

// Removes a new mail user.
func RemoveUser(c *Config, email string) error {

	body := fmt.Sprintf("email=%s", email)
	return execUser(c, "remove", body)
}

// Adds admin privileges to this user.
func AddPrivileges(c *Config, email string) error {

	body := fmt.Sprintf("email=%s&privilege=admin", email)
	return execUser(c, "privileges/add", body)
}

// Removes admin privileges to this user.
func RemovePrivileges(c *Config, email string) error {

	body := fmt.Sprintf("email=%s&privilege=admin", email)
	return execUser(c, "privileges/remove", body)
}
