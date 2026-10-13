package outbound

import (
	"testing"
	model "wms-api/models/outbound"
	repository "wms-api/repository/outbound"
)

func TestOutstandingDemandUsesBaseSnapshots(t *testing.T) {
	line := model.OutboundOrderLine{OrderedQty: "10", OrderedBaseQty: "120", AllocatedQty: "90", ShippedQty: "85", ShortAcceptedQty: "5", ClosedShortQty: "0"}
	if decimal(openDemand(line)) != "30.000000" || decimal(unallocatedDemand(line)) != "30.000000" {
		t.Fatal("base demand or checking shortage calculated incorrectly")
	}
	line.ClosedShortQty = "30"
	if hasOpenDemand([]repository.OutboundOrderLineRow{{OutboundOrderLine: line}}) {
		t.Fatal("closed-short quantity remains open")
	}
	if decimal(openDemand(line)) != "0.000000" {
		t.Fatal("closure was counted as a shipment")
	}
	line.ShippedQty = "80"
	if !hasOpenDemand([]repository.OutboundOrderLineRow{{OutboundOrderLine: line}}) {
		t.Fatal("in-process stock was hidden")
	}
}
