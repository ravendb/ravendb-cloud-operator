package controller

import (
	"testing"

	cloudv1 "github.com/ravendb/ravendb-cloud-operator/api/v1"
	"github.com/ravendb/ravendb-cloud-operator/internal/cloudapi"
)

// Half a GiB apart, with the names sorting the other way round: a comparison
// that rounds the difference to whole GiB calls them equal and lets the name
// pick the larger, dearer shape.
func TestTheDefaultInstanceTypeIsTheSmallestWhateverTheNames(t *testing.T) {
	available := []cloudapi.InstanceType{
		{Name: "a-large", Tier: cloudapi.Tier(cloudv1.TierDevelopment), VirtualCPUs: 2, RAMGiB: 1},
		{Name: "b-small", Tier: cloudapi.Tier(cloudv1.TierDevelopment), VirtualCPUs: 2, RAMGiB: 0.5},
	}
	got, err := smallestForTier(available, cloudv1.TierDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	if got != "b-small" {
		t.Errorf("default instance type %q, want b-small, the one with less RAM", got)
	}
}
