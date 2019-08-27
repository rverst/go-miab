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

// Status describes the status of an e-mail account.
type Status string

const (
	Active   = Status("active")   // Active describes an active e-mail account.
	Archived = Status("inactive") // Archived describes an archived (inactive) e-mail account.
)

const (
	usersPath = `admin/mail/users`
)

// MailDomains defines an array of MailDomain.
type MailDomains []MailDomain

// String returns a string representation of the MailDomains.
func (m MailDomains) String() string {
	r := strings.Builder{}
	for i, x := range m {
		r.WriteString(x.String())
		if i < len(m)-1 {
			r.WriteString("\n")
		}
	}
	return r.String()
}

// ToString returns a string representation of the MailDomains in the provided Format.
func (m MailDomains) ToString(format Format) string {
	s, err := toString(m, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

// Users defines an array of User
type Users []User

// MailDomain defines a domain with its Users
type MailDomain struct {
	Domain string `json:"domain"` // Domain is the domain name, e.g. example.org.
	Users  Users  `json:"users"`  // Users is a list of User of the domain.
}

type User struct {
	Email      string      `json:"email"`      // Email is the e-mail address.
	Privileges interface{} `json:"privileges"` // Privileges is a list of privileges, given to the user. Note: due to a bug in Mail-in-a-Box < v0.42, we have to use an generic interface, because the datatype differs in Archived users (string instead of array).
	Status     Status      `json:"Status"`     // Status is the status of the account (Active or Archived).
	Mailbox    string      `json:"mailbox"`    // Mailbox is the path to the mailbox on the server (only for archived accounts).
}

// String returns a string representation of the MailDomain.
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

// ToString returns a string representation of the MailDomain in the provided Format.
func (m MailDomain) ToString(format Format) string {
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

// GetUsers returns a list of existing e-mail users.
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

// AddUser adds a new e-mail user. Note: Adding an e-mail user with an unknown domain adds this domain also to the server.
func AddUser(c *Config, email, password string) error {
	body := fmt.Sprintf("email=%s&password=%s", email, password)
	return execUser(c, "add", body)
}

// DeleteUser removes an existing e-mail user.
func DeleteUser(c *Config, email string) error {
	body := fmt.Sprintf("email=%s", email)
	return execUser(c, "remove", body)
}

// AddPrivileges adds admin privileges to this user.
func AddPrivileges(c *Config, email string) error {
	body := fmt.Sprintf("email=%s&privilege=admin", email)
	return execUser(c, "privileges/add", body)
}

// RemovePrivileges removes the admin privileges from this user.
func RemovePrivileges(c *Config, email string) error {
	body := fmt.Sprintf("email=%s&privilege=admin", email)
	return execUser(c, "privileges/remove", body)
}
