package master

import "testing"

func TestDefaultDockCapabilities(t *testing.T) {
	dock := DefaultDockLocationType()
	if dock.Code != "DOCK" || !dock.IsActive || !dock.AllowsReceiving || !dock.AllowsShipping || dock.AllowsStorage || dock.AllowsPicking {
		t.Fatalf("dock must support loading/unloading without storage or picking: %+v", dock)
	}
}
