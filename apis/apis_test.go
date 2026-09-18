package apis

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

// TestAllKindsRegistered guards against the v0.4.1 outage where
// guest/sharedlink/team/customproperty addKnownTypes omitted
// s.AddKnownTypes, so the manager failed at startup with
// "no kind is registered for the type v1beta1.XList".
func TestAllKindsRegistered(t *testing.T) {
	s := runtime.NewScheme()
	if err := AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme failed: %v", err)
	}
	for gv, kinds := range map[string][]string{
		"site.plausible.m.crossplane.io/v1beta1":           {"Site", "SiteList"},
		"goal.plausible.m.crossplane.io/v1beta1":           {"Goal", "GoalList"},
		"guest.plausible.m.crossplane.io/v1beta1":          {"Guest", "GuestList"},
		"sharedlink.plausible.m.crossplane.io/v1beta1":     {"SharedLink", "SharedLinkList"},
		"team.plausible.m.crossplane.io/v1beta1":           {"Team", "TeamList"},
		"customproperty.plausible.m.crossplane.io/v1beta1": {"CustomProperty", "CustomPropertyList"},
	} {
		for _, k := range kinds {
			gvk := s.AllKnownTypes()
			found := false
			for known := range gvk {
				if known.GroupVersion().String() == gv && known.Kind == k {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("kind %s in %s not registered in scheme", k, gv)
			}
		}
	}
}
