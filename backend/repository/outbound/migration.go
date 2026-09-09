package outbound

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	mastermodel "wms-api/models/master"
	model "wms-api/models/outbound"
)

func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.OutboundValidationRule{}, &model.OutboundWaveType{}, &model.OutboundCheckResult{}, &model.OutboundCheckExceptionStatus{}, &model.OutboundCheckResolutionType{}, &model.DeliveryEventType{}, &model.DeliveryFailureReason{}, &model.OutboundOrder{}, &model.OutboundOrderLine{}, &model.OutboundValidationRun{}, &model.OutboundValidationResultDetail{}, &model.OutboundWave{}, &model.OutboundWaveOrder{}, &model.InventoryReservation{}, &model.PickTask{}, &model.PickExecution{}, &model.OutboundStaging{}, &model.OutboundStagingLine{}, &model.OutboundCheck{}, &model.OutboundCheckLine{}, &model.OutboundCheckException{}, &model.OutboundCheckResolution{}, &model.OutboundCheckResolutionMovement{}, &model.OutboundCheckResolutionReservation{}, &model.Packing{}, &model.PackingLine{}, &model.Carrier{}, &model.CarrierService{}, &model.CarrierDriver{}, &model.Shipment{}, &model.ShipmentDriver{}, &model.ShipmentOrder{}, &model.ShipmentPacking{}, &model.ShipmentLine{}, &model.Delivery{}, &model.DeliveryLine{}, &model.DeliveryEvent{}, &model.DeliveryEventLine{}, &model.OutboundReturnPolicy{}, &model.DeliveryReturnLine{}); err != nil {
			return err
		}
		statements := []string{
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_outbound_owner_do ON outbound_order(owner_id,client_delivery_order_no)",
			"CREATE INDEX IF NOT EXISTS ix_outbound_owner_warehouse_date ON outbound_order(owner_id,warehouse_id,business_date DESC)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_outbound_line_no ON outbound_order_line(outbound_id,line_no)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_validation_run_rule ON outbound_validation_result_detail(validation_run_id,outbound_validation_rule_id)",
			"CREATE INDEX IF NOT EXISTS ix_reservation_outbound_line ON inventory_reservation(outbound_line_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_pick_reservation ON pick_task(reservation_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_pick_execution_movement ON pick_execution(movement_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_staging_pick_execution ON outbound_staging_line(pick_execution_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_staging_order_wave ON outbound_staging(outbound_id,wave_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_check_line_no ON outbound_check_line(outbound_check_id,line_no)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_check_exception_line ON outbound_check_exception(outbound_check_line_id)",
			"DROP INDEX IF EXISTS uq_packing_outbound",
			"CREATE INDEX IF NOT EXISTS ix_packing_outbound ON packing(outbound_id)",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_active_shipment_order ON shipment_order(outbound_id) WHERE removed_at IS NULL",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_shipment_primary_driver ON shipment_driver(shipment_id) WHERE is_primary",
			"CREATE UNIQUE INDEX IF NOT EXISTS uq_delivery_shipment_outbound ON delivery(shipment_id,outbound_id)",
			"CREATE INDEX IF NOT EXISTS ix_delivery_owner_date ON delivery(outbound_id,business_date DESC)",
			"DO $$ BEGIN ALTER TABLE outbound_order ADD CONSTRAINT ck_outbound_version CHECK(version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_wave ADD CONSTRAINT ck_outbound_wave_version CHECK(version_no>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order_line ADD CONSTRAINT ck_outbound_line_qty CHECK(line_no>0 AND ordered_qty>0 AND allocated_qty>=0 AND picked_qty>=0 AND picked_qty<=allocated_qty AND allocated_qty<=ordered_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inventory_reservation ADD CONSTRAINT ck_reservation_qty CHECK(reserved_qty>0 AND picked_qty>=0 AND picked_qty<=reserved_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT ck_pick_qty CHECK(planned_qty>0 AND picked_qty>=0 AND short_qty>=0 AND picked_qty+short_qty<=planned_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_staging_line ADD CONSTRAINT ck_staging_line_qty CHECK(staged_qty>0 AND removed_qty>=0 AND removed_qty<=staged_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_line ADD CONSTRAINT ck_outbound_check_line_qty CHECK(line_no>0 AND expected_qty>0 AND (checked_qty IS NULL OR checked_qty>=0) AND exception_qty>=0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_exception ADD CONSTRAINT ck_outbound_check_exception_qty CHECK(exception_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_resolution ADD CONSTRAINT ck_outbound_check_resolution_qty CHECK(resolved_qty>0 AND ((approved_at IS NULL)=(approved_by IS NULL))); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing_line ADD CONSTRAINT ck_packing_line_qty CHECK(packed_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_line ADD CONSTRAINT ck_shipment_line_qty CHECK(shipped_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_line ADD CONSTRAINT ck_delivery_line_qty CHECK(planned_qty>0 AND delivered_qty>=0 AND returned_qty>=0 AND delivered_qty+returned_qty<=planned_qty); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_event_line ADD CONSTRAINT ck_delivery_event_line_qty CHECK(delivered_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_return_line ADD CONSTRAINT ck_delivery_return_line_qty CHECK(returned_qty>0); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_validation_rule ADD CONSTRAINT fk_outbound_validation_rule_severity FOREIGN KEY(validation_severity_id) REFERENCES validation_severity(validation_severity_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order ADD CONSTRAINT fk_outbound_order_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order ADD CONSTRAINT fk_outbound_order_owner FOREIGN KEY(owner_id) REFERENCES organization(organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order ADD CONSTRAINT fk_outbound_order_customer FOREIGN KEY(owner_id,customer_id) REFERENCES business_partner(owner_id,partner_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order ADD CONSTRAINT fk_outbound_order_ship_to FOREIGN KEY(owner_id,ship_to_partner_id) REFERENCES business_partner(owner_id,partner_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order ADD CONSTRAINT fk_outbound_order_warehouse_owner FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order_line ADD CONSTRAINT fk_outbound_line_header FOREIGN KEY(outbound_id) REFERENCES outbound_order(outbound_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order_line ADD CONSTRAINT fk_outbound_line_item FOREIGN KEY(item_id) REFERENCES item(item_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_order_line ADD CONSTRAINT fk_outbound_line_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_validation_run ADD CONSTRAINT fk_outbound_validation_header FOREIGN KEY(outbound_id) REFERENCES outbound_order(outbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_validation_run ADD CONSTRAINT fk_outbound_validation_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_validation_result_detail ADD CONSTRAINT fk_outbound_validation_result_run FOREIGN KEY(validation_run_id) REFERENCES outbound_validation_run(validation_run_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_validation_result_detail ADD CONSTRAINT fk_outbound_validation_result_rule FOREIGN KEY(outbound_validation_rule_id) REFERENCES outbound_validation_rule(outbound_validation_rule_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_wave ADD CONSTRAINT fk_outbound_wave_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_wave ADD CONSTRAINT fk_outbound_wave_type FOREIGN KEY(outbound_wave_type_id) REFERENCES outbound_wave_type(outbound_wave_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_wave ADD CONSTRAINT fk_outbound_wave_warehouse_owner FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_wave ADD CONSTRAINT fk_outbound_wave_strategy FOREIGN KEY(picking_strategy_id) REFERENCES picking_strategy(picking_strategy_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_wave_order ADD CONSTRAINT fk_outbound_wave_order_wave FOREIGN KEY(wave_id) REFERENCES outbound_wave(wave_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_wave_order ADD CONSTRAINT fk_outbound_wave_order_header FOREIGN KEY(outbound_id) REFERENCES outbound_order(outbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inventory_reservation ADD CONSTRAINT fk_reservation_outbound_line FOREIGN KEY(outbound_line_id) REFERENCES outbound_order_line(outbound_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inventory_reservation ADD CONSTRAINT fk_reservation_balance FOREIGN KEY(balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inventory_reservation ADD CONSTRAINT fk_reservation_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inventory_reservation ADD CONSTRAINT fk_reservation_strategy FOREIGN KEY(picking_strategy_id) REFERENCES picking_strategy(picking_strategy_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE inventory_reservation ADD CONSTRAINT fk_reservation_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_wave FOREIGN KEY(wave_id) REFERENCES outbound_wave(wave_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_reservation FOREIGN KEY(reservation_id) REFERENCES inventory_reservation(reservation_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_line FOREIGN KEY(outbound_line_id) REFERENCES outbound_order_line(outbound_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_task_type FOREIGN KEY(task_type_id) REFERENCES task_type(task_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_task_status FOREIGN KEY(task_status_id) REFERENCES task_status(task_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_task_priority FOREIGN KEY(task_priority_id) REFERENCES task_priority(task_priority_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_source_location FOREIGN KEY(source_location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_target_location FOREIGN KEY(target_location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_task ADD CONSTRAINT fk_pick_short_reason FOREIGN KEY(short_reason_code_id) REFERENCES reason_code(reason_code_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_execution ADD CONSTRAINT fk_pick_execution_task FOREIGN KEY(pick_task_id) REFERENCES pick_task(pick_task_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_execution ADD CONSTRAINT fk_pick_execution_source_balance FOREIGN KEY(source_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_execution ADD CONSTRAINT fk_pick_execution_staging_balance FOREIGN KEY(staging_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE pick_execution ADD CONSTRAINT fk_pick_execution_movement FOREIGN KEY(movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_staging ADD CONSTRAINT fk_staging_outbound FOREIGN KEY(outbound_id) REFERENCES outbound_order(outbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_staging ADD CONSTRAINT fk_staging_wave FOREIGN KEY(wave_id) REFERENCES outbound_wave(wave_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_staging ADD CONSTRAINT fk_staging_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_staging ADD CONSTRAINT fk_staging_location FOREIGN KEY(staging_location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_staging_line ADD CONSTRAINT fk_staging_line_header FOREIGN KEY(staging_id) REFERENCES outbound_staging(staging_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_staging_line ADD CONSTRAINT fk_staging_line_execution FOREIGN KEY(pick_execution_id) REFERENCES pick_execution(pick_execution_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_staging_line ADD CONSTRAINT fk_staging_line_balance FOREIGN KEY(staging_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check ADD CONSTRAINT fk_check_parent FOREIGN KEY(parent_check_id) REFERENCES outbound_check(outbound_check_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check ADD CONSTRAINT fk_check_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check ADD CONSTRAINT fk_check_staging FOREIGN KEY(staging_id) REFERENCES outbound_staging(staging_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_line ADD CONSTRAINT fk_check_line_header FOREIGN KEY(outbound_check_id) REFERENCES outbound_check(outbound_check_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_line ADD CONSTRAINT fk_check_line_staging FOREIGN KEY(staging_line_id) REFERENCES outbound_staging_line(staging_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_line ADD CONSTRAINT fk_check_line_result FOREIGN KEY(outbound_check_result_id) REFERENCES outbound_check_result(outbound_check_result_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_line ADD CONSTRAINT fk_check_line_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_exception ADD CONSTRAINT fk_check_exception_line FOREIGN KEY(outbound_check_line_id) REFERENCES outbound_check_line(outbound_check_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_exception ADD CONSTRAINT fk_check_exception_status FOREIGN KEY(status_id) REFERENCES outbound_check_exception_status(outbound_check_exception_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_resolution ADD CONSTRAINT fk_check_resolution_exception FOREIGN KEY(outbound_check_exception_id) REFERENCES outbound_check_exception(outbound_check_exception_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_resolution ADD CONSTRAINT fk_check_resolution_type FOREIGN KEY(outbound_check_resolution_type_id) REFERENCES outbound_check_resolution_type(outbound_check_resolution_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_resolution_movement ADD CONSTRAINT fk_check_resolution_movement_resolution FOREIGN KEY(outbound_check_resolution_id) REFERENCES outbound_check_resolution(outbound_check_resolution_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_resolution_movement ADD CONSTRAINT fk_check_resolution_movement_stock FOREIGN KEY(movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_resolution_reservation ADD CONSTRAINT fk_check_resolution_reservation_resolution FOREIGN KEY(outbound_check_resolution_id) REFERENCES outbound_check_resolution(outbound_check_resolution_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_check_resolution_reservation ADD CONSTRAINT fk_check_resolution_reservation_stock FOREIGN KEY(reservation_id) REFERENCES inventory_reservation(reservation_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing ADD CONSTRAINT fk_packing_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing ADD CONSTRAINT fk_packing_outbound FOREIGN KEY(outbound_id) REFERENCES outbound_order(outbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing ADD CONSTRAINT fk_packing_location FOREIGN KEY(packing_location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing_line ADD CONSTRAINT fk_packing_line_header FOREIGN KEY(packing_id) REFERENCES packing(packing_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing_line ADD CONSTRAINT fk_packing_line_check FOREIGN KEY(outbound_check_line_id) REFERENCES outbound_check_line(outbound_check_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing_line ADD CONSTRAINT fk_packing_line_task FOREIGN KEY(pick_task_id) REFERENCES pick_task(pick_task_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing_line ADD CONSTRAINT fk_packing_line_source_balance FOREIGN KEY(source_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing_line ADD CONSTRAINT fk_packing_line_target_balance FOREIGN KEY(packing_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing_line ADD CONSTRAINT fk_packing_line_movement FOREIGN KEY(movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE packing_line ADD CONSTRAINT fk_packing_line_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment ADD CONSTRAINT fk_shipment_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment ADD CONSTRAINT fk_shipment_warehouse_owner FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE carrier ADD CONSTRAINT fk_carrier_partner FOREIGN KEY(business_partner_id) REFERENCES business_partner(partner_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE carrier_service ADD CONSTRAINT fk_carrier_service_carrier FOREIGN KEY(carrier_id) REFERENCES carrier(carrier_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE carrier_driver ADD CONSTRAINT fk_carrier_driver_carrier FOREIGN KEY(carrier_id) REFERENCES carrier(carrier_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE carrier_driver ADD CONSTRAINT fk_carrier_driver_account FOREIGN KEY(account_id) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE carrier_driver ADD CONSTRAINT fk_carrier_driver_created_by FOREIGN KEY(created_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment ADD CONSTRAINT fk_shipment_carrier_service FOREIGN KEY(carrier_service_id) REFERENCES carrier_service(carrier_service_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_driver ADD CONSTRAINT fk_shipment_driver_shipment FOREIGN KEY(shipment_id) REFERENCES shipment(shipment_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_driver ADD CONSTRAINT fk_shipment_driver_driver FOREIGN KEY(driver_id) REFERENCES carrier_driver(driver_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_driver ADD CONSTRAINT fk_shipment_driver_assigned_by FOREIGN KEY(assigned_by) REFERENCES app_account(account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_order ADD CONSTRAINT fk_shipment_order_header FOREIGN KEY(shipment_id) REFERENCES shipment(shipment_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_order ADD CONSTRAINT fk_shipment_order_outbound FOREIGN KEY(outbound_id) REFERENCES outbound_order(outbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_packing ADD CONSTRAINT fk_shipment_packing_header FOREIGN KEY(shipment_id) REFERENCES shipment(shipment_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_packing ADD CONSTRAINT fk_shipment_packing_pack FOREIGN KEY(packing_id) REFERENCES packing(packing_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_line ADD CONSTRAINT fk_shipment_line_header FOREIGN KEY(shipment_id) REFERENCES shipment(shipment_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_line ADD CONSTRAINT fk_shipment_line_packing FOREIGN KEY(packing_line_id) REFERENCES packing_line(packing_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_line ADD CONSTRAINT fk_shipment_line_source_balance FOREIGN KEY(source_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_line ADD CONSTRAINT fk_shipment_line_movement FOREIGN KEY(movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE shipment_line ADD CONSTRAINT fk_shipment_line_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery ADD CONSTRAINT fk_delivery_status FOREIGN KEY(document_type_id,status_id) REFERENCES document_status(document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery ADD CONSTRAINT fk_delivery_shipment FOREIGN KEY(shipment_id) REFERENCES shipment(shipment_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery ADD CONSTRAINT fk_delivery_outbound FOREIGN KEY(outbound_id) REFERENCES outbound_order(outbound_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_line ADD CONSTRAINT fk_delivery_line_header FOREIGN KEY(delivery_id) REFERENCES delivery(delivery_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_line ADD CONSTRAINT fk_delivery_line_shipment FOREIGN KEY(shipment_line_id) REFERENCES shipment_line(shipment_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_line ADD CONSTRAINT fk_delivery_line_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_event ADD CONSTRAINT fk_delivery_event_header FOREIGN KEY(delivery_id) REFERENCES delivery(delivery_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_event ADD CONSTRAINT fk_delivery_event_type FOREIGN KEY(delivery_event_type_id) REFERENCES delivery_event_type(delivery_event_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_event ADD CONSTRAINT fk_delivery_event_failure_reason FOREIGN KEY(delivery_failure_reason_id) REFERENCES delivery_failure_reason(delivery_failure_reason_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_event_line ADD CONSTRAINT fk_delivery_event_line_event FOREIGN KEY(delivery_event_id) REFERENCES delivery_event(delivery_event_id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_event_line ADD CONSTRAINT fk_delivery_event_line_delivery FOREIGN KEY(delivery_line_id) REFERENCES delivery_line(delivery_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_return_policy ADD CONSTRAINT fk_return_policy_warehouse_owner FOREIGN KEY(owner_id,warehouse_id) REFERENCES warehouse_owner(owner_id,warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_return_policy ADD CONSTRAINT fk_return_policy_location FOREIGN KEY(return_location_id) REFERENCES warehouse_location(location_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE outbound_return_policy ADD CONSTRAINT fk_return_policy_status FOREIGN KEY(return_inventory_status_id) REFERENCES inventory_status(inventory_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_return_line ADD CONSTRAINT fk_delivery_return_delivery FOREIGN KEY(delivery_id) REFERENCES delivery(delivery_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_return_line ADD CONSTRAINT fk_delivery_return_line FOREIGN KEY(delivery_line_id) REFERENCES delivery_line(delivery_line_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_return_line ADD CONSTRAINT fk_delivery_return_balance FOREIGN KEY(returned_balance_id) REFERENCES inventory_balance(balance_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_return_line ADD CONSTRAINT fk_delivery_return_movement FOREIGN KEY(movement_id) REFERENCES inventory_movement(movement_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
			"DO $$ BEGIN ALTER TABLE delivery_return_line ADD CONSTRAINT fk_delivery_return_uom FOREIGN KEY(uom_id) REFERENCES uom(uom_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$",
		}
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func SeedReferenceData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		ctx := context.Background()
		repositories := NewRepositories(tx)
		if err := repositories.Master.ValidationSeverity.Seed(ctx, []mastermodel.ValidationSeverity{{Code: "ERROR", Name: "Error", BlocksProcessing: true, IsActive: true}, {Code: "WARNING", Name: "Warning", BlocksProcessing: false, IsActive: true}}); err != nil {
			return err
		}
		if err := repositories.Master.ReasonCode.Seed(ctx, []mastermodel.ReasonCode{{ModuleCode: "OUTBOUND", Code: "SHORT_PICK", Name: "Short pick", Description: stringPointer("Picked quantity was below requirement."), RequiresNote: true, IsActive: true}, {ModuleCode: "OUTBOUND", Code: "ALLOCATION_SHORT", Name: "Allocation shortage", Description: stringPointer("Insufficient allocatable inventory."), RequiresNote: true, IsActive: true}}); err != nil {
			return err
		}
		for _, wave := range []model.OutboundWaveType{{Code: "SINGLE_ORDER", Name: "Single order", IsActive: true}, {Code: "MANUAL", Name: "Manual", IsActive: true}, {Code: "ROUTE", Name: "Route", IsActive: true}, {Code: "CARRIER", Name: "Carrier", IsActive: true}, {Code: "CUT_OFF", Name: "Cut-off time", IsActive: true}} {
			value := wave
			if err := repositories.WaveType.Seed(ctx, &value); err != nil {
				return err
			}
		}
		for _, value := range []model.OutboundCheckResult{
			{Code: "PASS", Name: "Pass", IsPass: true, IsActive: true},
			{Code: "SHORT", Name: "Short", RequiresNote: true, IsActive: true},
			{Code: "OVER", Name: "Over", RequiresNote: true, IsActive: true},
			{Code: "WRONG_ITEM", Name: "Wrong item", RequiresNote: true, IsActive: true},
			{Code: "DAMAGED", Name: "Damaged", RequiresNote: true, IsActive: true},
		} {
			item := value
			if err := repositories.CheckResult.Seed(ctx, &item); err != nil {
				return err
			}
		}
		for _, value := range []model.OutboundCheckExceptionStatus{{Code: "OPEN", Name: "Open", IsActive: true}, {Code: "IN_PROGRESS", Name: "In progress", IsActive: true}, {Code: "RESOLVED", Name: "Resolved", IsFinal: true, IsActive: true}} {
			item := value
			if err := repositories.ExceptionStatus.Seed(ctx, &item); err != nil {
				return err
			}
		}
		for _, value := range []model.OutboundCheckResolutionType{
			{Code: "STOCK_CORRECTION", Name: "Stock correction", CountsAsStockCorrection: true, IsActive: true},
			{Code: "REPLACEMENT", Name: "Replacement stock", CountsAsReplacement: true, IsActive: true},
			{Code: "ACCEPT_SHORT", Name: "Accept short quantity", CountsAsShortAcceptance: true, RequiresApproval: true, IsActive: true},
		} {
			item := value
			if err := repositories.ResolutionType.Seed(ctx, &item); err != nil {
				return err
			}
		}
		for _, value := range []model.DeliveryEventType{
			{Code: "DEPARTED", Name: "Departed warehouse", IsActive: true},
			{Code: "ARRIVED_STORE", Name: "Arrived at store", IsActive: true},
			{Code: "DELIVERED", Name: "Delivered", MarksDelivered: true, IsActive: true},
			{Code: "DELIVERY_FAILED", Name: "Delivery failed", MarksFailed: true, IsActive: true},
			{Code: "RETURNED_TO_DEPOT", Name: "Returned to depot", MarksReturned: true, IsActive: true},
		} {
			item := value
			if err := repositories.DeliveryEventType.Seed(ctx, &item); err != nil {
				return err
			}
		}
		for _, value := range []model.DeliveryFailureReason{
			{Code: "STORE_CLOSED", Name: "Store closed", IsActive: true}, {Code: "RECIPIENT_REJECTED", Name: "Recipient rejected", IsActive: true},
			{Code: "ADDRESS_NOT_FOUND", Name: "Address not found", IsActive: true}, {Code: "DAMAGED_IN_TRANSIT", Name: "Damaged in transit", IsActive: true},
			{Code: "VEHICLE_ISSUE", Name: "Vehicle issue", IsActive: true}, {Code: "OTHER", Name: "Other", IsActive: true},
		} {
			item := value
			if err := repositories.DeliveryFailureReason.Seed(ctx, &item); err != nil {
				return err
			}
		}
		if err := repositories.Master.ReasonCode.Seed(ctx, []mastermodel.ReasonCode{{ModuleCode: "OUTBOUND", Code: "DELIVERY_RETURN", Name: "Delivery return", Description: stringPointer("Undelivered goods returned to controlled warehouse stock."), RequiresNote: true, IsActive: true}}); err != nil {
			return err
		}
		errorSeverity, err := repositories.Master.ValidationSeverity.ByCode(ctx, "ERROR")
		if err != nil {
			return fmt.Errorf("outbound validation ERROR severity is not configured: %w", err)
		}
		warningSeverity, err := repositories.Master.ValidationSeverity.ByCode(ctx, "WARNING")
		if err != nil {
			return fmt.Errorf("outbound validation WARNING severity is not configured: %w", err)
		}
		rules := []model.OutboundValidationRule{
			{Code: "DO_NUMBER", Name: "Delivery order number", Description: stringPointer("Client delivery order number is required."), HandlerCode: "REQUIRE_DO_NUMBER", ValidationSeverityID: errorSeverity.ID, DisplayOrder: 10, IsActive: true},
			{Code: "SHIP_TO", Name: "Ship-to destination", Description: stringPointer("An active ship-to address is required."), HandlerCode: "REQUIRE_SHIP_TO", ValidationSeverityID: errorSeverity.ID, DisplayOrder: 20, IsActive: true},
			{Code: "ORDER_LINES", Name: "Order lines", Description: stringPointer("At least one order line is required."), HandlerCode: "REQUIRE_ORDER_LINES", ValidationSeverityID: errorSeverity.ID, DisplayOrder: 30, IsActive: true},
			{Code: "BASE_UOM", Name: "Base UOM", Description: stringPointer("Every line must use the item base UOM."), HandlerCode: "REQUIRE_BASE_UOM", ValidationSeverityID: errorSeverity.ID, DisplayOrder: 40, IsActive: true},
			{Code: "REQUESTED_SHIP", Name: "Requested ship time", Description: stringPointer("Requested ship time is recommended."), HandlerCode: "REQUIRE_REQUESTED_SHIP", ValidationSeverityID: warningSeverity.ID, DisplayOrder: 50, IsActive: true},
		}
		for _, rule := range rules {
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "description", "handler_code", "validation_severity_id", "display_order", "is_active"})}).Create(&rule).Error; err != nil {
				return Error(err)
			}
		}
		return nil
	})
}

func stringPointer(v string) *string { return &v }
