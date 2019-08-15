package miab

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

type Formats string

const (
	CsvDnsHead   = `"domain name", "record type", "value"`
	CsvUserHead  = `"domain", "email", "privileges", "status", "mailbox"`
	CsvAliasHead = `"domain", address", "displayAddress", "forwardsTo", "permittedSenders", "required"`
)

const (
	JSON  = Formats(`json`)
	YAML  = Formats(`yaml`)
	CSV   = Formats(`csv`)
	PLAIN = Formats(`plain`)
)

func print(i interface{}, format Formats) {
	var s string
	var err error

	switch format {
	case JSON:
		s, err = marshallJson(i)
	case YAML:
		s, err = marshallYaml(i)
	case CSV:
		s, err = marshallCsv(i)
	default:
		s, err = toString(i, PLAIN)
	}

	if err != nil {
		fmt.Print("unexpected error", err)
		os.Exit(1)
	}
	fmt.Print(s)
}

func toString(i interface{}, format Formats) (string, error) {
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
		r.WriteString(CsvAliasHead)
	case MailDomains, MailDomain:
		r.WriteString(CsvUserHead)
	case Records, Record:
		r.WriteString(CsvDnsHead)
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
