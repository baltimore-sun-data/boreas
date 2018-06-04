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

// Invalidator is an object that can invalidate a specific CloudFront
// distibution or check the status of an existing invalidation.
type Invalidator struct {
	cf    *cloudfront.CloudFront
	dist  string
	ref   string
	paths []string
	wait  time.Duration
}

// New creates a useable new Invalidator. If cf is nil, New attempts to create
// an Amazon default session and panics on failure. If callerReference is
// "", a Unix timestamp is used.
func New(cf *cloudfront.CloudFront, callerReference, distID string, paths ...string) *Invalidator {
	if cf == nil {
		cf = cloudfront.New(session.Must(session.NewSession()))
	}
	if callerReference == "" {
		callerReference = fmt.Sprintf("%d", time.Now().Unix())
	}
	return &Invalidator{
		cf:    cf,
		dist:  distID,
		ref:   callerReference,
		paths: paths,
	}
}

// FromArgs creates a new Invalidator from command line arguments. Warning: Exits on error.
func FromArgs(args []string) *Invalidator {
	inv := &Invalidator{}
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
	return inv
}

// Execute runs the invalidator for the CLI.
func (inv *Invalidator) Execute() error {
	if inv.dist == "" {
		return errors.New("distribution must be set")
	}
	if inv.cf == nil {
		s, err := session.NewSession()
		if err != nil {
			return fmt.Errorf("could not create a new session: %v", err)
		}
		inv.cf = cloudfront.New(s)
	}

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

// Invalidate invalidates the underlying distribution. The id returned can be used with Done().
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

// Done returns a bool indicating whether the referenced invalidation has completed.
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
