package miab

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	regexUrl  = *regexp.MustCompile(`^(?P<schema>https?)://(?P<domain>.+)$`)
	errNoUser = errors.New("'user' not specified")
	errNoPass = errors.New("'password' not specified")
	errInvUrl = errors.New("'url' is not valid")
)

// Config holds the details to communicate with the Mail-in-a-Box API.
type Config struct {
	user     string
	password string
	scheme   string
	domain   string
}

// NewConfig creates a new configuration to access the Mail-in-a-Box API.
func NewConfig(user, password, url string) (*Config, error) {
	if user == "" {
		return nil, errNoUser
	}

	if password == "" {
		return nil, errNoPass
	}

	tUrl := strings.ToLower(strings.Trim(url, ` `))
	res := regexUrl.FindAllStringSubmatch(tUrl, -1)

	if res == nil || len(res) != 1 || len(res[0]) != 3 {
		return nil, errInvUrl
	}

	return &Config{
		user:     user,
		password: password,
		scheme:   res[0][1],
		domain:   strings.TrimRight(res[0][2], `/`),
	}, nil
}

func (c *Config) url() string {
	return fmt.Sprintf("%s://%s", c.scheme, c.domain)
}
