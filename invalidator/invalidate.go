// Package invalidator contains tools for invalidating a CloudFront distribution
package invalidator

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

type Invalidator struct {
	cf    *cloudfront.CloudFront
	dist  string
	ref   string
	paths []string
	wait  time.Duration
}

func FromArgs(args []string) (inv *Invalidator, err error) {
	inv = &Invalidator{}
	fl := flag.NewFlagSet("boreas", flag.ExitOnError)
	fl.StringVar(&inv.dist, "dist", "", "CloudFront distribution ID")
	fl.StringVar(&inv.ref, "ref", "",
		"CloudFront 'CallerReference', a unique identifier for this invalidation request. (default: Unix timestamp)")
	fl.DurationVar(&inv.wait, "wait", 15*time.Minute, "Time out for waiting on invalidation to complete. Set to 0 to exit without waiting.")
	fl.Usage = func() {
		fmt.Fprintf(os.Stderr,
			`Usage of boreas:

    boreas [options] <invalidation path>...

Invalidation path defaults to '/*'.

AWS credentials taken from ~/.aws/ or from "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", and other AWS configuration environment variables.


Options:

`,
		)
		fl.PrintDefaults()
	}
	_ = fl.Parse(args)

	if inv.dist == "" {
		return nil, errors.New("distribution must be set")
	}

	inv.paths = fl.Args()
	if len(inv.paths) < 1 {
		inv.paths = []string{"/*"}
	}
	for i, p := range inv.paths {
		if !strings.HasPrefix(p, "/") {
			inv.paths[i] = "/" + p
		}
	}

	if inv.ref == "" {
		inv.ref = fmt.Sprintf("%d", time.Now().Unix())
	}

	s, err := session.NewSession()
	inv.cf = cloudfront.New(s)
	return
}

func (inv *Invalidator) Execute() error {
	id, err := inv.Invalidate()
	if err != nil {
		return err
	}
	log.Printf("Invalidation ID: %q", id)

	if inv.wait < 1 {
		return nil
	}

	fmt.Print("Invalidation in progress")
	defer fmt.Println()
	deadline := time.Now().Add(inv.wait)
	for deadline.After(time.Now()) {
		time.Sleep(10 * time.Second)
		done, err := inv.Done(id)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		fmt.Print(".")
	}
	return fmt.Errorf("wait timeout of %v exceeded", inv.wait)
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

func (inv *Invalidator) Invalidate() (id string, err error) {
	result, err := inv.cf.CreateInvalidation(&cloudfront.CreateInvalidationInput{
		DistributionId: &inv.dist,
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: &inv.ref,
			Paths:           makepaths(inv.paths),
		},
	})
	if err != nil {
		return "", err
	}

	return *result.Invalidation.Id, nil
}

func (inv *Invalidator) Done(id string) (done bool, err error) {
	result, err := inv.cf.GetInvalidation(&cloudfront.GetInvalidationInput{
		DistributionId: &inv.dist,
		Id:             &id,
	})
	if err != nil {
		return false, err
	}
	return *result.Invalidation.Status == "Completed", nil
}
