package miab

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

type Formats string

const (
	JSON  = Formats(`json`)
	YAML  = Formats(`yaml`)
	CSV   = Formats(`csv`)
	PLAIN = Formats(`plain`)
)

func print(i interface{}, format Formats) {
	switch format {
	case JSON:
		fmt.Print(marshallJson(i))
	case YAML:
		fmt.Print(marshallYaml(i))
	case CSV:
		fmt.Print(marshallCsv(i))
	default:
		fmt.Print(toString(i, PLAIN))
	}
}

func toString(i interface{}, format Formats) string {
	switch format {
	case JSON:
		return marshallJson(i)
	case YAML:
		return marshallYaml(i)
	case CSV:
		return ""
	default:
		switch i.(type) {
		case AliasDomains:
			return i.(AliasDomains).String()
		case AliasDomain:
			return i.(AliasDomain).String()
		case Records:
			return i.(Records).String()
		case Record:
			return i.(Record).String()
		case MailDomains:
			return i.(MailDomains).String()
		case MailDomain:
			return i.(MailDomain).String()
		}
	}
	return ""
}

func marshallJson(i interface{}) string {
	r, err := json.Marshal(i)
	if err != nil {
		fmt.Printf("Error marshalling json: %v\n", err)
		os.Exit(1)
	}

	return string(r)
}

func marshallYaml(i interface{}) string {
	r, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("Error marshalling yaml: %v\n", err)
		os.Exit(1)
	}

	return string(r)
}

func marshallCsv(i interface{}) string {

	r := strings.Builder{}
	switch i.(type) {
	case AliasDomains, AliasDomain:
		r.WriteString(`"domain", address", "displayAddress", "forwardsTo", "permittedSenders", "required"`)
	case MailDomains, MailDomain:
		r.WriteString(`"domain", "email", "privileges", "status", "mailbox"`)
	default:
		return ""
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
	}

	return r.String()
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
