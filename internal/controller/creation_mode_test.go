package controller

import (
	"context"
	"strings"
	"testing"

	cloudv1 "github.com/ravendb/ravendb-cloud-operator/api/v1"
	"github.com/ravendb/ravendb-cloud-operator/internal/cloudsim"
)

func TestTakeOverOnlyNeverCreates(t *testing.T) {
	h := newHarness(t, credentialSecret(), readyAccount(), provisionCluster("orders", func(c *cloudv1.CloudCluster) {
		c.Spec.CreationMode = cloudv1.CreationTakeOverOnly
	}))
	h.reconcileCluster("orders")

	if n := h.cloud.CallCount("CreateCluster"); n != 0 {
		t.Errorf("sent %d creates under TakeOverOnly", n)
	}
	if got := conditionReason(h.getCluster("orders").Status.Conditions, cloudv1.ConditionReady); got != cloudv1.ReasonClusterNotFound {
		t.Errorf("Ready = %q, want ClusterNotFound", got)
	}
}

func TestCreateOnlyNeverTakesOver(t *testing.T) {
	h := newHarness(t, credentialSecret(), readyAccount(), provisionCluster("orders", func(c *cloudv1.CloudCluster) {
		c.Spec.CreationMode = cloudv1.CreationCreateOnly
	}))
	h.seedCloudCluster(cloudsim.SeedSpec{DisplayName: "orders"})
	h.reconcileCluster("orders")

	cluster := h.getCluster("orders")
	if cluster.Status.ProductID != "" {
		t.Errorf("bound %s under CreateOnly", cluster.Status.ProductID)
	}
	if got := conditionReason(cluster.Status.Conditions, cloudv1.ConditionReady); got != cloudv1.ReasonNameTaken {
		t.Errorf("Ready = %q, want NameTaken", got)
	}
}

// Terminate ends the cluster an object holds however it came to hold it.
// Whether the object created the cluster is known only from status, which a
// restore from git loses, and a Terminate that worked only until then would
// leave the cluster running and billed. Admission is what keeps Terminate to
// a cluster the manifest claims.
func TestTerminateEndsAClusterTakenOver(t *testing.T) {
	h := newHarness(t, credentialSecret(), readyAccount(), takeOverCluster("orders", func(c *cloudv1.CloudCluster) {
		c.Spec.CreationMode = cloudv1.CreationTakeOverOnly
		c.Spec.DeletionPolicy = cloudv1.DeletionTerminate
	}))
	seeded := h.seedCloudCluster(cloudsim.SeedSpec{DisplayName: "orders"})
	h.reconcileClusterUntilSettled("orders", 6)
	if err := h.k8s.Delete(context.Background(), h.getCluster("orders")); err != nil {
		t.Fatalf("deleting: %v", err)
	}
	h.reconcileCluster("orders")

	if n := h.cloud.CallCount("TerminateCluster"); n != 1 {
		t.Errorf("TerminateCluster called %d times, want once", n)
	}
	if got, _ := h.cloudCluster(seeded.ID); got.Status == "Active" {
		t.Error("the cluster is still Active")
	}
}

// A takeover nobody asked for warns; one the manifest asked for is routine.
func TestATakeoverWarnsUnlessTheManifestAskedForIt(t *testing.T) {
	for mode, want := range map[cloudv1.CreationMode]string{
		cloudv1.CreationTakeOverOrCreate: "Warning " + cloudv1.EventAdopted,
		cloudv1.CreationTakeOverOnly:     "Normal " + cloudv1.EventAdopted,
	} {
		t.Run(string(mode), func(t *testing.T) {
			h := newHarness(t, credentialSecret(), readyAccount(), takeOverCluster("orders", func(c *cloudv1.CloudCluster) {
				c.Spec.CreationMode = mode
			}))
			h.seedCloudCluster(cloudsim.SeedSpec{DisplayName: "orders"})
			h.reconcileCluster("orders")

			if events := strings.Join(h.events(), "\n"); !strings.Contains(events, want) {
				t.Errorf("no %q event; events:\n%s", want, events)
			}
		})
	}
}
