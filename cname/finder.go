// Package cname contains tools for finding the CNAMEs associated with a distribution in CloudFront.
package cname

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

// FromArgs takes command line arguments and prints matching CloudFront IDs. Warning: Exits for unknown flags.
func FromArgs(args []string) (err error) {
	fl := flag.NewFlagSet("boreas-find", flag.ExitOnError)
	fl.Usage = func() {
		fmt.Fprintf(os.Stderr,
			`Usage of boreas-find:

    boreas-find <cname>

Prints the IDs of CloudFront distributions with matching CNAME aliases to standard out.
`,
		)
		fl.PrintDefaults()
	}
	_ = fl.Parse(args)
	cname := fl.Arg(0)
	if cname == "" || fl.NArg() != 1 {
		fl.Usage()
		os.Exit(2)
	}

	s, err := session.NewSession()
	if err != nil {
		return err
	}
	cf := cloudfront.New(s)

	ids, err := GetID(cf, cname)
	if err != nil {
		return err
	}

	for _, id := range ids {
		fmt.Println(id)
	}

	return nil
}

// GetID returns a list of CloudFront IDs for distributions with CNAMEs matching cname.
func GetID(cf *cloudfront.CloudFront, cname string) (ids []string, err error) {
	idToCnames, err := List(cf)
	for id, cnames := range idToCnames {
		for _, name := range cnames {
			if strings.EqualFold(cname, name) {
				ids = append(ids, id)
			}
		}
	}
	return
}

// List returns a map from CloudFront distribution IDs to associate CNAME aliases.
func List(cf *cloudfront.CloudFront) (cnames map[string][]string, err error) {
	cnames = map[string][]string{}
	err = cf.ListDistributionsPages(&cloudfront.ListDistributionsInput{}, func(page *cloudfront.ListDistributionsOutput, lastPage bool) bool {
		for _, item := range page.DistributionList.Items {
			id := *item.Id
			for _, aliasp := range item.Aliases.Items {
				cnames[id] = append(cnames[id], *aliasp)
			}
		}
		return true
	})

	return
}
