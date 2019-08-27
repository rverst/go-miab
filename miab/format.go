package miab

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

// Format defines an output format (e.g. `json`, `yaml`, ...)
type Format string

const (
	csvDnsHead   = `"domain name", "record type", "value"`
	csvUserHead  = `"domain", "email", "privileges", "Status", "mailbox"`
	csvAliasHead = `"domain", address", "displayAddress", "forwardsTo", "permittedSenders", "required"`
)

const JSON = Format(`json`)   // JSON - output in json format
const YAML = Format(`yaml`)   // YAML - output in yaml format
const CSV = Format(`csv`)     // CSV - output in csv format, comma separated
const PLAIN = Format(`plain`) // PLAIN - output as plain text

func toString(i interface{}, format Format) (string, error) {
	switch format {
	case JSON:
		return marshallJson(i)
	case YAML:
		return marshallYaml(i)
	case CSV:
		return marshallCsv(i)
	default:
		switch i.(type) {
		case AliasDomains:
			return i.(AliasDomains).String(), nil
		case AliasDomain:
			return i.(AliasDomain).String(), nil
		case Records:
			return i.(Records).String(), nil
		case Record:
			return i.(Record).String(), nil
		case MailDomains:
			return i.(MailDomains).String(), nil
		case MailDomain:
			return i.(MailDomain).String(), nil
		default:
			return fmt.Sprint(i), nil
		}
	}
}

func marshallJson(i interface{}) (string, error) {
	r, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(r), nil
}

func marshallYaml(i interface{}) (string, error) {
	r, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("Error marshalling yaml: %v\n", err)
		os.Exit(1)
	}
	return string(r), nil
}

func marshallCsv(i interface{}) (string, error) {

	r := strings.Builder{}
	switch i.(type) {
	case AliasDomains, AliasDomain:
		r.WriteString(csvAliasHead)
	case MailDomains, MailDomain:
		r.WriteString(csvUserHead)
	case Records, Record:
		r.WriteString(csvDnsHead)
	default:
		return "", errors.New(fmt.Sprintf("unsupported type"))
	}

	r.WriteByte('\n')

	switch i.(type) {
	case AliasDomains:
		for _, a := range i.(AliasDomains) {
			csvAliasDomain(a, &r)
		}
	case AliasDomain:
		csvAliasDomain(i.(AliasDomain), &r)
	case MailDomains:
		for _, m := range i.(MailDomains) {
			csvMailDomain(m, &r)
		}
	case MailDomain:
		csvMailDomain(i.(MailDomain), &r)
	case Records:
		for _, x := range i.(Records) {
			csvRecord(x, &r)
		}
	case Record:
		csvRecord(i.(Record), &r)
	}
	return r.String(), nil
}

func csvAliasDomain(a AliasDomain, r *strings.Builder) {
	for _, x := range a.Aliases {
		r.WriteString(fmt.Sprintf(`"%s", "%s", "%s", "%s", "%s", %v`, a.Domain, x.Address, x.DisplayAddress, strings.Join(x.ForwardsTo, ";"), strings.Join(x.PermittedSenders, ";"), x.Required))
		r.WriteByte('\n')
	}
}

func csvMailDomain(m MailDomain, r *strings.Builder) {
	for _, u := range m.Users {
		p := ""
		switch u.Privileges.(type) {
		case []interface{}:
			for i := 0; i < len(u.Privileges.([]interface{})); i++ {
				switch (u.Privileges.([]interface{}))[i].(type) {
				case string:
					p = (u.Privileges.([]interface{}))[i].(string)
				case []string:
					p = strings.Join((u.Privileges.([]interface{}))[i].([]string), "; ")
				}
			}
		}

		r.WriteString(fmt.Sprintf(`"%s", "%s", "%s", "%s", "%s"`, m.Domain, u.Email, p, u.Status, u.Mailbox))
		r.WriteByte('\n')
	}
}

func csvRecord(x Record, r *strings.Builder) {

	r.WriteString(fmt.Sprintf(`"%s", "%s", "%s"`, x.QName, x.RType, x.Value))
	r.WriteByte('\n')
}
