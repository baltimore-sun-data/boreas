package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

func main() {
	enc, err := FromArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initial error: %v\n", err)
		os.Exit(3)
	}
	if err = enc.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Run time error: %v\n", err)
		os.Exit(1)
	}
}

func deferClose(err *error, f func() error) {
	newErr := f()
	if *err == nil {
		*err = newErr
	}
}

type Invalidator struct {
	s     *session.Session
	dist  string
	ref   string
	paths []string
}

func FromArgs(args []string) (i *Invalidator, err error) {
	i = &Invalidator{}
	fl := flag.NewFlagSet("boreas", flag.ExitOnError)
	fl.StringVar(&i.dist, "dist", "", "CloudFront distribution ID")
	fl.StringVar(&i.ref, "ref", "",
		"CloudFront 'CallerReference', a unique identifier for this invalidation request. (default: Unix timestamp)")
	_ = fl.Parse(args)

	if i.dist == "" {
		return nil, errors.New("distribution must be set")
	}

	i.paths = fl.Args()
	if len(i.paths) < 1 {
		return nil, errors.New("at least one invalidation path must be provided")
	}

	if i.ref == "" {
		i.ref = fmt.Sprintf("%d", time.Now().Unix())
	}

	i.s, err = session.NewSession()
	return
}

func (i *Invalidator) Execute() error {
	svc := cloudfront.New(i.s)
	input := &cloudfront.CreateInvalidationInput{
		DistributionId: &i.dist,
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: &i.ref,
			Paths:           makepaths(i.paths),
		},
	}

	result, err := svc.CreateInvalidation(input)
	if err != nil {
		return err
	}
	log.Printf("Invalidation ID: %q", *result.Invalidation.Id)
	return nil
}

func makepaths(paths []string) *cloudfront.Paths {
	items := make([]*string, len(paths))
	for i := range paths {
		items[i] = &paths[i]
	}
	quantity := int64(len(items))
	return &cloudfront.Paths{
		Items:    items,
		Quantity: &quantity,
	}
}
