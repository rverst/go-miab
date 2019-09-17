package command

import (
	"fmt"
	"github.com/rverst/go-miab/miab"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func init() {
	rootCmd.AddCommand(dnsGetCmd)
	dnsGetCmd.AddCommand(dnsSetCmd, dnsAddCmd, dnsDeleteCmd)

	dnsGetCmd.Flags().String("format", "plain", "the output format (plain, csv, json, yaml)")
	dnsGetCmd.Flags().String("domain", "", "Domain to filter the list of dns records, can be part of a domain (e.g. '.org'). Not considered if the qname-flag is set.")
	dnsGetCmd.Flags().String("rtype", "", "The resource type to filter the output. (A, AAAA, TXT, CNAME, MX, SRV, SSHFP, CAA, NS)")
	dnsGetCmd.Flags().String("qname", "", "The fully qualified domain to filter the output. NOTE: the rtype-flag defaults to 'A' if you use this flag, the domain-flag will be ignored.")
	dnsGetCmd.Flags().BoolP("short", "s", false, "If the qname-flag is set, the output given is just the records value.")

	dnsSetCmd.Flags().String("qname", "", "The fully qualified domain name for the record you are trying to set. It must be one of the domain names or a subdomain of one of the domain names hosted on the box. (Add mail users or aliases to add new domains.)")
	dnsSetCmd.Flags().String("rtype", "A", "The resource type. Defaults to A if omitted. Possible values: A (an IPv4 address), AAAA (an IPv6 address), TXT (a text string), CNAME (an alias, which is a fully qualified domain name — don’t forget the final period), MX, SRV, SSHFP, CAA or NS.")
	dnsSetCmd.Flags().String("value", "", "The record’s value. If the 'rtype' is A or AAAA and 'value' is empty or omitted, the IPv4 or IPv6 address of the remote host is used.")

	dnsAddCmd.Flags().String("qname", "", "The fully qualified domain name for the record you are trying to add. It must be one of the domain names or a subdomain of one of the domain names hosted on the box. (Add mail users or aliases to add new domains.)")
	dnsAddCmd.Flags().String("rtype", "A", "The resource type. Defaults to A if omitted. Possible values: A (an IPv4 address), AAAA (an IPv6 address), TXT (a text string), CNAME (an alias, which is a fully qualified domain name — don’t forget the final period), MX, SRV, SSHFP, CAA or NS.")
	dnsAddCmd.Flags().String("value", "", "The record’s value. If the 'rtype' is A or AAAA and 'value' is empty or omitted, the IPv4 or IPv6 address of the remote host is used.")

	dnsDeleteCmd.Flags().String("qname", "", "The fully qualified domain name for the record you are trying to delete.")
	dnsDeleteCmd.Flags().String("rtype", "A", "The resource type. Defaults to A if omitted. (A, AAAA, TXT, CNAME, MX, SRV, SSHFP, CAA, NS)")
	dnsDeleteCmd.Flags().String("value", "", "The record’s value. If 'value' is empty or omitted, all records matching the qname-flag and rtype-flag will be deleted.")
}

var dnsGetCmd = &cobra.Command{
	Use:   "dns",
	Short: "Get existing dns entries",
	Long: `Get all dns entries of the server, use the domain-flag to filter the output.
			Due to a wired behavior of the Mail-in-a-box API, you can use the qname-flag and the rtype-flag
			to filter exactly one record, in this case the the domain-flag will be ignored. 
			NOTE: If you use the qname-flag, the rtype-flag defaults to 'A'.`,
	Args:             cobra.NoArgs,
	Run:              getDns,
	PersistentPreRun: initConfig,
}

var dnsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets a custom DNS record replacing any existing records with the same `qname` and `rtype`",
	Long: `Sets a custom DNS record replacing any existing records with the same 'qname' and 'rtype'. 
Use 'set' (instead of 'add') when you only have one value for a 'qname' and 'rtype',
such as typical A records (without round-robin).`,
	Args: cobra.NoArgs,
	Run:  setDns,
}

var dnsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new custom DNS record. Use 'add' when you have multiple TXT records or round-robin A records",
	Long: `Adds a new custom DNS record. Use 'add' when you have multiple TXT records or round-robin A records.
('set' would delete previously added records.)`,
	Args: cobra.NoArgs,
	Run:  addDns,
}

var dnsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes custom DNS records",
	Long: `Deletes custom DNS records. If the value-flag is omitted, it deletes all records matching the
qname-flag and rtype-flag.`,
	Args: cobra.NoArgs,
	Run:  delDns,
}

func getDns(cmd *cobra.Command, args []string) {

	format := miab.PLAIN
	if f, err := cmd.Flags().GetString("format"); err == nil {
		format = miab.Format(f)
	}

	rtype := miab.NONE
	if r, err := cmd.Flags().GetString("rtype"); err == nil {
		rtype = miab.ResourceType(r)
	}
	qname, _ := cmd.Flags().GetString("qname")

	records, err := miab.GetDns(&config, qname, rtype)
	if err != nil {
		fmt.Printf("Error fetching dns records: %v\n", err)
		os.Exit(1)
	}

	if qname == "" {
		d, err := cmd.Flags().GetString("domain")
		if err == nil && d != "" {
			filtered := miab.Records{}
			for _, record := range records {
				if strings.Contains(record.QName, d) {

					if rtype != miab.NONE {
						if rtype == record.RType {
							filtered = append(filtered, record)
						}
					} else {
						filtered = append(filtered, record)
					}
				}
			}
			fmt.Println(filtered.ToString(format))
			return
		}
	}

	if qname != "" && len(records) == 1 {
		fmt.Print(records[0].Value)
		return
	}

	fmt.Println(records.ToString(format))
}

func setDns(cmd *cobra.Command, args []string) {
	rtype := miab.NONE
	if r, err := cmd.Flags().GetString("rtype"); err == nil {
		rtype = miab.ResourceType(r)
	}
	qname, _ := cmd.Flags().GetString("qname")
	value, _ := cmd.Flags().GetString("value")

	if value == "" {
		dynDnsUpdate(rtype, qname, value, false)
	}

	if s, err := miab.SetDns(&config, qname, rtype, value); !s || err != nil {
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

func addDns(cmd *cobra.Command, args []string) {
	rtype := miab.NONE
	if r, err := cmd.Flags().GetString("rtype"); err == nil {
		rtype = miab.ResourceType(r)
	}
	qname, _ := cmd.Flags().GetString("qname")
	value, _ := cmd.Flags().GetString("value")

	if value == "" {
		dynDnsUpdate(rtype, qname, value, true)
	}
	if s, err := miab.AddDns(&config, qname, rtype, value); !s || err != nil {
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

func delDns(cmd *cobra.Command, args []string) {
	rtype := miab.NONE
	if r, err := cmd.Flags().GetString("rtype"); err == nil {
		rtype = miab.ResourceType(r)
	}
	qname, _ := cmd.Flags().GetString("qname")
	value, _ := cmd.Flags().GetString("value")

	if s, err := miab.DeleteDns(&config, qname, rtype, value); !s || err != nil {
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

func dynDnsUpdate(rtype miab.ResourceType, qname, value string, add bool) {
	if rtype == miab.A {
		if s, err := miab.SetOrAddAddressRecord(&config, miab.TCP4, qname, value, add); !s || err != nil {
			if err != nil {
				fmt.Println(err)
			}
			os.Exit(1)
		}
	} else if rtype == miab.AAAA {
		if s, err := miab.SetOrAddAddressRecord(&config, miab.TCP6, qname, value, add); !s || err != nil {
			if err != nil {
				fmt.Println(err)
			}
			os.Exit(1)
		} else {
			fmt.Println("The value-flag can only be omitted is rtype-flag is 'A' or 'AAAA'.")
			os.Exit(1)
		}
	}
}
