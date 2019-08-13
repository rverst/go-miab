package miab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ResourceType string
type NetworkType string

var (
	regexQname     = *regexp.MustCompile(`^\*?\.?(?:[a-zA-Z0-9-]{2,63}\.?)+\.([a-zA-Z]{2,})$`)
	errInvQname    = errors.New("'qname' seems to be invalid")
	errInvNet      = errors.New("'network' has to be `tcp4` or `tcp6`")
	errRtypeNotSet = errors.New("'rtype' has to be set")
)

const (
	NONE  = ResourceType(``)
	A     = ResourceType(`A`)
	AAAA  = ResourceType(`AAAA`)
	TXT   = ResourceType(`TXT`)
	CNAME = ResourceType(`CNAME`)
	MX    = ResourceType(`MX`)
	SRV   = ResourceType(`SRV`)
	SSHFP = ResourceType(`SSHFP`)
	CAA   = ResourceType(`CAA`)
	NS    = ResourceType(`NS`)
)

const  (
	TCP4 = NetworkType(`tcp4`)
	TCP6 = NetworkType(`tcp6`)
)

var AllResourceTypes = []ResourceType{A, AAAA, TXT, CNAME, MX, SRV, SSHFP, CAA, NS}

func (r *ResourceType) IsValid() bool {

	for _, t := range AllResourceTypes {
		if (*r) == t {
			return true
		}
	}
	return false
}

func ParseDnsResource(value string) (ResourceType, error) {

	var rtype = ResourceType(strings.ToUpper(value))
	if rtype.IsValid() {
		return rtype, nil
	}

	return NONE, errors.New(fmt.Sprintf("'%s' is not a valid resource type", value))
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

type Records []Record

func (r Records) Print(format Formats) {
	print(r, format)
}

func (r Records) ToString(format Formats) string {
	return toString(r, format)
}

func (r Records) String() string {
	res := strings.Builder{}
	for _, x := range r {
		res.WriteString(x.String())
		res.WriteString("\n")
	}
	return res.String()
}

type Record struct {
	QName string       `json:"qname"`
	RType ResourceType `json:"rtype"`
	Value string       `json:"value"`
}

func (r Record) Print(format Formats) {
	print(r, format)
}

func (r Record) ToString(format Formats) string {
	return toString(r, format)
}

func (r Record) String() string {
	return fmt.Sprintf("%s\t%s\t%s", r.QName, r.RType, r.Value)
}

func execDns(c *Config, method, qname string, rtype ResourceType, value string) (bool, error) {

	if !regexQname.MatchString(qname) {
		return false, errInvQname
	}

	if rtype == NONE {
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

	if res.StatusCode != 200 {
		return false, errors.New(fmt.Sprintf("response error: %s (%d)", res.Status, res.StatusCode))
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(string(bodyBytes), "updated DNS:"), nil
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
		return nil, errors.New(fmt.Sprintf("response error: %s (%d)", res.Status, res.StatusCode))
	}

	var result Records
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Sets a custom DNS record replacing any existing records with the same qname and rtype.
// Use SetDns (instead of AddDns) when you only have one value for a qname and rtype,
// such as typical A records (without round-robin).
// Returns true if the DNS was updated
func SetDns(c *Config, qname string, rtype ResourceType, value string) (bool, error) {

	return execDns(c, http.MethodPut, qname, rtype, value)
}

// Adds a new custom DNS recorc. Use AddDns when you have multiple TXT records or round-robin A records.
// Returns true if the DNS was updated
func AddDns(c *Config, qname string, rtype ResourceType, value string) (bool, error) {

	return execDns(c, http.MethodPost, qname, rtype, value)
}

// Deletes custom DNS records. If the value empty, deletes all records matching the qname and rtype.
// If the value is present, deletes only the record matching the qname, rtype and value.
// Returns true if the DNS was updated
func DeleteDns(c *Config, qname string, rtype ResourceType, value string) (bool, error) {

	return execDns(c, http.MethodDelete, qname, rtype, value)
}

// Sets or Adds a custom A or AAAA record of the qname. If the value is empty, the server will take the
// IPv4 or IPv6 address of the remote host as the value - quite handy for dynamic DNS!
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

	if res.StatusCode != 200 {
		return false, errors.New(fmt.Sprintf("response error: %s (%d)", res.Status, res.StatusCode))
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(string(bodyBytes), "updated DNS:"), nil
}

// Updates a custom A record of the qname. If the value is empty, the server will take the
// IPv4 address of the remote host as the value - quite handy for dynamic DNS!
// Returns true if the DNS was updated
func UpdateDns4(c *Config, qname, value string) (bool, error) {
	return SetOrAddAddressRecord(c, "tcp4", qname, value, false)
}

// Updates a custom AAAA record of the qname. If the value is empty, the server will take the
// IPv6 address of the remote host as the value - quite handy for dynamic DNS!
// Returns true if the DNS was updated
func UpdateDns6(c *Config, qname, value string) (bool, error) {
	return SetOrAddAddressRecord(c, "tcp6", qname, value, false)
}
