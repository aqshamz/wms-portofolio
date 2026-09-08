-- Idempotent starter data for a new WMS database.
-- Review the names and workflow choices before production use.
-- Existing rows with the same business key are preserved.

BEGIN;
SET search_path TO wms, public;

-- Application modules are referenced by permissions, reasons, documents, and
-- billable services. Keeping them here prevents unchecked repeated text values.
INSERT INTO app_module (code, name, display_order)
VALUES
    ('AUTH',      'Authentication and access', 10),
    ('SECURITY',  'Security administration',   15),
    ('MASTER',    'Master data',               20),
    ('INBOUND',   'Inbound operations',        30),
    ('INVENTORY', 'Inventory control',         40),
    ('OUTBOUND',  'Outbound operations',       50),
    ('BILLING',   'Billing and receivables',   60),
    ('REPORTING', 'Reports and analytics',     70),
    ('GENERAL',   'General/shared',             90)
ON CONFLICT (code) DO NOTHING;

-- Authentication configuration ------------------------------------------------

INSERT INTO authentication_policy (
    code, name, max_failed_attempts, lockout_seconds, session_ttl_seconds,
    is_default, is_active
) VALUES (
    'DEFAULT', 'Default authentication policy', 5, 900, 28800, false, true
)
ON CONFLICT (code) DO NOTHING;

UPDATE authentication_policy target
SET is_default = true
WHERE target.code = 'DEFAULT'
  AND target.is_active
  AND NOT EXISTS (
      SELECT 1
      FROM authentication_policy current_default
      WHERE current_default.is_default
        AND current_default.is_active
  );

INSERT INTO account_status (code, name, description, allows_login, is_active)
VALUES
    ('ACTIVE',   'Active',   'Account may authenticate.', true,  true),
    ('PENDING',  'Pending',  'Account is waiting for activation.', false, true),
    ('LOCKED',   'Locked',   'Account is administratively locked.', false, true),
    ('DISABLED', 'Disabled', 'Account is no longer permitted to authenticate.', false, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO session_revocation_reason (code, name, description)
VALUES
    ('USER_LOGOUT',      'User logout',      'The user ended the current session.'),
    ('USER_LOGOUT_ALL',  'Logout all',       'The user ended every active session.'),
    ('PASSWORD_CHANGED', 'Password changed', 'Sessions were revoked after a password change.'),
    ('ACCOUNT_DISABLED', 'Account disabled', 'Sessions were revoked because the account was disabled.'),
    ('ADMIN_REVOKED',    'Administrator revoked', 'An administrator revoked the session.')
ON CONFLICT (code) DO NOTHING;

-- Partner and warehouse reference data ----------------------------------------

INSERT INTO partner_type (code, name, description)
VALUES
    ('SUPPLIER', 'Supplier', 'Supplies goods to an owner.'),
    ('FACTORY',  'Factory',  'Manufactures or dispatches goods for an owner.'),
    ('CUSTOMER', 'Customer', 'Receives outbound goods.'),
    ('STORE',    'Store',    'Retail or branch destination for store delivery.'),
    ('CARRIER',  'Carrier',  'Transports goods.'),
    ('OTHER',    'Other',    'Other business relationship.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO location_type (
    code, name, description,
    allows_receiving, allows_storage, allows_picking, allows_shipping
)
VALUES
    ('RECEIVING',  'Receiving',  'Inbound receiving and staging.', true,  false, false, false),
    ('STORAGE',    'Storage',    'Reserve storage location.',       false, true,  false, false),
    ('PICK_FACE',  'Pick face',  'Forward picking location.',       false, true,  true,  false),
    ('STAGING',    'Staging',    'Outbound stock staged by delivery order.', false, true, false, false),
    ('CHECKING',   'Checking',   'Outbound verification area.',      false, true,  false, false),
    ('PACKING',    'Packing',    'Packing workstation or area.',    false, false, false, false),
    ('SHIPPING',   'Shipping',   'Outbound shipping and staging.',  false, false, false, true),
    ('QUARANTINE', 'Quarantine', 'Stock awaiting disposition.',     false, true,  false, false)
ON CONFLICT (code) DO NOTHING;

-- Inventory reference data -----------------------------------------------------

INSERT INTO uom (code, name, decimal_scale)
VALUES
    ('EA',  'Each',      0),
    ('BOX', 'Box',       0),
    ('CTN', 'Carton',    0),
    ('PLT', 'Pallet',    0),
    ('KG',  'Kilogram',  3),
    ('G',   'Gram',      3),
    ('L',   'Litre',     3),
    ('M',   'Metre',     3)
ON CONFLICT (code) DO NOTHING;

INSERT INTO inventory_status (
    code, name, description, is_allocatable, is_pickable
)
VALUES
    ('QC_PENDING', 'QC pending', 'Stock received and waiting for quality inspection.', false, false),
    ('PUTAWAY_PENDING', 'Putaway pending', 'Quality-approved stock waiting for physical putaway.', false, false),
    ('AVAILABLE',  'Available',  'Stock available for allocation and picking.', true,  true),
    ('HOLD',       'Hold',       'Stock held from allocation.',                 false, false),
    ('QUARANTINE', 'Quarantine', 'Stock awaiting inspection or disposition.',   false, false),
    ('DAMAGED',    'Damaged',    'Damaged stock.',                              false, false),
    ('EXPIRED',    'Expired',    'Expired stock.',                              false, false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO quarantine_disposition_type (
    code, name, description,
    releases_to_available, requires_reinspection, removes_inventory
)
VALUES
    ('ACCEPT',  'Accept',  'Client accepts quarantined stock for storage.', true,  false, false),
    ('REWORK',  'Rework',  'Stock must be reworked and inspected again.',    false, true,  false),
    ('RETURN',  'Return',  'Stock is returned to the vendor or factory.',    false, false, true),
    ('DISPOSE', 'Dispose', 'Stock is destroyed or otherwise disposed.',      false, false, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO quality_status (code, name, description)
VALUES
    ('PENDING',  'Pending',  'Inspection has not been completed.'),
    ('PASSED',   'Passed',   'Quality requirements were satisfied.'),
    ('FAILED',   'Failed',   'Quality requirements were not satisfied.'),
    ('WAIVED',   'Waived',   'Inspection was waived by an authorized user.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO inspection_result (code, name, description, is_accepted)
VALUES
    ('ACCEPTED', 'Accepted', 'All inspected stock was accepted.', true),
    ('PARTIAL',  'Partially accepted', 'Some inspected stock was accepted.', true),
    ('REJECTED', 'Rejected', 'Inspected stock was rejected.', false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO handling_unit_type (code, name)
VALUES
    ('PALLET', 'Pallet'),
    ('CARTON', 'Carton'),
    ('TOTE',   'Tote'),
    ('BIN',    'Bin')
ON CONFLICT (code) DO NOTHING;

INSERT INTO movement_type (code, name, description)
VALUES
    ('RECEIVE',          'Receive',          'Inventory received into the warehouse.'),
    ('PUTAWAY',          'Putaway',          'Inventory moved from staging to storage.'),
    ('PICK',             'Pick',             'Inventory moved for outbound picking.'),
    ('SHIP',             'Ship',             'Inventory shipped from the warehouse.'),
    ('TRANSFER',         'Transfer',         'Inventory moved between locations or warehouses.'),
    ('ADJUSTMENT',       'Adjustment',       'Authorized quantity adjustment.'),
    ('STATUS_CHANGE',    'Status change',    'Inventory status changed.'),
    ('COUNT_CORRECTION', 'Count correction', 'Inventory corrected from a stock count.'),
    ('RETURN_TO_VENDOR', 'Return to vendor', 'Rejected inventory returned to a vendor or factory.'),
    ('DISPOSE',          'Dispose',          'Rejected inventory removed through disposal.'),
    ('RECEIPT_REVERSAL', 'Receipt reversal', 'Untouched received inventory removed by an authorized receipt reversal.'),
    ('PUTAWAY_REVERSAL', 'Putaway reversal', 'Completed putaway returned to QC for controlled correction.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO movement_type (code, name, description)
VALUES
    ('INTERNAL_MOVE', 'Internal move', 'Inventory moved between locations in one warehouse.'),
    ('TRANSFER_OUT',  'Transfer out',  'Inventory dispatched from the source warehouse.'),
    ('TRANSFER_IN',   'Transfer in',   'Dispatched inventory received by the target warehouse.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO movement_type (code, name, description)
VALUES
    ('PACK',            'Pack',            'Checked inventory moved from staging into a packing location or package.'),
    ('DELIVERY_RETURN', 'Delivery return', 'Undelivered shipped inventory returned to the warehouse.'),
    ('OUTBOUND_CHECK_CORRECTION', 'Outbound check correction',
     'Failed-check stock moved out of active staging for investigation or replacement.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO validation_severity (code, name, blocks_processing)
VALUES
    ('ERROR',   'Error',   true),
    ('WARNING', 'Warning', false)
ON CONFLICT (code) DO NOTHING;

WITH rules(code, name, description, handler_code, severity_code, display_order) AS (
    VALUES
        ('DO_NUMBER',       'Delivery-order number', 'Client DO number must be present and unique.', 'REQUIRE_DO_NUMBER', 'ERROR', 10),
        ('SHIP_TO',         'Ship-to destination',   'An active store/customer and address are required.', 'REQUIRE_SHIP_TO', 'ERROR', 20),
        ('ORDER_LINES',     'Order lines',           'At least one valid item line is required.', 'REQUIRE_ORDER_LINES', 'ERROR', 30),
        ('BASE_UOM',        'Base UOM',              'Outbound quantities must use the item base UOM.', 'REQUIRE_BASE_UOM', 'ERROR', 40),
        ('REQUESTED_SHIP',  'Requested ship time',   'Requested shipment time should be present.', 'REQUIRE_REQUESTED_SHIP', 'WARNING', 50)
)
INSERT INTO outbound_validation_rule (
    code, name, description, handler_code, validation_severity_id, display_order
)
SELECT rules.code, rules.name, rules.description, rules.handler_code,
       severity.validation_severity_id,
       rules.display_order
FROM rules
JOIN validation_severity severity ON severity.code = rules.severity_code
ON CONFLICT (code) DO NOTHING;

INSERT INTO outbound_wave_type (code, name, description)
VALUES
    ('SINGLE_ORDER', 'Single order', 'One client DO per wave.'),
    ('MANUAL',       'Manual',       'Orders manually grouped by an operator.'),
    ('ROUTE',        'Route',        'Orders grouped for a delivery route.'),
    ('CARRIER',      'Carrier',      'Orders grouped by carrier or carrier service.'),
    ('CUT_OFF',      'Cut-off time', 'Orders grouped by dispatch cut-off time.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO outbound_check_result (
    code, name, description, is_pass, requires_note
)
VALUES
    ('PASS',       'Pass',       'Item and quantity match the DO and staged stock.', true,  false),
    ('SHORT',      'Short',      'Checked quantity is below expected quantity.',     false, true),
    ('OVER',       'Over',       'Checked quantity is above expected quantity.',     false, true),
    ('WRONG_ITEM', 'Wrong item', 'Staged item does not match the expected item.',    false, true),
    ('DAMAGED',    'Damaged',    'Stock was damaged before packing.',                false, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO outbound_check_exception_status (code, name, is_final)
VALUES
    ('OPEN',        'Open',        false),
    ('IN_PROGRESS', 'In progress', false),
    ('RESOLVED',    'Resolved',    true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO outbound_check_resolution_type (
    code, name, description,
    counts_as_stock_correction, counts_as_replacement,
    counts_as_short_acceptance, requires_approval
)
VALUES
    ('STOCK_CORRECTION', 'Stock correction',
     'Remove, hold, or adjust discrepant stock and link the inventory movement.',
     true, false, false, false),
    ('REPLACEMENT', 'Replacement stock',
     'Allocate and pick replacement stock linked to the exception.',
     false, true, false, false),
    ('ACCEPT_SHORT', 'Accept short quantity',
     'Authorized client decision to close demand without replacement.',
     false, false, true, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO delivery_event_type (
    code, name, description, marks_delivered, marks_failed, marks_returned
)
VALUES
    ('DEPARTED',          'Departed warehouse', 'Shipment departed from the warehouse.', false, false, false),
    ('ARRIVED_STORE',     'Arrived at store',   'Vehicle arrived at the store.',          false, false, false),
    ('DELIVERED',         'Delivered',          'Store accepted the shipment.',           true,  false, false),
    ('DELIVERY_FAILED',   'Delivery failed',    'Store delivery could not be completed.', false, true,  false),
    ('RETURNED_TO_DEPOT', 'Returned to depot', 'Undelivered stock returned to warehouse.',false,false,true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO delivery_failure_reason (code, name, description)
VALUES
    ('STORE_CLOSED',       'Store closed',       'The destination store was closed.'),
    ('RECIPIENT_REJECTED', 'Recipient rejected', 'The recipient rejected the delivery.'),
    ('ADDRESS_NOT_FOUND',  'Address not found',  'The driver could not locate the destination.'),
    ('DAMAGED_IN_TRANSIT', 'Damaged in transit', 'Stock was damaged during transport.'),
    ('VEHICLE_ISSUE',      'Vehicle issue',      'A vehicle problem prevented delivery.'),
    ('OTHER',              'Other',              'Another documented delivery failure.')
ON CONFLICT (code) DO NOTHING;

-- Billing reference data ------------------------------------------------------

INSERT INTO currency (code, name, decimal_scale)
VALUES
    ('IDR', 'Indonesian rupiah', 2),
    ('USD', 'United States dollar', 2)
ON CONFLICT (code) DO NOTHING;

INSERT INTO billing_cycle (code, name, period_days, period_months, is_on_demand)
VALUES
    ('DAILY',     'Daily',     1,    NULL, false),
    ('WEEKLY',    'Weekly',    7,    NULL, false),
    ('MONTHLY',   'Monthly',   NULL, 1,    false),
    ('ON_DEMAND', 'On demand', NULL, NULL, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO payment_term (code, name, due_days)
VALUES
    ('DUE_ON_RECEIPT', 'Due on receipt', 0),
    ('NET_7',          'Net 7 days',      7),
    ('NET_14',         'Net 14 days',     14),
    ('NET_30',         'Net 30 days',     30)
ON CONFLICT (code) DO NOTHING;

INSERT INTO charge_basis (
    code, name, description, requires_uom, use_event_quantity
)
VALUES
    ('PER_UNIT',     'Per unit',      'Charge for each billable inventory unit.', true,  true),
    ('PER_DOCUMENT', 'Per document',  'One charge per source document.',          false, false),
    ('PER_LINE',     'Per line',      'One charge per source document line.',      false, false),
    ('PER_TASK',     'Per task',      'One charge per completed warehouse task.', false, false),
    ('PER_WEIGHT',   'Per weight',    'Charge by weight quantity.',                true,  true),
    ('PER_VOLUME',   'Per volume',    'Charge by volume quantity.',                true,  true),
    ('PER_PALLET_DAY','Per pallet-day','Daily pallet storage charge.',             true,  true),
    ('PER_UNIT_DAY', 'Per unit-day',  'Daily storage charge by inventory unit.',   true,  true),
    ('FLAT',         'Flat charge',   'Fixed charge for an event.',                false, false)
ON CONFLICT (code) DO NOTHING;

WITH services(code, name, module_code, description, basis_code, uom_code) AS (
    VALUES
        ('INBOUND_RECEIVING', 'Inbound receiving', 'INBOUND', 'Goods physically received.', 'PER_UNIT', 'EA'),
        ('INBOUND_QC',        'Inbound QC',        'INBOUND', 'Quality inspection performed.', 'PER_UNIT', 'EA'),
        ('PUTAWAY',           'Putaway',           'INBOUND', 'Accepted stock put away.', 'PER_UNIT', 'EA'),
        ('STORAGE_UNIT_DAY',  'Storage unit-day',  'INVENTORY','Daily stored inventory.', 'PER_UNIT_DAY', 'EA'),
        ('STORAGE_PALLET_DAY','Storage pallet-day','INVENTORY','Daily occupied pallet/HU storage.', 'PER_PALLET_DAY', 'PLT'),
        ('INTERNAL_MOVE',     'Internal movement', 'INVENTORY','Internal relocation or replenishment.', 'PER_UNIT', 'EA'),
        ('WAREHOUSE_TRANSFER','Warehouse transfer','INVENTORY','Inter-warehouse transfer handling.', 'PER_UNIT', 'EA'),
        ('REWORK',            'Rework',            'INBOUND', 'Inbound rework activity.', 'PER_UNIT', 'EA'),
        ('OUTBOUND_PICK',     'Outbound picking',  'OUTBOUND','Inventory picked.', 'PER_UNIT', 'EA'),
        ('OUTBOUND_PACK',     'Outbound packing',  'OUTBOUND','Checked inventory packed.', 'PER_UNIT', 'EA'),
        ('OUTBOUND_SHIP',     'Outbound shipment', 'OUTBOUND','Packed inventory shipped.', 'PER_UNIT', 'EA'),
        ('STORE_DELIVERY',    'Store delivery',    'OUTBOUND','Shipment delivered to store.', 'PER_DOCUMENT', NULL),
        ('RETURN_HANDLING',   'Return handling',   'OUTBOUND','Failed delivery returned to depot.', 'PER_UNIT', 'EA'),
        ('DISPOSAL',          'Disposal handling', 'INBOUND', 'Rejected stock disposed.', 'PER_UNIT', 'EA')
)
INSERT INTO billing_service (
    code, name, module_code, description,
    default_charge_basis_id, default_uom_id
)
SELECT services.code, services.name, services.module_code, services.description,
       basis.charge_basis_id, unit.uom_id
FROM services
JOIN charge_basis basis ON basis.code = services.basis_code
LEFT JOIN uom unit ON unit.code = services.uom_code
ON CONFLICT (code) DO NOTHING;

WITH service_movement(service_code, movement_code) AS (
    VALUES
        ('INBOUND_RECEIVING', 'RECEIVE'),
        ('PUTAWAY',           'PUTAWAY'),
        ('INTERNAL_MOVE',     'INTERNAL_MOVE'),
        ('WAREHOUSE_TRANSFER','TRANSFER_OUT'),
        ('OUTBOUND_PICK',     'PICK'),
        ('OUTBOUND_PACK',     'PACK'),
        ('OUTBOUND_SHIP',     'SHIP'),
        ('RETURN_HANDLING',   'DELIVERY_RETURN'),
        ('DISPOSAL',          'DISPOSE')
)
INSERT INTO billing_service_movement_type (
    billing_service_id, movement_type_id
)
SELECT service.billing_service_id, movement.movement_type_id
FROM service_movement mapping
JOIN billing_service service ON service.code = mapping.service_code
JOIN movement_type movement ON movement.code = mapping.movement_code
ON CONFLICT (billing_service_id, movement_type_id) DO NOTHING;

INSERT INTO billing_adjustment_type (code, name, amount_effect)
VALUES
    ('SURCHARGE', 'Surcharge',  1),
    ('DISCOUNT',  'Discount',  -1)
ON CONFLICT (code) DO NOTHING;

INSERT INTO payment_method (code, name, requires_reference)
VALUES
    ('BANK_TRANSFER', 'Bank transfer', true),
    ('VIRTUAL_ACCOUNT','Virtual account', true),
    ('CASH',          'Cash', false),
    ('OTHER',         'Other', true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO internal_move_type (code, name, description, requires_approval)
VALUES
    ('AD_HOC',        'Ad hoc movement', 'Operator-requested movement between locations.', false),
    ('REPLENISHMENT', 'Replenishment',   'Move reserve stock into a forward pick location.', false),
    ('CONSOLIDATION', 'Consolidation',   'Combine compatible stock into fewer locations or handling units.', false),
    ('RELOCATION',    'Relocation',      'Planned relocation of stock within a warehouse.', true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO inventory_adjustment_type (
    code, name, description, quantity_effect, requires_approval
)
VALUES
    ('INCREASE', 'Increase stock', 'Increase on-hand stock after an authorized investigation.',  1, true),
    ('DECREASE', 'Decrease stock', 'Decrease on-hand stock after an authorized investigation.', -1, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO stock_count_type (code, name, description, requires_freeze)
VALUES
    ('CYCLE',    'Cycle count',    'Recurring count of selected locations or items.', false),
    ('FULL',     'Full count',     'Count all in-scope warehouse stock.', true),
    ('LOCATION', 'Location count', 'Count selected warehouse locations.', false),
    ('ITEM',     'Item count',     'Count selected items across in-scope locations.', false)
ON CONFLICT (code) DO NOTHING;

UPDATE quarantine_disposition_type disposition
SET removal_movement_type_id = movement.movement_type_id
FROM movement_type movement
WHERE (disposition.code = 'RETURN' AND movement.code = 'RETURN_TO_VENDOR')
   OR (disposition.code = 'DISPOSE' AND movement.code = 'DISPOSE');

INSERT INTO reason_code (module_code, code, name, description, requires_note)
VALUES
    ('INVENTORY', 'DAMAGE',          'Damage',          'Stock was damaged.', true),
    ('INVENTORY', 'EXPIRY',          'Expiry',          'Stock expired.', false),
    ('INVENTORY', 'COUNT_VARIANCE',  'Count variance',  'System and physical counts differed.', true),
    ('INBOUND',   'OVER_RECEIPT',    'Over receipt',    'Received quantity exceeded expectation.', true),
    ('INBOUND',   'UNDER_RECEIPT',   'Under receipt',   'Received quantity was below expectation.', true),
    ('OUTBOUND',  'SHORT_PICK',      'Short pick',      'Picked quantity was below requirement.', true),
    ('GENERAL',   'CANCELLATION',    'Cancellation',    'Document or task was cancelled.', true)
ON CONFLICT (module_code, code) DO NOTHING;

INSERT INTO reason_code (module_code, code, name, description, requires_note)
VALUES
    ('BILLING', 'MANUAL_ADJUSTMENT', 'Manual adjustment', 'Authorized invoice adjustment.', true),
    ('BILLING', 'SERVICE_DISPUTE',   'Service dispute',   'Client disputed a billed service.', true),
    ('BILLING', 'BILLING_ERROR',     'Billing error',     'Correction of a billing error.', true),
    ('BILLING', 'GOODWILL_CREDIT',   'Goodwill credit',   'Commercial goodwill credit.', true),
    ('BILLING', 'EVENT_EXCLUDED',    'Event excluded',    'Billable event intentionally excluded.', true),
    ('BILLING', 'PAYMENT_REVERSAL',  'Payment reversal',  'Previously received payment was reversed.', true)
ON CONFLICT (module_code, code) DO NOTHING;

INSERT INTO reason_code (module_code, code, name, description, requires_note)
VALUES
    ('INVENTORY', 'REPLENISHMENT',   'Replenishment',    'Stock replenished into a forward location.', false),
    ('INVENTORY', 'RELOCATION',      'Relocation',       'Stock relocated within a warehouse.', false),
    ('INVENTORY', 'CONSOLIDATION',   'Consolidation',    'Compatible stock consolidated.', false),
    ('INVENTORY', 'STATUS_HOLD',     'Place on hold',    'Stock placed on a non-allocatable hold status.', true),
    ('INVENTORY', 'STATUS_RELEASE',  'Release hold',     'Stock released from hold.', true),
    ('INVENTORY', 'MANUAL_ADJUSTMENT','Manual adjustment','Authorized manual stock adjustment.', true),
    ('INVENTORY', 'TRANSFER_VARIANCE','Transfer variance','Difference found during transfer receipt.', true)
ON CONFLICT (module_code, code) DO NOTHING;

INSERT INTO reason_code (module_code, code, name, description, requires_note)
VALUES
    ('OUTBOUND', 'ALLOCATION_SHORT', 'Allocation shortage', 'Insufficient allocatable inventory.', true),
    ('OUTBOUND', 'CHECK_FAILED',     'Checking failed',     'Outbound checking found a discrepancy.', true),
    ('OUTBOUND', 'PACKING_DAMAGE',   'Packing damage',      'Damage was found during packing.', true),
    ('OUTBOUND', 'DELIVERY_RETURN',  'Delivery return',     'Undelivered goods returned to the warehouse.', true)
ON CONFLICT (module_code, code) DO NOTHING;

-- Task configuration -----------------------------------------------------------

INSERT INTO task_type (code, name, description)
VALUES
    ('PUTAWAY',       'Putaway',       'Move received stock into storage.'),
    ('PICK',          'Pick',          'Pick stock for outbound fulfillment.'),
    ('REPLENISHMENT', 'Replenishment', 'Move stock into a picking location.'),
    ('STOCK_COUNT',   'Stock count',   'Count stock at a warehouse location.'),
    ('REWORK',        'Rework',        'Rework quarantined inbound stock.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO task_status (code, name, is_initial, is_final, is_cancelled)
VALUES
    ('OPEN',        'Open',        true,  false, false),
    ('ASSIGNED',    'Assigned',    false, false, false),
    ('IN_PROGRESS', 'In progress', false, false, false),
    ('COMPLETED',   'Completed',   false, true,  false),
    ('CANCELLED',   'Cancelled',   false, true,  true),
    ('REVERSED',    'Reversed',    false, true,  false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO task_priority (code, name, priority_value)
VALUES
    ('LOW',    'Low',    100),
    ('NORMAL', 'Normal', 200),
    ('HIGH',   'High',   300),
    ('URGENT', 'Urgent', 400)
ON CONFLICT (code) DO NOTHING;

INSERT INTO picking_sort_method (code, name, description)
VALUES
    ('FEFO',     'First expiry, first out', 'Pick the stock with the earliest expiry date first.'),
    ('FIFO',     'First in, first out',     'Pick the oldest received stock first.'),
    ('LOCATION', 'Location sequence',       'Pick according to warehouse location sequence.'),
    ('LOT',      'Lot sequence',            'Pick according to lot-number sequence.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO picking_strategy (
    owner_id, warehouse_id, code, name, description, is_active
)
VALUES (
    NULL, NULL, 'DEFAULT_FEFO', 'Default FEFO',
    'Prefer allocatable stock with the earliest expiry date.', true
)
ON CONFLICT (owner_id, warehouse_id, code) DO NOTHING;

INSERT INTO picking_strategy_rule (
    picking_strategy_id, sequence_no, inventory_status_id,
    zone_id, picking_sort_method_id, is_active
)
SELECT strategy.picking_strategy_id, 10, status.inventory_status_id,
       NULL, sort_method.picking_sort_method_id, true
FROM picking_strategy strategy
JOIN inventory_status status ON status.code = 'AVAILABLE'
JOIN picking_sort_method sort_method ON sort_method.code = 'FEFO'
WHERE strategy.owner_id IS NULL
  AND strategy.warehouse_id IS NULL
  AND strategy.code = 'DEFAULT_FEFO'
ON CONFLICT (picking_strategy_id, sequence_no) DO NOTHING;

-- Document types and number formats -------------------------------------------

INSERT INTO document_type (code, name, module_code, description)
VALUES
    ('PURCHASE_ORDER',    'Purchase order',     'INBOUND',   'Client purchase order expected by the 3PL.'),
    ('INBOUND',            'Inbound order',      'INBOUND',   'Expected inbound goods.'),
    ('RECEIPT',            'Receipt',            'INBOUND',   'Physical receipt of goods.'),
    ('QUALITY_INSPECTION', 'Quality inspection', 'INBOUND',   'Inbound quality inspection.'),
    ('PUTAWAY_TASK',       'Putaway task',       'INBOUND',   'Inbound putaway work.'),
    ('OUTBOUND',           'Outbound order',     'OUTBOUND',  'Outbound fulfillment order.'),
    ('RESERVATION',        'Reservation',        'OUTBOUND',  'Inventory reservation.'),
    ('PICK_TASK',          'Pick task',          'OUTBOUND',  'Outbound picking work.'),
    ('PACKING',            'Packing',            'OUTBOUND',  'Packed outbound goods.'),
    ('SHIPMENT',           'Shipment',           'OUTBOUND',  'Shipment from a warehouse.'),
    ('OUTBOUND_VALIDATION','Outbound validation','OUTBOUND',  'Recorded validation of a client delivery order.'),
    ('OUTBOUND_WAVE',      'Outbound wave',      'OUTBOUND',  'Group of allocated orders released for picking.'),
    ('OUTBOUND_STAGING',   'Outbound staging',   'OUTBOUND',  'Picked stock staged by order.'),
    ('OUTBOUND_CHECK',     'Outbound check',     'OUTBOUND',  'Quantity and item verification before packing.'),
    ('DELIVERY',           'Store delivery',     'OUTBOUND',  'Shipment delivery and proof of delivery.'),
    ('BILLING_CONTRACT',   'Billing contract',   'BILLING',   'Commercial billing agreement.'),
    ('RATE_CARD',          'Rate card',          'BILLING',   'Versioned service pricing.'),
    ('BILLABLE_EVENT',     'Billable event',     'BILLING',   'Immutable operational billing event.'),
    ('BILLING_RUN',        'Billing run',        'BILLING',   'Rated events for a billing period.'),
    ('BILLING_CHARGE',     'Billing charge',     'BILLING',   'Calculated event charge snapshot.'),
    ('INVOICE',            'Invoice',            'BILLING',   'Client billing invoice.'),
    ('CREDIT_NOTE',        'Credit note',        'BILLING',   'Issued invoice credit.'),
    ('PAYMENT',            'Payment',            'BILLING',   'Client payment receipt.'),
        ('TRANSFER',           'Transfer order',     'INVENTORY', 'Inter-warehouse transfer.'),
        ('INTERNAL_MOVE',      'Internal movement',  'INVENTORY', 'Movement inside one warehouse.'),
        ('TRANSFER_DISPATCH',  'Transfer dispatch',  'INVENTORY', 'Dispatch from a source warehouse.'),
        ('TRANSFER_RECEIPT',   'Transfer receipt',   'INVENTORY', 'Receipt at a target warehouse.'),
        ('INVENTORY_STATUS_CHANGE', 'Inventory status change', 'INVENTORY', 'Authorized stock status change.'),
        ('INVENTORY_ADJUSTMENT', 'Inventory adjustment', 'INVENTORY', 'Authorized on-hand quantity adjustment.'),
        ('STOCK_COUNT',        'Stock count',        'INVENTORY', 'Physical inventory count.'),
    ('MOVEMENT',           'Inventory movement', 'INVENTORY', 'Inventory ledger movement.'),
    ('INVENTORY_BALANCE',  'Inventory balance',  'INVENTORY', 'Current stock balance identity.'),
    ('HANDLING_UNIT',      'Handling unit',      'INVENTORY', 'Pallet, carton, tote, or bin.'),
    ('LOT',                'Inventory lot',      'INVENTORY', 'Lot-controlled stock identity.'),
    ('SERIAL',             'Serial number',      'INVENTORY', 'Serial-controlled stock identity.'),
    ('QUARANTINE_CASE',    'Quarantine case',    'INBOUND',   'Client disposition case for rejected stock.'),
    ('QUARANTINE_DISPOSITION', 'Quarantine disposition', 'INBOUND', 'Client decision for quarantined stock.')
ON CONFLICT (code) DO NOTHING;

-- One configured number rule per type. The counter itself resets per type/day.
WITH rules(document_type_code, prefix, include_partner, include_warehouse) AS (
    VALUES
        ('PURCHASE_ORDER',      'PO',  true,  true),
        ('INBOUND',            'INB', true,  true),
        ('RECEIPT',            'RCV', true,  true),
        ('QUALITY_INSPECTION', 'QIN', false, true),
        ('PUTAWAY_TASK',       'PUT', false, true),
        ('OUTBOUND',           'OUT', true,  true),
        ('RESERVATION',        'RSV', false, true),
        ('PICK_TASK',          'PCK', false, true),
        ('PACKING',            'PKG', false, true),
        ('SHIPMENT',           'SHP', false, true),
        ('OUTBOUND_VALIDATION','OVL', false, true),
        ('OUTBOUND_WAVE',      'WAV', false, true),
        ('OUTBOUND_STAGING',   'STG', false, true),
        ('OUTBOUND_CHECK',     'CHK', false, true),
        ('DELIVERY',           'DLV', false, true),
        ('BILLING_CONTRACT',   'BCT', false, false),
        ('RATE_CARD',          'RAT', false, false),
        ('BILLABLE_EVENT',     'BEV', false, true),
        ('BILLING_RUN',        'BRN', false, false),
        ('BILLING_CHARGE',     'CHG', false, false),
        ('INVOICE',            'INV', false, false),
        ('CREDIT_NOTE',        'CRN', false, false),
        ('PAYMENT',            'PAY', false, false),
        ('TRANSFER',           'TRF', false, true),
        ('INTERNAL_MOVE',      'IMV', false, true),
        ('TRANSFER_DISPATCH',  'TDS', false, true),
        ('TRANSFER_RECEIPT',   'TRC', false, true),
        ('INVENTORY_STATUS_CHANGE', 'ISC', false, true),
        ('INVENTORY_ADJUSTMENT', 'ADJ', false, true),
        ('STOCK_COUNT',        'CNT', false, true),
        ('MOVEMENT',           'MOV', false, true),
        ('INVENTORY_BALANCE',  'BAL', false, true),
        ('HANDLING_UNIT',      'HU',  false, true),
        ('LOT',                'LOT', false, false),
        ('SERIAL',             'SER', false, false),
        ('QUARANTINE_CASE',    'QCS', false, true),
        ('QUARANTINE_DISPOSITION', 'QDS', false, true)
)
INSERT INTO document_number_rule (
    document_type_id,
    prefix,
    separator,
    sequence_length,
    include_partner_code,
    include_warehouse_code,
    is_active
)
SELECT
    dt.document_type_id,
    rules.prefix,
    '-',
    6,
    rules.include_partner,
    rules.include_warehouse,
    true
FROM rules
JOIN document_type dt ON dt.code = rules.document_type_code
WHERE NOT EXISTS (
    SELECT 1
    FROM document_number_rule existing
    WHERE existing.document_type_id = dt.document_type_id
      AND existing.is_active
);

-- Document statuses. These are starter values and remain editable master data.
WITH statuses(
    document_type_code, status_code, status_name,
    is_initial, is_final, is_cancelled, display_order
) AS (
    VALUES
        ('PURCHASE_ORDER','DRAFT',              'Draft',              true,  false, false, 10),
        ('PURCHASE_ORDER','APPROVED',           'Approved',           false, false, false, 20),
        ('PURCHASE_ORDER','PARTIALLY_RECEIVED', 'Partially received', false, false, false, 30),
        ('PURCHASE_ORDER','RECEIVED',           'Received',           false, true,  false, 40),
        ('PURCHASE_ORDER','CLOSED',             'Closed short',       false, true,  false, 50),
        ('PURCHASE_ORDER','CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('INBOUND',     'DRAFT',              'Draft',              true,  false, false, 10),
        ('INBOUND',     'RELEASED',           'Released',           false, false, false, 20),
        ('INBOUND',     'PARTIALLY_RECEIVED', 'Partially received', false, false, false, 30),
        ('INBOUND',     'RECEIVED',           'Received',           false, true,  false, 40),
        ('INBOUND',     'CLOSED',             'Closed short',       false, true,  false, 50),
        ('INBOUND',     'CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('RECEIPT',     'OPEN',               'Open',               true,  false, false, 10),
        ('RECEIPT',     'COMPLETED',          'Completed',          false, true,  false, 20),
        ('RECEIPT',     'REVERSED',           'Reversed',           false, true,  false, 30),
        ('RECEIPT',     'CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('OUTBOUND',    'DRAFT',              'Draft',              true,  false, false, 10),
        ('OUTBOUND',    'VALIDATED',          'Validated',          false, false, false, 15),
        ('OUTBOUND',    'RELEASED',           'Released',           false, false, false, 20),
        ('OUTBOUND',    'PARTIALLY_ALLOCATED','Partially allocated', false, false, false, 25),
        ('OUTBOUND',    'ALLOCATED',          'Allocated',          false, false, false, 30),
        ('OUTBOUND',    'WAVED',              'Waved',              false, false, false, 35),
        ('OUTBOUND',    'PICKING',            'Picking',            false, false, false, 40),
        ('OUTBOUND',    'STAGED',             'Staged',             false, false, false, 45),
        ('OUTBOUND',    'CHECKING',           'Checking',           false, false, false, 50),
        ('OUTBOUND',    'CHECK_FAILED',       'Check failed',       false, false, false, 55),
        ('OUTBOUND',    'CHECKED',            'Checked',            false, false, false, 60),
        ('OUTBOUND',    'PACKING',            'Packing',            false, false, false, 65),
        ('OUTBOUND',    'PACKED',             'Packed',             false, false, false, 70),
        ('OUTBOUND',    'SHIPPED',            'Shipped',            false, false, false, 75),
        ('OUTBOUND',    'DELIVERED',          'Delivered',          false, true,  false, 80),
        ('OUTBOUND',    'DELIVERY_FAILED',    'Delivery failed',    false, false, false, 85),
        ('OUTBOUND',    'PARTIALLY_DELIVERED','Partially delivered',false, true,  false, 87),
        ('OUTBOUND',    'RETURNED',           'Returned',           false, true,  false, 88),
        ('OUTBOUND',    'CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('OUTBOUND_VALIDATION','PENDING',     'Pending',            true,  false, false, 10),
        ('OUTBOUND_VALIDATION','PASSED',      'Passed',             false, true,  false, 20),
        ('OUTBOUND_VALIDATION','FAILED',      'Failed',             false, true,  false, 30),
        ('RESERVATION', 'ACTIVE',             'Active',             true,  false, false, 10),
        ('RESERVATION', 'PARTIALLY_PICKED',   'Partially picked',   false, false, false, 20),
        ('RESERVATION', 'CONSUMED',           'Consumed',           false, true,  false, 30),
        ('RESERVATION', 'RELEASED',           'Released',           false, true,  true,  90),
        ('OUTBOUND_WAVE','DRAFT',             'Draft',              true,  false, false, 10),
        ('OUTBOUND_WAVE','RELEASED',          'Released',           false, false, false, 20),
        ('OUTBOUND_WAVE','IN_PROGRESS',       'In progress',        false, false, false, 30),
        ('OUTBOUND_WAVE','COMPLETED',         'Completed',          false, true,  false, 40),
        ('OUTBOUND_WAVE','CANCELLED',         'Cancelled',          false, true,  true,  90),
        ('OUTBOUND_STAGING','OPEN',           'Open',               true,  false, false, 10),
        ('OUTBOUND_STAGING','COMPLETED',      'Completed',          false, true,  false, 20),
        ('OUTBOUND_STAGING','CANCELLED',      'Cancelled',          false, true,  true,  90),
        ('OUTBOUND_CHECK','OPEN',             'Open',               true,  false, false, 10),
        ('OUTBOUND_CHECK','PASSED',           'Passed',             false, true,  false, 20),
        ('OUTBOUND_CHECK','FAILED',           'Failed',             false, true,  false, 30),
        ('OUTBOUND_CHECK','CANCELLED',        'Cancelled',          false, true,  true,  90),
        ('PACKING',     'OPEN',               'Open',               true,  false, false, 10),
        ('PACKING',     'COMPLETED',          'Completed',          false, true,  false, 20),
        ('PACKING',     'CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('SHIPMENT',    'PLANNED',            'Planned',            true,  false, false, 10),
        ('SHIPMENT',    'SHIPPED',            'Shipped',            false, true,  false, 20),
        ('SHIPMENT',    'CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('DELIVERY',    'PLANNED',            'Planned',            true,  false, false, 10),
        ('DELIVERY',    'IN_TRANSIT',         'In transit',         false, false, false, 20),
        ('DELIVERY',    'ARRIVED',            'Arrived',            false, false, false, 30),
        ('DELIVERY',    'FAILED',             'Failed',             false, false, false, 40),
        ('DELIVERY',    'DELIVERED',          'Delivered',          false, true,  false, 50),
        ('DELIVERY',    'PARTIALLY_DELIVERED','Partially delivered',false, true,  false, 55),
        ('DELIVERY',    'RETURNED',           'Returned',           false, true,  false, 60),
        ('DELIVERY',    'CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('BILLING_CONTRACT','DRAFT',          'Draft',              true,  false, false, 10),
        ('BILLING_CONTRACT','ACTIVE',         'Active',             false, false, false, 20),
        ('BILLING_CONTRACT','SUSPENDED',      'Suspended',          false, false, false, 30),
        ('BILLING_CONTRACT','EXPIRED',        'Expired',            false, true,  false, 40),
        ('BILLING_CONTRACT','TERMINATED',     'Terminated',         false, true,  true,  90),
        ('RATE_CARD',    'DRAFT',             'Draft',              true,  false, false, 10),
        ('RATE_CARD',    'APPROVED',          'Approved',           false, false, false, 20),
        ('RATE_CARD',    'ACTIVE',            'Active',             false, false, false, 30),
        ('RATE_CARD',    'EXPIRED',           'Expired',            false, true,  false, 40),
        ('RATE_CARD',    'CANCELLED',         'Cancelled',          false, true,  true,  90),
        ('BILLABLE_EVENT','PENDING',          'Pending',            true,  false, false, 10),
        ('BILLABLE_EVENT','RATED',            'Rated',              false, true,  false, 20),
        ('BILLABLE_EVENT','EXCLUDED',         'Excluded',           false, true,  true,  80),
        ('BILLABLE_EVENT','CANCELLED',        'Cancelled',          false, true,  true,  90),
        ('BILLING_RUN',  'DRAFT',             'Draft',              true,  false, false, 10),
        ('BILLING_RUN',  'CALCULATED',        'Calculated',         false, false, false, 20),
        ('BILLING_RUN',  'REVIEWED',          'Reviewed',           false, false, false, 30),
        ('BILLING_RUN',  'INVOICED',          'Invoiced',           false, true,  false, 40),
        ('BILLING_RUN',  'CANCELLED',         'Cancelled',          false, true,  true,  90),
        ('INVOICE',      'DRAFT',             'Draft',              true,  false, false, 10),
        ('INVOICE',      'REVIEWED',          'Reviewed',           false, false, false, 20),
        ('INVOICE',      'ISSUED',            'Issued',             false, false, false, 30),
        ('INVOICE',      'PARTIALLY_PAID',    'Partially paid',     false, false, false, 40),
        ('INVOICE',      'PAID',              'Paid',               false, true,  false, 50),
        ('INVOICE',      'SETTLED',           'Settled by credit',  false, true,  false, 60),
        ('INVOICE',      'VOID',              'Void',               false, true,  true,  90),
        ('CREDIT_NOTE',  'DRAFT',             'Draft',              true,  false, false, 10),
        ('CREDIT_NOTE',  'ISSUED',            'Issued',             false, true,  false, 20),
        ('CREDIT_NOTE',  'VOID',              'Void',               false, true,  true,  90),
        ('PAYMENT',      'RECEIVED',          'Received',           true,  false, false, 10),
        ('PAYMENT',      'PARTIALLY_ALLOCATED','Partially allocated',false,false,false,20),
        ('PAYMENT',      'ALLOCATED',         'Allocated',          false, true,  false, 30),
        ('PAYMENT',      'VOID',              'Void',               false, true,  true,  90),
        ('TRANSFER',    'DRAFT',              'Draft',              true,  false, false, 10),
        ('TRANSFER',    'APPROVED',           'Approved',           false, false, false, 20),
        ('TRANSFER',    'PARTIALLY_DISPATCHED','Partially dispatched',false,false,false,30),
        ('TRANSFER',    'IN_TRANSIT',         'In transit',         false, false, false, 40),
        ('TRANSFER',    'PARTIALLY_RECEIVED', 'Partially received', false, false, false, 50),
        ('TRANSFER',    'RECEIVED',           'Received',           false, true,  false, 60),
        ('TRANSFER',    'CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('INTERNAL_MOVE','DRAFT',             'Draft',              true,  false, false, 10),
        ('INTERNAL_MOVE','APPROVED',          'Approved',           false, false, false, 20),
        ('INTERNAL_MOVE','IN_PROGRESS',       'In progress',        false, false, false, 30),
        ('INTERNAL_MOVE','COMPLETED',         'Completed',          false, true,  false, 40),
        ('INTERNAL_MOVE','CANCELLED',         'Cancelled',          false, true,  true,  90),
        ('TRANSFER_DISPATCH','DRAFT',         'Draft',              true,  false, false, 10),
        ('TRANSFER_DISPATCH','DISPATCHED',    'Dispatched',         false, true,  false, 20),
        ('TRANSFER_DISPATCH','CANCELLED',     'Cancelled',          false, true,  true,  90),
        ('TRANSFER_RECEIPT','DRAFT',          'Draft',              true,  false, false, 10),
        ('TRANSFER_RECEIPT','RECEIVED',       'Received',           false, true,  false, 20),
        ('TRANSFER_RECEIPT','CANCELLED',      'Cancelled',          false, true,  true,  90),
        ('INVENTORY_STATUS_CHANGE','DRAFT',   'Draft',              true,  false, false, 10),
        ('INVENTORY_STATUS_CHANGE','APPROVED','Approved',           false, false, false, 20),
        ('INVENTORY_STATUS_CHANGE','POSTED',  'Posted',             false, true,  false, 30),
        ('INVENTORY_STATUS_CHANGE','CANCELLED','Cancelled',         false, true,  true,  90),
        ('INVENTORY_ADJUSTMENT','DRAFT',      'Draft',              true,  false, false, 10),
        ('INVENTORY_ADJUSTMENT','APPROVED',   'Approved',           false, false, false, 20),
        ('INVENTORY_ADJUSTMENT','POSTED',     'Posted',             false, true,  false, 30),
        ('INVENTORY_ADJUSTMENT','CANCELLED',  'Cancelled',          false, true,  true,  90),
        ('STOCK_COUNT', 'DRAFT',              'Draft',              true,  false, false, 10),
        ('STOCK_COUNT', 'COUNTING',           'Counting',           false, false, false, 20),
        ('STOCK_COUNT', 'REVIEW',             'Review',             false, false, false, 30),
        ('STOCK_COUNT', 'POSTED',             'Posted',             false, true,  false, 40),
        ('STOCK_COUNT', 'CANCELLED',          'Cancelled',          false, true,  true,  90),
        ('QUARANTINE_CASE', 'OPEN',           'Open',               true,  false, false, 10),
        ('QUARANTINE_CASE', 'PARTIALLY_DECIDED','Partially decided', false, false, false, 20),
        ('QUARANTINE_CASE', 'CLOSED',         'Closed',             false, true,  false, 30),
        ('QUARANTINE_DISPOSITION', 'DECIDED', 'Decided',            true,  false, false, 10),
        ('QUARANTINE_DISPOSITION', 'PROCESSED','Processed',         false, true,  false, 20),
        ('QUARANTINE_DISPOSITION', 'CANCELLED','Cancelled',         false, true,  true,  90)
)
INSERT INTO document_status (
    document_type_id,
    code,
    name,
    is_initial,
    is_final,
    is_cancelled,
    display_order
)
SELECT
    dt.document_type_id,
    statuses.status_code,
    statuses.status_name,
    statuses.is_initial,
    statuses.is_final,
    statuses.is_cancelled,
    statuses.display_order
FROM statuses
JOIN document_type dt ON dt.code = statuses.document_type_code
ON CONFLICT (document_type_id, code) DO NOTHING;

WITH transitions(document_type_code, from_code, to_code) AS (
    VALUES
        ('PURCHASE_ORDER','DRAFT',              'APPROVED'),
        ('PURCHASE_ORDER','DRAFT',              'CANCELLED'),
        ('PURCHASE_ORDER','APPROVED',           'PARTIALLY_RECEIVED'),
        ('PURCHASE_ORDER','APPROVED',           'RECEIVED'),
        ('PURCHASE_ORDER','APPROVED',           'CANCELLED'),
        ('PURCHASE_ORDER','PARTIALLY_RECEIVED', 'RECEIVED'),
        ('PURCHASE_ORDER','APPROVED',           'CLOSED'),
        ('PURCHASE_ORDER','PARTIALLY_RECEIVED', 'CLOSED'),
        ('INBOUND',     'DRAFT',              'RELEASED'),
        ('INBOUND',     'DRAFT',              'CANCELLED'),
        ('INBOUND',     'RELEASED',           'PARTIALLY_RECEIVED'),
        ('INBOUND',     'RELEASED',           'RECEIVED'),
        ('INBOUND',     'RELEASED',           'CANCELLED'),
        ('INBOUND',     'PARTIALLY_RECEIVED', 'RECEIVED'),
        ('INBOUND',     'PARTIALLY_RECEIVED', 'CANCELLED'),
        ('INBOUND',     'RELEASED',           'CLOSED'),
        ('INBOUND',     'PARTIALLY_RECEIVED', 'CLOSED'),
        ('RECEIPT',     'OPEN',               'COMPLETED'),
        ('RECEIPT',     'OPEN',               'CANCELLED'),
        ('OUTBOUND',    'DRAFT',              'VALIDATED'),
        ('OUTBOUND',    'DRAFT',              'CANCELLED'),
        ('OUTBOUND',    'VALIDATED',          'RELEASED'),
        ('OUTBOUND',    'VALIDATED',          'CANCELLED'),
        ('OUTBOUND',    'RELEASED',           'PARTIALLY_ALLOCATED'),
        ('OUTBOUND',    'RELEASED',           'ALLOCATED'),
        ('OUTBOUND',    'RELEASED',           'CANCELLED'),
        ('OUTBOUND',    'PARTIALLY_ALLOCATED','ALLOCATED'),
        ('OUTBOUND',    'PARTIALLY_ALLOCATED','CANCELLED'),
        ('OUTBOUND',    'ALLOCATED',          'WAVED'),
        ('OUTBOUND',    'ALLOCATED',          'CANCELLED'),
        ('OUTBOUND',    'WAVED',              'PICKING'),
        ('OUTBOUND',    'PICKING',            'STAGED'),
        ('OUTBOUND',    'STAGED',             'CHECKING'),
        ('OUTBOUND',    'CHECKING',           'CHECKED'),
        ('OUTBOUND',    'CHECKING',           'CHECK_FAILED'),
        ('OUTBOUND',    'CHECK_FAILED',       'CHECKING'),
        ('OUTBOUND',    'CHECKED',            'PACKING'),
        ('OUTBOUND',    'PACKING',            'PACKED'),
        ('OUTBOUND',    'PACKED',             'SHIPPED'),
        ('OUTBOUND',    'SHIPPED',            'DELIVERED'),
        ('OUTBOUND',    'SHIPPED',            'DELIVERY_FAILED'),
        ('OUTBOUND',    'DELIVERY_FAILED',    'DELIVERED'),
        ('OUTBOUND',    'DELIVERY_FAILED',    'PARTIALLY_DELIVERED'),
        ('OUTBOUND',    'DELIVERY_FAILED',    'RETURNED'),
        ('OUTBOUND_VALIDATION','PENDING',     'PASSED'),
        ('OUTBOUND_VALIDATION','PENDING',     'FAILED'),
        ('RESERVATION', 'ACTIVE',             'PARTIALLY_PICKED'),
        ('RESERVATION', 'ACTIVE',             'CONSUMED'),
        ('RESERVATION', 'ACTIVE',             'RELEASED'),
        ('RESERVATION', 'PARTIALLY_PICKED',   'CONSUMED'),
        ('RESERVATION', 'PARTIALLY_PICKED',   'RELEASED'),
        ('OUTBOUND_WAVE','DRAFT',             'RELEASED'),
        ('OUTBOUND_WAVE','DRAFT',             'CANCELLED'),
        ('OUTBOUND_WAVE','RELEASED',          'IN_PROGRESS'),
        ('OUTBOUND_WAVE','RELEASED',          'COMPLETED'),
        ('OUTBOUND_WAVE','IN_PROGRESS',       'COMPLETED'),
        ('OUTBOUND_STAGING','OPEN',           'COMPLETED'),
        ('OUTBOUND_STAGING','OPEN',           'CANCELLED'),
        ('OUTBOUND_CHECK','OPEN',             'PASSED'),
        ('OUTBOUND_CHECK','OPEN',             'FAILED'),
        ('OUTBOUND_CHECK','OPEN',             'CANCELLED'),
        ('PACKING',     'OPEN',               'COMPLETED'),
        ('PACKING',     'OPEN',               'CANCELLED'),
        ('SHIPMENT',    'PLANNED',            'SHIPPED'),
        ('SHIPMENT',    'PLANNED',            'CANCELLED'),
        ('DELIVERY',    'PLANNED',            'IN_TRANSIT'),
        ('DELIVERY',    'PLANNED',            'CANCELLED'),
        ('DELIVERY',    'IN_TRANSIT',         'ARRIVED'),
        ('DELIVERY',    'IN_TRANSIT',         'FAILED'),
        ('DELIVERY',    'ARRIVED',            'DELIVERED'),
        ('DELIVERY',    'ARRIVED',            'FAILED'),
        ('DELIVERY',    'FAILED',             'IN_TRANSIT'),
        ('DELIVERY',    'FAILED',             'DELIVERED'),
        ('DELIVERY',    'FAILED',             'PARTIALLY_DELIVERED'),
        ('DELIVERY',    'FAILED',             'RETURNED'),
        ('BILLING_CONTRACT','DRAFT',          'ACTIVE'),
        ('BILLING_CONTRACT','DRAFT',          'TERMINATED'),
        ('BILLING_CONTRACT','ACTIVE',         'SUSPENDED'),
        ('BILLING_CONTRACT','ACTIVE',         'EXPIRED'),
        ('BILLING_CONTRACT','ACTIVE',         'TERMINATED'),
        ('BILLING_CONTRACT','SUSPENDED',      'ACTIVE'),
        ('BILLING_CONTRACT','SUSPENDED',      'TERMINATED'),
        ('RATE_CARD',    'DRAFT',             'APPROVED'),
        ('RATE_CARD',    'DRAFT',             'CANCELLED'),
        ('RATE_CARD',    'APPROVED',          'ACTIVE'),
        ('RATE_CARD',    'APPROVED',          'CANCELLED'),
        ('RATE_CARD',    'ACTIVE',            'EXPIRED'),
        ('BILLABLE_EVENT','PENDING',          'RATED'),
        ('BILLABLE_EVENT','PENDING',          'EXCLUDED'),
        ('BILLABLE_EVENT','PENDING',          'CANCELLED'),
        ('BILLING_RUN',  'DRAFT',             'CALCULATED'),
        ('BILLING_RUN',  'DRAFT',             'CANCELLED'),
        ('BILLING_RUN',  'CALCULATED',        'REVIEWED'),
        ('BILLING_RUN',  'CALCULATED',        'DRAFT'),
        ('BILLING_RUN',  'CALCULATED',        'CANCELLED'),
        ('BILLING_RUN',  'REVIEWED',          'INVOICED'),
        ('INVOICE',      'DRAFT',             'REVIEWED'),
        ('INVOICE',      'DRAFT',             'VOID'),
        ('INVOICE',      'REVIEWED',          'ISSUED'),
        ('INVOICE',      'REVIEWED',          'DRAFT'),
        ('INVOICE',      'REVIEWED',          'VOID'),
        ('INVOICE',      'ISSUED',            'PARTIALLY_PAID'),
        ('INVOICE',      'ISSUED',            'PAID'),
        ('INVOICE',      'ISSUED',            'VOID'),
        ('INVOICE',      'ISSUED',            'SETTLED'),
        ('INVOICE',      'PARTIALLY_PAID',    'PAID'),
        ('INVOICE',      'PARTIALLY_PAID',    'SETTLED'),
        ('CREDIT_NOTE',  'DRAFT',             'ISSUED'),
        ('CREDIT_NOTE',  'DRAFT',             'VOID'),
        ('PAYMENT',      'RECEIVED',          'PARTIALLY_ALLOCATED'),
        ('PAYMENT',      'RECEIVED',          'ALLOCATED'),
        ('PAYMENT',      'RECEIVED',          'VOID'),
        ('PAYMENT',      'PARTIALLY_ALLOCATED','ALLOCATED'),
        ('PAYMENT',      'PARTIALLY_ALLOCATED','VOID'),
        ('TRANSFER',    'DRAFT',              'APPROVED'),
        ('TRANSFER',    'DRAFT',              'CANCELLED'),
        ('TRANSFER',    'APPROVED',           'PARTIALLY_DISPATCHED'),
        ('TRANSFER',    'APPROVED',           'IN_TRANSIT'),
        ('TRANSFER',    'APPROVED',           'CANCELLED'),
        ('TRANSFER',    'PARTIALLY_DISPATCHED','IN_TRANSIT'),
        ('TRANSFER',    'PARTIALLY_DISPATCHED','PARTIALLY_RECEIVED'),
        ('TRANSFER',    'PARTIALLY_DISPATCHED','CANCELLED'),
        ('TRANSFER',    'IN_TRANSIT',         'PARTIALLY_RECEIVED'),
        ('TRANSFER',    'IN_TRANSIT',         'RECEIVED'),
        ('TRANSFER',    'PARTIALLY_RECEIVED', 'RECEIVED'),
        ('INTERNAL_MOVE','DRAFT',             'APPROVED'),
        ('INTERNAL_MOVE','DRAFT',             'CANCELLED'),
        ('INTERNAL_MOVE','APPROVED',          'IN_PROGRESS'),
        ('INTERNAL_MOVE','APPROVED',          'COMPLETED'),
        ('INTERNAL_MOVE','APPROVED',          'CANCELLED'),
        ('INTERNAL_MOVE','IN_PROGRESS',       'COMPLETED'),
        ('INTERNAL_MOVE','IN_PROGRESS',       'CANCELLED'),
        ('TRANSFER_DISPATCH','DRAFT',         'DISPATCHED'),
        ('TRANSFER_DISPATCH','DRAFT',         'CANCELLED'),
        ('TRANSFER_RECEIPT','DRAFT',          'RECEIVED'),
        ('TRANSFER_RECEIPT','DRAFT',          'CANCELLED'),
        ('INVENTORY_STATUS_CHANGE','DRAFT',   'APPROVED'),
        ('INVENTORY_STATUS_CHANGE','DRAFT',   'CANCELLED'),
        ('INVENTORY_STATUS_CHANGE','APPROVED','POSTED'),
        ('INVENTORY_STATUS_CHANGE','APPROVED','CANCELLED'),
        ('INVENTORY_ADJUSTMENT','DRAFT',      'APPROVED'),
        ('INVENTORY_ADJUSTMENT','DRAFT',      'CANCELLED'),
        ('INVENTORY_ADJUSTMENT','APPROVED',   'POSTED'),
        ('INVENTORY_ADJUSTMENT','APPROVED',   'CANCELLED'),
        ('STOCK_COUNT', 'DRAFT',              'COUNTING'),
        ('STOCK_COUNT', 'DRAFT',              'CANCELLED'),
        ('STOCK_COUNT', 'COUNTING',           'REVIEW'),
        ('STOCK_COUNT', 'COUNTING',           'CANCELLED'),
        ('STOCK_COUNT', 'REVIEW',             'POSTED'),
        ('STOCK_COUNT', 'REVIEW',             'COUNTING'),
        ('STOCK_COUNT', 'REVIEW',             'CANCELLED'),
        ('QUARANTINE_CASE', 'OPEN',              'PARTIALLY_DECIDED'),
        ('QUARANTINE_CASE', 'OPEN',              'CLOSED'),
        ('QUARANTINE_CASE', 'PARTIALLY_DECIDED', 'CLOSED'),
        ('QUARANTINE_DISPOSITION', 'DECIDED',    'PROCESSED'),
        ('QUARANTINE_DISPOSITION', 'DECIDED',    'CANCELLED')
)
INSERT INTO document_status_transition (
    document_type_id, from_status_id, to_status_id
)
SELECT
    dt.document_type_id,
    from_status.status_id,
    to_status.status_id
FROM transitions
JOIN document_type dt ON dt.code = transitions.document_type_code
JOIN document_status from_status
  ON from_status.document_type_id = dt.document_type_id
 AND from_status.code = transitions.from_code
JOIN document_status to_status
  ON to_status.document_type_id = dt.document_type_id
 AND to_status.code = transitions.to_code
ON CONFLICT (document_type_id, from_status_id, to_status_id) DO NOTHING;

WITH task_transitions(from_code, to_code) AS (
    VALUES
        ('OPEN',        'ASSIGNED'),
        ('OPEN',        'IN_PROGRESS'),
        ('OPEN',        'CANCELLED'),
        ('ASSIGNED',    'IN_PROGRESS'),
        ('ASSIGNED',    'CANCELLED'),
        ('IN_PROGRESS', 'COMPLETED'),
        ('IN_PROGRESS', 'CANCELLED')
)
INSERT INTO task_status_transition (from_status_id, to_status_id)
SELECT from_status.task_status_id, to_status.task_status_id
FROM task_transitions
JOIN task_status from_status ON from_status.code = task_transitions.from_code
JOIN task_status to_status ON to_status.code = task_transitions.to_code
ON CONFLICT (from_status_id, to_status_id) DO NOTHING;

-- Basic administration permissions and menus ----------------------------------

INSERT INTO app_permission (code, name, module_code, description)
VALUES
    ('ACCOUNT.READ',       'View accounts',       'SECURITY', 'View account records.'),
    ('ACCOUNT.CREATE',     'Create accounts',     'SECURITY', 'Create account records.'),
    ('ACCOUNT.UPDATE',     'Update accounts',     'SECURITY', 'Update account records.'),
    ('ACCOUNT.DISABLE',    'Disable accounts',    'SECURITY', 'Disable account records.'),
    ('ROLE.MANAGE',        'Manage roles',        'SECURITY', 'Manage roles and permissions.'),
    ('MASTER.ORGANIZATION','Manage organizations','MASTER',   'Manage organization master data.'),
    ('MASTER.WAREHOUSE',   'Manage warehouses',   'MASTER',   'Manage warehouse and location master data.'),
    ('MASTER.PARTNER',     'Manage partners',     'MASTER',   'Manage business partner master data.'),
    ('MASTER.ITEM',        'Manage items',        'MASTER',   'Manage item and UOM master data.'),
    ('MASTER.CONFIG',      'Manage configuration','MASTER',   'Manage operational configuration.'),
    ('INBOUND.PO.READ',    'View purchase orders','INBOUND',  'View inbound purchase orders.'),
    ('INBOUND.PO.CREATE',  'Create purchase orders','INBOUND','Create inbound purchase orders.'),
    ('INBOUND.PO.APPROVE', 'Approve purchase orders','INBOUND','Approve inbound purchase orders.'),
    ('INBOUND.RECEIVE',    'Receive inbound goods','INBOUND', 'Create receipts and received stock.'),
    ('INBOUND.QC',         'Perform inbound QC',  'INBOUND',  'Inspect received stock.'),
    ('INBOUND.DISPOSITION','Decide quarantine stock','INBOUND','Record client quarantine decisions.'),
    ('INBOUND.REWORK',     'Perform rework',      'INBOUND',  'Perform and complete inbound rework.'),
    ('INBOUND.PUTAWAY',    'Perform putaway',     'INBOUND',  'Move accepted stock into storage.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_permission (code, name, module_code, description)
VALUES
    ('REPORT.MASTER',    'View master reports',        'REPORTING', 'View organization, warehouse, partner, item, and configuration reports.'),
    ('REPORT.INBOUND',   'View inbound reports',       'REPORTING', 'View PO, receiving, QC, quarantine, and putaway reports.'),
    ('REPORT.INVENTORY', 'View stock-control reports', 'REPORTING', 'View stock, movement, transfer, adjustment, and count reports.'),
    ('REPORT.OUTBOUND',  'View outbound reports',      'REPORTING', 'View fulfillment, picking, shipment, and delivery reports.'),
    ('REPORT.BILLING',   'View billing reports',       'REPORTING', 'View charges, invoices, credits, payments, and receivables reports.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_permission (code, name, module_code, description)
VALUES
    ('OUTBOUND.DO.READ',    'View delivery orders',    'OUTBOUND', 'View client delivery orders.'),
    ('OUTBOUND.DO.CREATE',  'Create delivery orders',  'OUTBOUND', 'Create and edit client delivery orders.'),
    ('OUTBOUND.VALIDATE',   'Validate delivery orders','OUTBOUND', 'Run and approve DO validation.'),
    ('OUTBOUND.ALLOCATE',   'Allocate inventory',      'OUTBOUND', 'Reserve inventory for outbound orders.'),
    ('OUTBOUND.WAVE',       'Manage outbound waves',   'OUTBOUND', 'Group orders and release pick work.'),
    ('OUTBOUND.PICK',       'Perform outbound picking','OUTBOUND', 'Execute and close pick tasks.'),
    ('OUTBOUND.CHECK',      'Check staged inventory',  'OUTBOUND', 'Verify staged stock before packing.'),
    ('OUTBOUND.PACK',       'Pack outbound inventory','OUTBOUND', 'Pack checked stock.'),
    ('OUTBOUND.SHIP',       'Dispatch shipments',      'OUTBOUND', 'Create and dispatch shipments.'),
    ('OUTBOUND.DELIVERY',   'Manage store delivery',   'OUTBOUND', 'Record delivery events, failures, returns, and POD.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_permission (code, name, module_code, description)
VALUES
    ('STOCK.READ',          'View stock',              'INVENTORY', 'View balances, transit, and movement history.'),
    ('STOCK.MOVE',          'Perform internal moves',  'INVENTORY', 'Create and execute internal movements.'),
    ('STOCK.TRANSFER',      'Manage warehouse transfers','INVENTORY','Create, dispatch, and receive warehouse transfers.'),
    ('STOCK.STATUS_CHANGE', 'Change stock status',     'INVENTORY', 'Hold, release, or otherwise change inventory status.'),
    ('STOCK.ADJUST',        'Adjust inventory',        'INVENTORY', 'Approve and post inventory adjustments.'),
    ('STOCK.COUNT',         'Perform stock counts',    'INVENTORY', 'Create, count, review, and post physical counts.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_permission (code, name, module_code, description)
VALUES
    ('BILLING.MASTER',    'Manage billing masters', 'BILLING', 'Manage accounts, cycles, currencies, terms, taxes, and services.'),
    ('BILLING.CONTRACT',  'Manage contracts',       'BILLING', 'Create and activate billing contracts.'),
    ('BILLING.RATE_CARD', 'Manage rate cards',      'BILLING', 'Create, approve, and activate service rates.'),
    ('BILLING.EVENT',     'Manage billable events', 'BILLING', 'Capture, exclude, and inspect billable events.'),
    ('BILLING.RUN',       'Run billing',             'BILLING', 'Calculate and review periodic charges.'),
    ('BILLING.INVOICE',   'Manage invoices',        'BILLING', 'Create, adjust, review, and issue invoices.'),
    ('BILLING.CREDIT',    'Issue credit notes',     'BILLING', 'Create and issue invoice credits.'),
    ('BILLING.PAYMENT',   'Manage payments',        'BILLING', 'Record and allocate client payments.'),
    ('BILLING.READ',      'View billing',           'BILLING', 'View billing history and balances.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_role (code, name, description)
VALUES ('SUPER_ADMIN', 'Super administrator', 'Full access to configured application permissions.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permission (role_id, permission_id)
SELECT role.role_id, permission.permission_id
FROM app_role role
CROSS JOIN app_permission permission
WHERE role.code = 'SUPER_ADMIN'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
VALUES
    (NULL, 'SECURITY', 'Security', NULL, 'shield', 800, false),
    (NULL, 'MASTER',   'Master Data', NULL, 'database', 100, false),
    (NULL, 'INBOUND',  'Inbound', NULL, 'truck', 200, false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
VALUES (NULL, 'STOCK_CONTROL', 'Stock Control', NULL, 'layers', 300, false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
VALUES (NULL, 'OUTBOUND_MENU', 'Outbound', NULL, 'send', 400, false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
VALUES (NULL, 'BILLING_MENU', 'Billing', NULL, 'receipt', 500, false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
VALUES (NULL, 'REPORT_MENU', 'Reports', NULL, 'bar-chart-2', 600, false)
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
SELECT parent.menu_id, child.code, child.label, child.route, child.icon_name,
       child.display_order, false
FROM app_menu parent
JOIN (VALUES
    ('REPORT_MENU', 'REPORT_MASTER',    'Master Reports',        '/reports/master',    'database', 10),
    ('REPORT_MENU', 'REPORT_INBOUND',   'Inbound Reports',       '/reports/inbound',   'download', 20),
    ('REPORT_MENU', 'REPORT_INVENTORY', 'Stock Control Reports', '/reports/inventory', 'layers',   30),
    ('REPORT_MENU', 'REPORT_OUTBOUND',  'Outbound Reports',      '/reports/outbound',  'upload',   40),
    ('REPORT_MENU', 'REPORT_BILLING',   'Billing Reports',       '/reports/billing',   'pie-chart',50)
) AS child(parent_code, code, label, route, icon_name, display_order)
  ON child.parent_code = parent.code
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
SELECT parent.menu_id, child.code, child.label, child.route, child.icon_name,
       child.display_order, false
FROM app_menu parent
JOIN (VALUES
    ('SECURITY', 'ACCOUNT',      'Accounts',      '/security/accounts', 'users',     10),
    ('SECURITY', 'ROLE',         'Roles',         '/security/roles',    'key',       20),
    ('MASTER',   'ORGANIZATION', 'Organizations', '/master/organizations', 'building', 10),
    ('MASTER',   'WAREHOUSE',    'Warehouses',    '/master/warehouses', 'warehouse', 20),
    ('MASTER',   'PARTNER',      'Partners',      '/master/partners',   'handshake', 30),
    ('MASTER',   'ITEM',         'Items',         '/master/items',      'box',       40),
    ('MASTER',   'CONFIG',       'Configuration', '/master/config',     'settings',  50),
    ('INBOUND',  'PURCHASE_ORDER','Purchase Orders','/inbound/purchase-orders', 'file-text', 10),
    ('INBOUND',  'RECEIVING',    'Receiving',      '/inbound/receiving', 'package',   20),
    ('INBOUND',  'QUALITY_CONTROL','Quality Control','/inbound/qc',      'check-circle', 30),
    ('INBOUND',  'QUARANTINE',   'Quarantine',     '/inbound/quarantine','alert-triangle', 40),
    ('INBOUND',  'PUTAWAY',      'Putaway',        '/inbound/putaway',   'archive',   50)
) AS child(parent_code, code, label, route, icon_name, display_order)
  ON child.parent_code = parent.code
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
SELECT parent.menu_id, child.code, child.label, child.route, child.icon_name,
       child.display_order, false
FROM app_menu parent
JOIN (VALUES
    ('BILLING_MENU', 'BILLING_ACCOUNT', 'Billing Accounts', '/billing/accounts',  'building',   10),
    ('BILLING_MENU', 'BILLING_CONTRACT','Contracts',        '/billing/contracts', 'file-text',  20),
    ('BILLING_MENU', 'RATE_CARD',       'Rate Cards',       '/billing/rate-cards','tag',        30),
    ('BILLING_MENU', 'BILLABLE_EVENT',  'Billable Events', '/billing/events',    'activity',   40),
    ('BILLING_MENU', 'BILLING_RUN',     'Billing Runs',     '/billing/runs',      'calculator', 50),
    ('BILLING_MENU', 'BILLING_INVOICE', 'Invoices',         '/billing/invoices',  'receipt',    60),
    ('BILLING_MENU', 'CREDIT_NOTE',     'Credit Notes',     '/billing/credits',   'minus-circle',70),
    ('BILLING_MENU', 'BILLING_PAYMENT', 'Payments',         '/billing/payments',  'credit-card',80)
) AS child(parent_code, code, label, route, icon_name, display_order)
  ON child.parent_code = parent.code
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
SELECT parent.menu_id, child.code, child.label, child.route, child.icon_name,
       child.display_order, false
FROM app_menu parent
JOIN (VALUES
    ('OUTBOUND_MENU', 'DELIVERY_ORDER',   'Client Delivery Orders','/outbound/orders',     'file-text', 10),
    ('OUTBOUND_MENU', 'ALLOCATION',       'Allocation',            '/outbound/allocation', 'lock',      20),
    ('OUTBOUND_MENU', 'OUTBOUND_WAVE',    'Outbound Waves',        '/outbound/waves',      'layers',    30),
    ('OUTBOUND_MENU', 'OUTBOUND_PICKING', 'Picking',               '/outbound/picking',    'check-square',40),
    ('OUTBOUND_MENU', 'OUTBOUND_STAGING', 'Staging & Checking',    '/outbound/staging',    'inbox',     50),
    ('OUTBOUND_MENU', 'OUTBOUND_PACKING', 'Packing',               '/outbound/packing',    'package',   60),
    ('OUTBOUND_MENU', 'OUTBOUND_SHIPMENT','Shipment',              '/outbound/shipments',  'truck',     70),
    ('OUTBOUND_MENU', 'STORE_DELIVERY',   'Store Delivery',        '/outbound/deliveries', 'map-pin',   80)
) AS child(parent_code, code, label, route, icon_name, display_order)
  ON child.parent_code = parent.code
ON CONFLICT (code) DO NOTHING;

INSERT INTO app_menu (
    parent_menu_id, code, label, route, icon_name, display_order,
    require_all_permissions
)
SELECT parent.menu_id, child.code, child.label, child.route, child.icon_name,
       child.display_order, false
FROM app_menu parent
JOIN (VALUES
    ('STOCK_CONTROL', 'STOCK_INQUIRY',       'Stock Inquiry',       '/stock/inquiry',       'search',       10),
    ('STOCK_CONTROL', 'INTERNAL_MOVEMENT',   'Internal Movement',   '/stock/movements',     'move',         20),
    ('STOCK_CONTROL', 'WAREHOUSE_TRANSFER',  'Warehouse Transfer',  '/stock/transfers',     'repeat',       30),
    ('STOCK_CONTROL', 'INVENTORY_STATUS',    'Status Change',       '/stock/status-change', 'shield',       40),
    ('STOCK_CONTROL', 'INVENTORY_ADJUSTMENT','Inventory Adjustment','/stock/adjustments',  'edit-3',       50),
    ('STOCK_CONTROL', 'STOCK_COUNT',         'Stock Count',         '/stock/counts',        'clipboard',    60)
) AS child(parent_code, code, label, route, icon_name, display_order)
  ON child.parent_code = parent.code
ON CONFLICT (code) DO NOTHING;

INSERT INTO menu_permission (menu_id, permission_id)
SELECT menu.menu_id, permission.permission_id
FROM (VALUES
    ('ACCOUNT',      'ACCOUNT.READ'),
    ('ROLE',         'ROLE.MANAGE'),
    ('ORGANIZATION', 'MASTER.ORGANIZATION'),
    ('WAREHOUSE',    'MASTER.WAREHOUSE'),
    ('PARTNER',      'MASTER.PARTNER'),
    ('ITEM',         'MASTER.ITEM'),
    ('CONFIG',       'MASTER.CONFIG'),
    ('PURCHASE_ORDER','INBOUND.PO.READ'),
    ('RECEIVING',    'INBOUND.RECEIVE'),
    ('QUALITY_CONTROL','INBOUND.QC'),
    ('QUARANTINE',   'INBOUND.DISPOSITION'),
    ('PUTAWAY',      'INBOUND.PUTAWAY')
) mapping(menu_code, permission_code)
JOIN app_menu menu ON menu.code = mapping.menu_code
JOIN app_permission permission ON permission.code = mapping.permission_code
ON CONFLICT (menu_id, permission_id) DO NOTHING;

INSERT INTO menu_permission (menu_id, permission_id)
SELECT menu.menu_id, permission.permission_id
FROM (VALUES
    ('REPORT_MENU',      'REPORT.MASTER'),
    ('REPORT_MENU',      'REPORT.INBOUND'),
    ('REPORT_MENU',      'REPORT.INVENTORY'),
    ('REPORT_MENU',      'REPORT.OUTBOUND'),
    ('REPORT_MENU',      'REPORT.BILLING'),
    ('REPORT_MASTER',    'REPORT.MASTER'),
    ('REPORT_INBOUND',   'REPORT.INBOUND'),
    ('REPORT_INVENTORY', 'REPORT.INVENTORY'),
    ('REPORT_OUTBOUND',  'REPORT.OUTBOUND'),
    ('REPORT_BILLING',   'REPORT.BILLING')
) mapping(menu_code, permission_code)
JOIN app_menu menu ON menu.code = mapping.menu_code
JOIN app_permission permission ON permission.code = mapping.permission_code
ON CONFLICT (menu_id, permission_id) DO NOTHING;

INSERT INTO menu_permission (menu_id, permission_id)
SELECT menu.menu_id, permission.permission_id
FROM (VALUES
    ('DELIVERY_ORDER',    'OUTBOUND.DO.READ'),
    ('DELIVERY_ORDER',    'OUTBOUND.DO.CREATE'),
    ('DELIVERY_ORDER',    'OUTBOUND.VALIDATE'),
    ('ALLOCATION',        'OUTBOUND.ALLOCATE'),
    ('OUTBOUND_WAVE',     'OUTBOUND.WAVE'),
    ('OUTBOUND_PICKING',  'OUTBOUND.PICK'),
    ('OUTBOUND_STAGING',  'OUTBOUND.CHECK'),
    ('OUTBOUND_PACKING',  'OUTBOUND.PACK'),
    ('OUTBOUND_SHIPMENT', 'OUTBOUND.SHIP'),
    ('STORE_DELIVERY',    'OUTBOUND.DELIVERY')
) mapping(menu_code, permission_code)
JOIN app_menu menu ON menu.code = mapping.menu_code
JOIN app_permission permission ON permission.code = mapping.permission_code
ON CONFLICT (menu_id, permission_id) DO NOTHING;

INSERT INTO menu_permission (menu_id, permission_id)
SELECT menu.menu_id, permission.permission_id
FROM (VALUES
    ('BILLING_ACCOUNT',  'BILLING.MASTER'),
    ('BILLING_CONTRACT', 'BILLING.CONTRACT'),
    ('RATE_CARD',        'BILLING.RATE_CARD'),
    ('BILLABLE_EVENT',   'BILLING.EVENT'),
    ('BILLING_RUN',      'BILLING.RUN'),
    ('BILLING_INVOICE',  'BILLING.INVOICE'),
    ('CREDIT_NOTE',      'BILLING.CREDIT'),
    ('BILLING_PAYMENT',  'BILLING.PAYMENT')
) mapping(menu_code, permission_code)
JOIN app_menu menu ON menu.code = mapping.menu_code
JOIN app_permission permission ON permission.code = mapping.permission_code
ON CONFLICT (menu_id, permission_id) DO NOTHING;

INSERT INTO menu_permission (menu_id, permission_id)
SELECT menu.menu_id, permission.permission_id
FROM (VALUES
    ('STOCK_INQUIRY',        'STOCK.READ'),
    ('INTERNAL_MOVEMENT',    'STOCK.MOVE'),
    ('WAREHOUSE_TRANSFER',   'STOCK.TRANSFER'),
    ('INVENTORY_STATUS',     'STOCK.STATUS_CHANGE'),
    ('INVENTORY_ADJUSTMENT', 'STOCK.ADJUST'),
    ('STOCK_COUNT',          'STOCK.COUNT')
) mapping(menu_code, permission_code)
JOIN app_menu menu ON menu.code = mapping.menu_code
JOIN app_permission permission ON permission.code = mapping.permission_code
ON CONFLICT (menu_id, permission_id) DO NOTHING;

COMMIT;
