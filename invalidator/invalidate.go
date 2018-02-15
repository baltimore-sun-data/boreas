// Package invalidator contains tools for invalidating a CloudFront distribution
package invalidator

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

type Invalidator struct {
	s     *session.Session
	dist  string
	ref   string
	paths []string
}

func FromArgs(args []string) (inv *Invalidator, err error) {
	inv = &Invalidator{}
	fl := flag.NewFlagSet("boreas", flag.ExitOnError)
	fl.StringVar(&inv.dist, "dist", "", "CloudFront distribution ID")
	fl.StringVar(&inv.ref, "ref", "",
		"CloudFront 'CallerReference', a unique identifier for this invalidation request. (default: Unix timestamp)")
	_ = fl.Parse(args)

	if inv.dist == "" {
		return nil, errors.New("distribution must be set")
	}

	inv.paths = fl.Args()
	if len(inv.paths) < 1 {
		return nil, errors.New("at least one invalidation path must be provided")
	}

	if inv.ref == "" {
		inv.ref = fmt.Sprintf("%d", time.Now().Unix())
	}

	inv.s, err = session.NewSession()
	return
}

func (inv *Invalidator) Execute() error {
	svc := cloudfront.New(inv.s)
	input := &cloudfront.CreateInvalidationInput{
		DistributionId: &inv.dist,
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: &inv.ref,
			Paths:           makepaths(inv.paths),
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
