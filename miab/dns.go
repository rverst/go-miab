package miab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// ResourceType defines a dns resource type (e.g. 'A', 'AAAA', TXT...)
type ResourceType string

// NetworkType defines a valid network type ('tcp4' or 'tcp6')
type NetworkType string

var (
	regexQname     = *regexp.MustCompile(`^\*?\.?(?:[a-zA-Z0-9-]{2,63}\.?)+\.([a-zA-Z]{2,})$`)
	errInvQname    = errors.New("'qname' seems to be invalid")
	errInvNet      = errors.New("'network' has to be `tcp4` or `tcp6`")
	errRtypeNotSet = errors.New("'rtype' has to be set")
)

// NONE means no resource type specified
const NONE = ResourceType(``)

// A - IPv4 address record (RFC 1035)
const A = ResourceType(`A`)

// AAAA - IPv4 address record (RFC 3596)
const AAAA = ResourceType(`AAAA`)

// CAA - certification authority authorization (RFC 6844)
const CAA = ResourceType(`CAA`)

// CNAME - canonical name record (RFC 1035)
const CNAME = ResourceType(`CNAME`)

// MX - mail exchange record (RFC 1035 and RFC 7505)
const MX = ResourceType(`MX`)

// NS - name server record (RFC 1035)
const NS = ResourceType(`NS`)

// SRV - service locator (RFC 2782)
const SRV = ResourceType(`SRV`)

// SSHFP - SSH public key fingerprint (RFC 4255)
const SSHFP = ResourceType(`SSHFP`)

// TXT - text record (RFC 1035)
const TXT = ResourceType(`TXT`)

// TCP4 - transport via TCP/IPv4
const TCP4 = NetworkType(`tcp4`)

// TCP6 - transport via TCP/IPv6
const TCP6 = NetworkType(`tcp6`)

var allResourceTypes = []ResourceType{A, AAAA, TXT, CNAME, MX, SRV, SSHFP, CAA, NS}

// IsValid checks if the ResourceType is valid (supported by the Mail-in-a-Box API).
func (r *ResourceType) IsValid() bool {

	for _, t := range allResourceTypes {
		if (*r) == t {
			return true
		}
	}
	return false
}

// ParseDnsResource parses the provided string to a ResourceType if possible.
func ParseDnsResource(value string) (ResourceType, error) {

	var rtype = ResourceType(strings.ToUpper(value))
	if rtype.IsValid() {
		return rtype, nil
	}

	return NONE, fmt.Errorf("'%s' is not a valid resource type", value)
}

func dnsPath(qname string, rtype ResourceType) string {

	const path = "admin/dns/custom"
	if qname == "" {
		return path
	}

	if rtype == NONE {
		return fmt.Sprintf("%s/%s", path, qname)
	}

	return fmt.Sprintf("%s/%s/%s", path, qname, rtype)
}

// Records defines an array of Record
type Records []Record

// ToString returns a string of the Records in the provided Format
func (r Records) ToString(format Format) string {
	s, err := toString(r, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

// String returns a string representation of the Records
func (r Records) String() string {
	res := strings.Builder{}
	for i, x := range r {
		res.WriteString(x.String())
		if i < len(r)-1 {
			res.WriteByte('\n')
		}
	}
	return res.String()
}

// Record defines a dns record
type Record struct {
	QName string       `json:"qname"` // QName holds the fully qualified domain name of the record
	RType ResourceType `json:"rtype"` // RType holds the ResourceType of the record
	Value string       `json:"value"` // Value holds the value of the record
}

// ToString returns a string of the Record in the provided Format
func (r Record) ToString(format Format) string {
	s, err := toString(r, format)
	if err != nil {
		fmt.Println("unexpected error", err)
		os.Exit(1)
	}
	return s
}

// String returns a string representation of the Record
func (r Record) String() string {
	return fmt.Sprintf("%s\t%s\t%s", r.QName, r.RType, r.Value)
}

func execDns(c *Config, method, qname string, rtype ResourceType, value string) (bool, error) {

	if !regexQname.MatchString(qname) {
		return false, errInvQname
	}

	if !rtype.IsValid() {
		return false, errRtypeNotSet
	}

	client := &http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.url(), dnsPath(qname, rtype)), strings.NewReader(value))
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(c.user, c.password)
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	bodyString := string(bodyBytes)

	if res.StatusCode != 200 {
		if len(bodyString) > 0 {
			return false, fmt.Errorf("response error (%d): %s", res.StatusCode, bodyString)
		}
		return false, fmt.Errorf("response error (%d)", res.StatusCode)
	}
	if strings.HasPrefix(bodyString, "updated DNS:") {
		return true, nil
	}
	return false, fmt.Errorf("unexpected response body: %s", bodyString)
}

// GetDns returns matching custom DNS records. The optional qname and rtype parameters
// filter the records returned. NOTE: Due to a weired behavior in the Mail-in-a-Box api, if the qname is given
// and the rtype not (NONE), the rtype defaults to A records.
func GetDns(c *Config, qname string, rtype ResourceType) (Records, error) {

	if qname != "" && !regexQname.MatchString(qname) {
		return nil, errInvQname
	}

	client := &http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.url(), dnsPath(qname, rtype)), nil)
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

	var result Records
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetDns sets a custom DNS record replacing any existing records with the same qname and rtype.
// Use SetDns (instead of AddDns) when you only have one value for a qname and rtype,
// such as typical A records (without round-robin).
// Returns true if the DNS was updated
func SetDns(c *Config, qname string, rtype ResourceType, value string) (bool, error) {

	return execDns(c, http.MethodPut, qname, rtype, value)
}

// AddDns adds a new custom DNS record. Use AddDns when you have multiple TXT records or round-robin A records.
// Returns true if the DNS was updated
func AddDns(c *Config, qname string, rtype ResourceType, value string) (bool, error) {

	return execDns(c, http.MethodPost, qname, rtype, value)
}

// DeleteDns removes custom DNS records. If the value empty, deletes all records matching the qname and rtype.
// If the value is present, deletes only the record matching the qname, rtype and value.
// Returns true if the DNS was updated
func DeleteDns(c *Config, qname string, rtype ResourceType, value string) (bool, error) {

	return execDns(c, http.MethodDelete, qname, rtype, value)
}

// SetOrAddAddressRecord sets or adds a custom A or AAAA record of the qname. If the value is empty, the server
// will take the IPv4 or IPv6 address of the remote host as the value - quite handy for dynamic DNS!
// You have to explicitly set network to `tcp4` or `tcp6` to set the correct record!
// Consider using UpdateDns4 or UpdateDns6 for dynamic DNS!
// Returns true if the DNS was set or updated.
func SetOrAddAddressRecord(c *Config, network NetworkType, qname, value string, add bool) (bool, error) {

	if !regexQname.MatchString(qname) {
		return false, errInvQname
	}

	if network != TCP4 && network != TCP6 {
		return false, errInvNet
	}

	dialer := &net.Dialer{
		Timeout:   time.Second * 30,
		KeepAlive: 0,
		DualStack: false,
	}

	tr := &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}}

	rtype := A
	if network == TCP6 {
		rtype = AAAA
	}
	client := &http.Client{Transport: tr}

	method := http.MethodPut
	if add {
		method = http.MethodPost
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.url(), dnsPath(qname, rtype)), strings.NewReader(value))
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(c.user, c.password)
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	bodyString := string(bodyBytes)

	if res.StatusCode != 200 {
		if len(bodyString) > 0 {
			return false, fmt.Errorf("response error (%d): %s", res.StatusCode, bodyString)
		}
		return false, fmt.Errorf("response error (%d)", res.StatusCode)
	}
	if bodyString == "OK" || strings.HasPrefix(bodyString, "updated DNS:") {
		return true, nil
	}
	return false, fmt.Errorf("unexpected response body: %s", bodyString)
}

// UpdateDns4 updates a custom A record for the provided qname. If the value is empty, the server will take the
// IPv4 address of the remote host as the value - quite handy for dynamic DNS!
// Returns true if the DNS was updated
func UpdateDns4(c *Config, qname, value string) (bool, error) {
	return SetOrAddAddressRecord(c, "tcp4", qname, value, false)
}

// UpdateDns6 updates a custom AAAA record for the provided qname. If the value is empty, the server will take the
// IPv6 address of the remote host as the value - quite handy for dynamic DNS!
// Returns true if the DNS was updated
func UpdateDns6(c *Config, qname, value string) (bool, error) {
	return SetOrAddAddressRecord(c, "tcp6", qname, value, false)
}
