package middleware

// Action-level permissions complement the module READ/WRITE permissions.
// READ remains the shared inquiry permission; mutations use the narrowest
// action permission exposed by their route.
const (
	PermissionInboundRead              = "INBOUND.READ"
	PermissionInboundPlan              = "INBOUND.PLAN"
	PermissionInboundApprove           = "INBOUND.APPROVE"
	PermissionInboundReceive           = "INBOUND.RECEIVE"
	PermissionInboundQC                = "INBOUND.QC"
	PermissionInboundPutaway           = "INBOUND.PUTAWAY"
	PermissionInboundAssign            = "INBOUND.ASSIGN"
	PermissionInboundQuarantineDispose = "INBOUND.QUARANTINE_DISPOSE"
	PermissionInboundRework            = "INBOUND.REWORK"
	PermissionInboundCancel            = "INBOUND.CANCEL"

	PermissionInventoryRead         = "INVENTORY.READ"
	PermissionInventoryIdentity     = "INVENTORY.IDENTITY"
	PermissionInventoryMove         = "INVENTORY.MOVE"
	PermissionInventoryStatusChange = "INVENTORY.STATUS_CHANGE"
	PermissionInventoryAdjust       = "INVENTORY.ADJUST"
	PermissionInventoryCount        = "INVENTORY.COUNT"
	PermissionInventoryTransfer     = "INVENTORY.TRANSFER"

	PermissionOutboundRead      = "OUTBOUND.READ"
	PermissionOutboundPlan      = "OUTBOUND.PLAN"
	PermissionOutboundPick      = "OUTBOUND.PICK"
	PermissionOutboundStage     = "OUTBOUND.STAGE"
	PermissionOutboundCheck     = "OUTBOUND.CHECK"
	PermissionOutboundPack      = "OUTBOUND.PACK"
	PermissionOutboundTransport = "OUTBOUND.TRANSPORT"
	PermissionOutboundShip      = "OUTBOUND.SHIP"
	PermissionOutboundDeliver   = "OUTBOUND.DELIVER"
	PermissionOutboundCancel    = "OUTBOUND.CANCEL"
	PermissionOutboundConfig    = "OUTBOUND.CONFIG"

	PermissionBillingRead      = "BILLING.READ"
	PermissionBillingConfigure = "BILLING.CONFIGURE"
	PermissionBillingPrepare   = "BILLING.PREPARE"
	PermissionBillingApprove   = "BILLING.APPROVE"
	PermissionBillingIssue     = "BILLING.ISSUE"
	PermissionBillingPayment   = "BILLING.PAYMENT"
)
