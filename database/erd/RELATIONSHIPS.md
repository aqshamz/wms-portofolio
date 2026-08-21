# WMS relationship catalog

| Child table/column | Parent table/column | Required | Delete action |
|---|---|:---:|---|
| `account_owner_access(account_id)` | `app_account(account_id)` | Yes | CASCADE |
| `account_owner_access(granted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `account_owner_access(owner_id)` | `organization(organization_id)` | Yes | CASCADE |
| `account_role(account_id)` | `app_account(account_id)` | Yes | CASCADE |
| `account_role(assigned_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `account_role(role_id)` | `app_role(role_id)` | Yes | CASCADE |
| `account_warehouse_access(account_id)` | `app_account(account_id)` | Yes | CASCADE |
| `account_warehouse_access(granted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `account_warehouse_access(warehouse_id)` | `warehouse(warehouse_id)` | Yes | CASCADE |
| `app_account(account_status_id)` | `account_status(account_status_id)` | Yes | RESTRICT/NO ACTION |
| `app_account(authentication_policy_id)` | `authentication_policy(authentication_policy_id)` | No | RESTRICT/NO ACTION |
| `app_account(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `app_account(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `app_menu(parent_menu_id)` | `app_menu(menu_id)` | No | RESTRICT/NO ACTION |
| `app_permission(module_code)` | `app_module(code)` | Yes | RESTRICT/NO ACTION |
| `app_role(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `app_session(account_id)` | `app_account(account_id)` | Yes | CASCADE |
| `app_session(session_revocation_reason_id)` | `session_revocation_reason(session_revocation_reason_id)` | No | RESTRICT/NO ACTION |
| `billable_event(billing_account_id)` | `billing_account(billing_account_id)` | Yes | RESTRICT/NO ACTION |
| `billable_event(billing_contract_id)` | `billing_contract(billing_contract_id)` | Yes | RESTRICT/NO ACTION |
| `billable_event(billing_service_id)` | `billing_service(billing_service_id)` | Yes | RESTRICT/NO ACTION |
| `billable_event(category_id)` | `item_category(category_id)` | No | RESTRICT/NO ACTION |
| `billable_event(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billable_event(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `billable_event(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `billable_event(excluded_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `billable_event(exclusion_reason_code_id)` | `reason_code(reason_code_id)` | No | RESTRICT/NO ACTION |
| `billable_event(handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `billable_event(item_id)` | `item(item_id)` | No | RESTRICT/NO ACTION |
| `billable_event(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `billable_event(rated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `billable_event(source_document_type_id)` | `document_type(document_type_id)` | No | RESTRICT/NO ACTION |
| `billable_event(uom_id)` | `uom(uom_id)` | No | RESTRICT/NO ACTION |
| `billable_event(warehouse_id)` | `warehouse(warehouse_id)` | No | RESTRICT/NO ACTION |
| `billing_account(bill_to_partner_id)` | `business_partner(partner_id)` | No | RESTRICT/NO ACTION |
| `billing_account(billing_cycle_id)` | `billing_cycle(billing_cycle_id)` | Yes | RESTRICT/NO ACTION |
| `billing_account(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `billing_account(currency_id)` | `currency(currency_id)` | Yes | RESTRICT/NO ACTION |
| `billing_account(default_tax_rule_id)` | `tax_rule(tax_rule_id)` | No | RESTRICT/NO ACTION |
| `billing_account(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `billing_account(owner_id, bill_to_partner_id)` | `business_partner(owner_id, partner_id)` | No | RESTRICT/NO ACTION |
| `billing_account(payment_term_id)` | `payment_term(payment_term_id)` | Yes | RESTRICT/NO ACTION |
| `billing_account(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `billing_charge(billable_event_id)` | `billable_event(billable_event_id)` | Yes | RESTRICT/NO ACTION |
| `billing_charge(billing_run_id)` | `billing_run(billing_run_id)` | Yes | CASCADE |
| `billing_charge(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_charge(currency_id)` | `currency(currency_id)` | Yes | RESTRICT/NO ACTION |
| `billing_charge(rate_card_line_id)` | `rate_card_line(rate_card_line_id)` | Yes | RESTRICT/NO ACTION |
| `billing_charge(tax_rule_id)` | `tax_rule(tax_rule_id)` | No | RESTRICT/NO ACTION |
| `billing_charge(uom_id)` | `uom(uom_id)` | No | RESTRICT/NO ACTION |
| `billing_contract(billing_account_id)` | `billing_account(billing_account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_contract(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_contract(currency_id)` | `currency(currency_id)` | Yes | RESTRICT/NO ACTION |
| `billing_contract(default_tax_rule_id)` | `tax_rule(tax_rule_id)` | No | RESTRICT/NO ACTION |
| `billing_contract(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `billing_contract(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `billing_contract(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `billing_contract(owner_id, billing_account_id)` | `billing_account(owner_id, billing_account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_contract(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | No | RESTRICT/NO ACTION |
| `billing_contract(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_contract(warehouse_id)` | `warehouse(warehouse_id)` | No | RESTRICT/NO ACTION |
| `billing_credit_note_line(billing_invoice_line_id)` | `billing_invoice_line(billing_invoice_line_id)` | No | RESTRICT/NO ACTION |
| `billing_credit_note_line(credit_note_id)` | `billing_credit_note(credit_note_id)` | Yes | CASCADE |
| `billing_credit_note(billing_invoice_id)` | `billing_invoice(billing_invoice_id)` | Yes | RESTRICT/NO ACTION |
| `billing_credit_note(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_credit_note(currency_id)` | `currency(currency_id)` | Yes | RESTRICT/NO ACTION |
| `billing_credit_note(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `billing_credit_note(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `billing_credit_note(issued_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `billing_credit_note(reason_code_id)` | `reason_code(reason_code_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice_adjustment(billing_adjustment_type_id)` | `billing_adjustment_type(billing_adjustment_type_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice_adjustment(billing_invoice_id)` | `billing_invoice(billing_invoice_id)` | Yes | CASCADE |
| `billing_invoice_adjustment(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice_adjustment(reason_code_id)` | `reason_code(reason_code_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice_adjustment(tax_rule_id)` | `tax_rule(tax_rule_id)` | No | RESTRICT/NO ACTION |
| `billing_invoice_line(billing_charge_id)` | `billing_charge(billing_charge_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice_line(billing_invoice_id)` | `billing_invoice(billing_invoice_id)` | Yes | CASCADE |
| `billing_invoice_line(billing_service_id)` | `billing_service(billing_service_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice_line(tax_rule_id)` | `tax_rule(tax_rule_id)` | No | RESTRICT/NO ACTION |
| `billing_invoice_line(uom_id)` | `uom(uom_id)` | No | RESTRICT/NO ACTION |
| `billing_invoice(billing_account_id)` | `billing_account(billing_account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice(billing_contract_id)` | `billing_contract(billing_contract_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice(billing_run_id)` | `billing_run(billing_run_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice(currency_id)` | `currency(currency_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `billing_invoice(issued_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `billing_invoice(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_payment_allocation(allocated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_payment_allocation(billing_invoice_id)` | `billing_invoice(billing_invoice_id)` | Yes | RESTRICT/NO ACTION |
| `billing_payment_allocation(payment_id)` | `billing_payment(payment_id)` | Yes | CASCADE |
| `billing_payment(billing_account_id)` | `billing_account(billing_account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_payment(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_payment(currency_id)` | `currency(currency_id)` | Yes | RESTRICT/NO ACTION |
| `billing_payment(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `billing_payment(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `billing_payment(payment_method_id)` | `payment_method(payment_method_id)` | Yes | RESTRICT/NO ACTION |
| `billing_run(billing_account_id)` | `billing_account(billing_account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_run(billing_contract_id)` | `billing_contract(billing_contract_id)` | Yes | RESTRICT/NO ACTION |
| `billing_run(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `billing_run(currency_id)` | `currency(currency_id)` | Yes | RESTRICT/NO ACTION |
| `billing_run(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `billing_run(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `billing_run(reviewed_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `billing_service_movement_type(billing_service_id)` | `billing_service(billing_service_id)` | Yes | RESTRICT/NO ACTION |
| `billing_service_movement_type(movement_type_id)` | `movement_type(movement_type_id)` | Yes | RESTRICT/NO ACTION |
| `billing_service(default_charge_basis_id)` | `charge_basis(charge_basis_id)` | Yes | RESTRICT/NO ACTION |
| `billing_service(default_uom_id)` | `uom(uom_id)` | No | RESTRICT/NO ACTION |
| `billing_service(module_code)` | `app_module(code)` | Yes | RESTRICT/NO ACTION |
| `business_partner_type(partner_id)` | `business_partner(partner_id)` | Yes | CASCADE |
| `business_partner_type(partner_type_id)` | `partner_type(partner_type_id)` | Yes | RESTRICT/NO ACTION |
| `business_partner(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `business_partner(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `business_partner(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `carrier_driver(account_id)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `carrier_driver(carrier_id)` | `carrier(carrier_id)` | Yes | RESTRICT/NO ACTION |
| `carrier_driver(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `carrier_service(carrier_id)` | `carrier(carrier_id)` | Yes | RESTRICT/NO ACTION |
| `carrier(business_partner_id)` | `business_partner(partner_id)` | No | RESTRICT/NO ACTION |
| `delivery_event_line(delivery_event_id)` | `delivery_event(delivery_event_id)` | Yes | CASCADE |
| `delivery_event_line(delivery_line_id)` | `delivery_line(delivery_line_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_event(delivery_event_type_id)` | `delivery_event_type(delivery_event_type_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_event(delivery_failure_reason_id)` | `delivery_failure_reason(delivery_failure_reason_id)` | No | RESTRICT/NO ACTION |
| `delivery_event(delivery_id)` | `delivery(delivery_id)` | Yes | CASCADE |
| `delivery_event(recorded_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_line(delivery_id)` | `delivery(delivery_id)` | Yes | CASCADE |
| `delivery_line(shipment_line_id)` | `shipment_line(shipment_line_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_return_line(delivery_id)` | `delivery(delivery_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_return_line(delivery_line_id)` | `delivery_line(delivery_line_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_return_line(movement_id)` | `inventory_movement(movement_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_return_line(return_location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_return_line(returned_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_return_line(returned_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `delivery_return_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `delivery(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `delivery(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `delivery(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `delivery(outbound_id)` | `outbound_order(outbound_id)` | Yes | RESTRICT/NO ACTION |
| `delivery(shipment_id)` | `shipment(shipment_id)` | Yes | RESTRICT/NO ACTION |
| `delivery(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `document_daily_counter(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `document_number_rule(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `document_number_rule(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `document_status_transition(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `document_status_transition(document_type_id, from_status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `document_status_transition(document_type_id, to_status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `document_status_transition(required_permission_id)` | `app_permission(permission_id)` | No | RESTRICT/NO ACTION |
| `document_status(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `document_type(module_code)` | `app_module(code)` | Yes | RESTRICT/NO ACTION |
| `handling_unit(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `handling_unit(current_location_id)` | `warehouse_location(location_id)` | No | RESTRICT/NO ACTION |
| `handling_unit(handling_unit_type_id)` | `handling_unit_type(handling_unit_type_id)` | Yes | RESTRICT/NO ACTION |
| `handling_unit(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `handling_unit(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `handling_unit(parent_handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `handling_unit(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order_line(inbound_id)` | `inbound_order(inbound_id)` | Yes | CASCADE |
| `inbound_order_line(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order_line(purchase_order_line_id)` | `purchase_order_line(purchase_order_line_id)` | No | RESTRICT/NO ACTION |
| `inbound_order_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order(owner_id, vendor_id)` | `business_partner(owner_id, partner_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `inbound_order(vendor_id)` | `business_partner(partner_id)` | Yes | RESTRICT/NO ACTION |
| `inbound_order(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_execution(destination_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_execution(executed_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_execution(internal_move_line_id)` | `internal_move_order_line(internal_move_line_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_execution(movement_id)` | `inventory_movement(movement_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_execution(source_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order_line(assigned_to)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `internal_move_order_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order_line(internal_move_id)` | `internal_move_order(internal_move_id)` | Yes | CASCADE |
| `internal_move_order_line(source_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order_line(target_location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order(approved_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `internal_move_order(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order(internal_move_type_id)` | `internal_move_type(internal_move_type_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order(reason_code_id)` | `reason_code(reason_code_id)` | No | RESTRICT/NO ACTION |
| `internal_move_order(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `internal_move_order(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment_line(handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `inventory_adjustment_line(inventory_adjustment_id)` | `inventory_adjustment(inventory_adjustment_id)` | Yes | CASCADE |
| `inventory_adjustment_line(inventory_adjustment_type_id)` | `inventory_adjustment_type(inventory_adjustment_type_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment_line(inventory_status_id)` | `inventory_status(inventory_status_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment_line(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment_line(location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment_line(lot_id)` | `inventory_lot(lot_id)` | No | RESTRICT/NO ACTION |
| `inventory_adjustment_line(movement_id)` | `inventory_movement(movement_id)` | No | RESTRICT/NO ACTION |
| `inventory_adjustment_line(posted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `inventory_adjustment_line(resulting_balance_id)` | `inventory_balance(balance_id)` | No | RESTRICT/NO ACTION |
| `inventory_adjustment_line(source_balance_id)` | `inventory_balance(balance_id)` | No | RESTRICT/NO ACTION |
| `inventory_adjustment_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment(approved_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `inventory_adjustment(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment(reason_code_id)` | `reason_code(reason_code_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_adjustment(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_balance(handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `inventory_balance(inventory_status_id)` | `inventory_status(inventory_status_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_balance(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_balance(location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_balance(lot_id)` | `inventory_lot(lot_id)` | No | RESTRICT/NO ACTION |
| `inventory_balance(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_balance(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_balance(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_balance(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_lot(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `inventory_lot(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_lot(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_lot(quality_status_id)` | `quality_status(quality_status_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_movement(from_location_id)` | `warehouse_location(location_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(from_status_id)` | `inventory_status(inventory_status_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_movement(lot_id)` | `inventory_lot(lot_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(movement_type_id)` | `movement_type(movement_type_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_movement(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_movement(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_movement(reason_code_id)` | `reason_code(reason_code_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(serial_id)` | `serial_number(serial_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(to_location_id)` | `warehouse_location(location_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(to_status_id)` | `inventory_status(inventory_status_id)` | No | RESTRICT/NO ACTION |
| `inventory_movement(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_movement(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_reservation(balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_reservation(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_reservation(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_reservation(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_reservation(outbound_line_id)` | `outbound_order_line(outbound_line_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_reservation(picking_strategy_id)` | `picking_strategy(picking_strategy_id)` | No | RESTRICT/NO ACTION |
| `inventory_reservation(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change_line(destination_balance_id)` | `inventory_balance(balance_id)` | No | RESTRICT/NO ACTION |
| `inventory_status_change_line(inventory_status_change_id)` | `inventory_status_change(inventory_status_change_id)` | Yes | CASCADE |
| `inventory_status_change_line(movement_id)` | `inventory_movement(movement_id)` | No | RESTRICT/NO ACTION |
| `inventory_status_change_line(posted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `inventory_status_change_line(source_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change_line(target_inventory_status_id)` | `inventory_status(inventory_status_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change(approved_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `inventory_status_change(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change(reason_code_id)` | `reason_code(reason_code_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `inventory_status_change(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `item_barcode(item_id)` | `item(item_id)` | Yes | CASCADE |
| `item_barcode(uom_id)` | `uom(uom_id)` | No | RESTRICT/NO ACTION |
| `item_category(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `item_category(owner_id, parent_category_id)` | `item_category(owner_id, category_id)` | No | RESTRICT/NO ACTION |
| `item_uom(item_id)` | `item(item_id)` | Yes | CASCADE |
| `item_uom(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `item(base_uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `item(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `item(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `item(owner_id, category_id)` | `item_category(owner_id, category_id)` | No | RESTRICT/NO ACTION |
| `item(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `menu_permission(menu_id)` | `app_menu(menu_id)` | Yes | CASCADE |
| `menu_permission(permission_id)` | `app_permission(permission_id)` | Yes | CASCADE |
| `organization(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `organization(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_check_exception(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_exception(outbound_check_line_id)` | `outbound_check_line(outbound_check_line_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_exception(resolved_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_check_exception(status_id)` | `outbound_check_exception_status(outbound_check_exception_status_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_line(checked_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_check_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_line(outbound_check_id)` | `outbound_check(outbound_check_id)` | Yes | CASCADE |
| `outbound_check_line(outbound_check_result_id)` | `outbound_check_result(outbound_check_result_id)` | No | RESTRICT/NO ACTION |
| `outbound_check_line(staging_line_id)` | `outbound_staging_line(staging_line_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_resolution_movement(movement_id)` | `inventory_movement(movement_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_resolution_movement(outbound_check_resolution_id)` | `outbound_check_resolution(outbound_check_resolution_id)` | Yes | CASCADE |
| `outbound_check_resolution_reservation(outbound_check_resolution_id)` | `outbound_check_resolution(outbound_check_resolution_id)` | Yes | CASCADE |
| `outbound_check_resolution_reservation(reservation_id)` | `inventory_reservation(reservation_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_resolution(approved_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_check_resolution(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_resolution(outbound_check_exception_id)` | `outbound_check_exception(outbound_check_exception_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check_resolution(outbound_check_resolution_type_id)` | `outbound_check_resolution_type(outbound_check_resolution_type_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check(checked_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_check(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_check(parent_check_id)` | `outbound_check(outbound_check_id)` | No | RESTRICT/NO ACTION |
| `outbound_check(staging_id)` | `outbound_staging(staging_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order_line(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order_line(outbound_id)` | `outbound_order(outbound_id)` | Yes | CASCADE |
| `outbound_order_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order(customer_id)` | `business_partner(partner_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order(owner_id, customer_id)` | `business_partner(owner_id, partner_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order(owner_id, ship_to_partner_id)` | `business_partner(owner_id, partner_id)` | No | RESTRICT/NO ACTION |
| `outbound_order(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_order(ship_to_partner_id)` | `business_partner(partner_id)` | No | RESTRICT/NO ACTION |
| `outbound_order(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_order(validated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_order(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_return_policy(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_return_policy(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_return_policy(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_return_policy(return_inventory_status_id)` | `inventory_status(inventory_status_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_return_policy(return_location_id, warehouse_id)` | `warehouse_location(location_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_return_policy(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_return_policy(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging_line(pick_execution_id)` | `pick_execution(pick_execution_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging_line(staging_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging_line(staging_id)` | `outbound_staging(staging_id)` | Yes | CASCADE |
| `outbound_staging_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging(outbound_id)` | `outbound_order(outbound_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging(staged_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_staging(staging_location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_staging(wave_id)` | `outbound_wave(wave_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_validation_result_detail(outbound_validation_rule_id)` | `outbound_validation_rule(outbound_validation_rule_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_validation_result_detail(validation_run_id)` | `outbound_validation_run(validation_run_id)` | Yes | CASCADE |
| `outbound_validation_rule(validation_severity_id)` | `validation_severity(validation_severity_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_validation_run(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_validation_run(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_validation_run(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_validation_run(outbound_id)` | `outbound_order(outbound_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_validation_run(validated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `outbound_wave_order(added_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave_order(outbound_id)` | `outbound_order(outbound_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave_order(wave_id)` | `outbound_wave(wave_id)` | Yes | CASCADE |
| `outbound_wave(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave(outbound_wave_type_id)` | `outbound_wave_type(outbound_wave_type_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave(picking_strategy_id)` | `picking_strategy(picking_strategy_id)` | No | RESTRICT/NO ACTION |
| `outbound_wave(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `outbound_wave(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `packing_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `packing_line(handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `packing_line(movement_id)` | `inventory_movement(movement_id)` | Yes | RESTRICT/NO ACTION |
| `packing_line(outbound_check_line_id)` | `outbound_check_line(outbound_check_line_id)` | Yes | RESTRICT/NO ACTION |
| `packing_line(packing_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `packing_line(packing_id)` | `packing(packing_id)` | Yes | CASCADE |
| `packing_line(pick_task_id)` | `pick_task(pick_task_id)` | Yes | RESTRICT/NO ACTION |
| `packing_line(source_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `packing_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `packing(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `packing(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `packing(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `packing(outbound_id)` | `outbound_order(outbound_id)` | Yes | RESTRICT/NO ACTION |
| `packing(packed_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `packing(packing_location_id)` | `warehouse_location(location_id)` | No | RESTRICT/NO ACTION |
| `packing(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `pick_execution(movement_id)` | `inventory_movement(movement_id)` | Yes | RESTRICT/NO ACTION |
| `pick_execution(pick_task_id)` | `pick_task(pick_task_id)` | Yes | RESTRICT/NO ACTION |
| `pick_execution(picked_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `pick_execution(source_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `pick_execution(staging_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `pick_execution(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(assigned_to)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `pick_task(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(outbound_line_id)` | `outbound_order_line(outbound_line_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(reservation_id)` | `inventory_reservation(reservation_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(short_reason_code_id)` | `reason_code(reason_code_id)` | No | RESTRICT/NO ACTION |
| `pick_task(source_location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(target_location_id)` | `warehouse_location(location_id)` | No | RESTRICT/NO ACTION |
| `pick_task(task_priority_id)` | `task_priority(task_priority_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(task_status_id)` | `task_status(task_status_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(task_type_id)` | `task_type(task_type_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `pick_task(wave_id)` | `outbound_wave(wave_id)` | Yes | RESTRICT/NO ACTION |
| `picking_strategy_rule(inventory_status_id)` | `inventory_status(inventory_status_id)` | No | RESTRICT/NO ACTION |
| `picking_strategy_rule(picking_sort_method_id)` | `picking_sort_method(picking_sort_method_id)` | Yes | RESTRICT/NO ACTION |
| `picking_strategy_rule(picking_strategy_id)` | `picking_strategy(picking_strategy_id)` | Yes | CASCADE |
| `picking_strategy_rule(zone_id)` | `warehouse_zone(zone_id)` | No | RESTRICT/NO ACTION |
| `picking_strategy(owner_id)` | `organization(organization_id)` | No | RESTRICT/NO ACTION |
| `picking_strategy(warehouse_id)` | `warehouse(warehouse_id)` | No | RESTRICT/NO ACTION |
| `purchase_order_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order_line(owner_id, item_id)` | `item(owner_id, item_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order_line(purchase_order_id, owner_id)` | `purchase_order(purchase_order_id, owner_id)` | Yes | CASCADE |
| `purchase_order_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order(owner_id, vendor_id)` | `business_partner(owner_id, partner_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `purchase_order(vendor_id)` | `business_partner(partner_id)` | Yes | RESTRICT/NO ACTION |
| `purchase_order(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_strategy_rule(category_id)` | `item_category(category_id)` | No | RESTRICT/NO ACTION |
| `putaway_strategy_rule(location_type_id)` | `location_type(location_type_id)` | No | RESTRICT/NO ACTION |
| `putaway_strategy_rule(putaway_strategy_id)` | `putaway_strategy(putaway_strategy_id)` | Yes | CASCADE |
| `putaway_strategy_rule(zone_id)` | `warehouse_zone(zone_id)` | No | RESTRICT/NO ACTION |
| `putaway_strategy(owner_id)` | `organization(organization_id)` | No | RESTRICT/NO ACTION |
| `putaway_strategy(warehouse_id)` | `warehouse(warehouse_id)` | No | RESTRICT/NO ACTION |
| `putaway_task(assigned_to)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `putaway_task(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `putaway_task(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(lot_id)` | `inventory_lot(lot_id)` | No | RESTRICT/NO ACTION |
| `putaway_task(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(receipt_inventory_id)` | `receipt_inventory(receipt_inventory_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(source_balance_id)` | `inventory_balance(balance_id)` | No | RESTRICT/NO ACTION |
| `putaway_task(source_location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(target_location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(task_priority_id)` | `task_priority(task_priority_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(task_status_id)` | `task_status(task_status_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(task_type_id)` | `task_type(task_type_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `putaway_task(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `quality_inspection(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `quality_inspection(inspected_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `quality_inspection(inspection_result_id)` | `inspection_result(inspection_result_id)` | No | RESTRICT/NO ACTION |
| `quality_inspection(parent_inspection_id)` | `quality_inspection(inspection_id)` | No | RESTRICT/NO ACTION |
| `quality_inspection(quality_status_id)` | `quality_status(quality_status_id)` | Yes | RESTRICT/NO ACTION |
| `quality_inspection(receipt_inventory_id)` | `receipt_inventory(receipt_inventory_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(inspection_id)` | `quality_inspection(inspection_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(parent_quarantine_case_id)` | `quarantine_case(quarantine_case_id)` | No | RESTRICT/NO ACTION |
| `quarantine_case(quarantine_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(receipt_inventory_id)` | `receipt_inventory(receipt_inventory_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_case(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `quarantine_case(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_disposition_type(removal_movement_type_id)` | `movement_type(movement_type_id)` | No | RESTRICT/NO ACTION |
| `quarantine_disposition(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_disposition(decided_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_disposition(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_disposition(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_disposition(inventory_movement_id)` | `inventory_movement(movement_id)` | No | RESTRICT/NO ACTION |
| `quarantine_disposition(quarantine_case_id)` | `quarantine_case(quarantine_case_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_disposition(quarantine_disposition_type_id)` | `quarantine_disposition_type(quarantine_disposition_type_id)` | Yes | RESTRICT/NO ACTION |
| `quarantine_disposition(resulting_balance_id)` | `inventory_balance(balance_id)` | No | RESTRICT/NO ACTION |
| `quarantine_disposition(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `rate_card_line(billing_service_id)` | `billing_service(billing_service_id)` | Yes | RESTRICT/NO ACTION |
| `rate_card_line(category_id)` | `item_category(category_id)` | No | RESTRICT/NO ACTION |
| `rate_card_line(charge_basis_id)` | `charge_basis(charge_basis_id)` | Yes | RESTRICT/NO ACTION |
| `rate_card_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `rate_card_line(item_id)` | `item(item_id)` | No | RESTRICT/NO ACTION |
| `rate_card_line(rate_card_id)` | `rate_card(rate_card_id)` | Yes | CASCADE |
| `rate_card_line(tax_rule_id)` | `tax_rule(tax_rule_id)` | No | RESTRICT/NO ACTION |
| `rate_card_line(uom_id)` | `uom(uom_id)` | No | RESTRICT/NO ACTION |
| `rate_card_line(warehouse_id)` | `warehouse(warehouse_id)` | No | RESTRICT/NO ACTION |
| `rate_card_tier(rate_card_line_id)` | `rate_card_line(rate_card_line_id)` | Yes | CASCADE |
| `rate_card(billing_contract_id)` | `billing_contract(billing_contract_id)` | Yes | RESTRICT/NO ACTION |
| `rate_card(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `rate_card(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `rate_card(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `rate_card(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `reason_code(module_code)` | `app_module(code)` | Yes | RESTRICT/NO ACTION |
| `receipt_inventory(base_uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_inventory(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_inventory(handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `receipt_inventory(initial_balance_id)` | `inventory_balance(balance_id)` | No | RESTRICT/NO ACTION |
| `receipt_inventory(initial_inventory_status_id)` | `inventory_status(inventory_status_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_inventory(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_inventory(lot_id)` | `inventory_lot(lot_id)` | No | RESTRICT/NO ACTION |
| `receipt_inventory(receipt_line_id)` | `receipt_line(receipt_line_id)` | Yes | CASCADE |
| `receipt_inventory(received_location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_inventory(source_uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_line_serial(receipt_inventory_id)` | `receipt_inventory(receipt_inventory_id)` | Yes | CASCADE |
| `receipt_line_serial(serial_id)` | `serial_number(serial_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_line(inbound_line_id)` | `inbound_order_line(inbound_line_id)` | No | RESTRICT/NO ACTION |
| `receipt_line(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `receipt_line(receipt_id)` | `receipt(receipt_id)` | Yes | CASCADE |
| `receipt_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `receipt(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `receipt(dock_location_id)` | `warehouse_location(location_id)` | No | RESTRICT/NO ACTION |
| `receipt(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `receipt(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `receipt(inbound_id)` | `inbound_order(inbound_id)` | No | RESTRICT/NO ACTION |
| `receipt(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `receipt(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `receipt(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `rework_task(assigned_to)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `rework_task(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `rework_task(quarantine_disposition_id)` | `quarantine_disposition(quarantine_disposition_id)` | Yes | RESTRICT/NO ACTION |
| `rework_task(reinspection_id)` | `quality_inspection(inspection_id)` | No | RESTRICT/NO ACTION |
| `rework_task(task_priority_id)` | `task_priority(task_priority_id)` | Yes | RESTRICT/NO ACTION |
| `rework_task(task_status_id)` | `task_status(task_status_id)` | Yes | RESTRICT/NO ACTION |
| `rework_task(task_type_id)` | `task_type(task_type_id)` | Yes | RESTRICT/NO ACTION |
| `rework_task(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `role_permission(granted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `role_permission(permission_id)` | `app_permission(permission_id)` | Yes | CASCADE |
| `role_permission(role_id)` | `app_role(role_id)` | Yes | CASCADE |
| `serial_number(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `serial_number(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `serial_number(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_driver(assigned_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_driver(driver_id)` | `carrier_driver(driver_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_driver(shipment_id)` | `shipment(shipment_id)` | Yes | CASCADE |
| `shipment_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_line(movement_id)` | `inventory_movement(movement_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_line(packing_line_id)` | `packing_line(packing_line_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_line(shipment_id)` | `shipment(shipment_id)` | Yes | CASCADE |
| `shipment_line(source_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_order(added_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_order(outbound_id)` | `outbound_order(outbound_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_order(removed_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `shipment_order(shipment_id)` | `shipment(shipment_id)` | Yes | CASCADE |
| `shipment_packing(packing_id)` | `packing(packing_id)` | Yes | RESTRICT/NO ACTION |
| `shipment_packing(shipment_id)` | `shipment(shipment_id)` | Yes | CASCADE |
| `shipment(carrier_service_id)` | `carrier_service(carrier_service_id)` | No | RESTRICT/NO ACTION |
| `shipment(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `shipment(dispatched_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `shipment(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `shipment(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `shipment(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `shipment(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `shipment(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count_line(balance_id)` | `inventory_balance(balance_id)` | No | RESTRICT/NO ACTION |
| `stock_count_line(counted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `stock_count_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count_line(handling_unit_id)` | `handling_unit(handling_unit_id)` | No | RESTRICT/NO ACTION |
| `stock_count_line(inventory_status_id)` | `inventory_status(inventory_status_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count_line(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count_line(location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count_line(lot_id)` | `inventory_lot(lot_id)` | No | RESTRICT/NO ACTION |
| `stock_count_line(posted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `stock_count_line(stock_count_id)` | `stock_count(stock_count_id)` | Yes | CASCADE |
| `stock_count_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count_line(variance_movement_id)` | `inventory_movement(movement_id)` | No | RESTRICT/NO ACTION |
| `stock_count(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count(owner_id, warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count(posted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `stock_count(reason_code_id)` | `reason_code(reason_code_id)` | No | RESTRICT/NO ACTION |
| `stock_count(reviewed_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `stock_count(stock_count_type_id)` | `stock_count_type(stock_count_type_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `stock_count(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `task_status_transition(from_status_id)` | `task_status(task_status_id)` | Yes | RESTRICT/NO ACTION |
| `task_status_transition(required_permission_id)` | `app_permission(permission_id)` | No | RESTRICT/NO ACTION |
| `task_status_transition(to_status_id)` | `task_status(task_status_id)` | Yes | RESTRICT/NO ACTION |
| `tax_rule(owner_id)` | `organization(organization_id)` | No | RESTRICT/NO ACTION |
| `transfer_dispatch_line(confirmed_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `transfer_dispatch_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_dispatch_line(movement_id)` | `inventory_movement(movement_id)` | No | RESTRICT/NO ACTION |
| `transfer_dispatch_line(source_balance_id)` | `inventory_balance(balance_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_dispatch_line(transfer_dispatch_id)` | `transfer_dispatch(transfer_dispatch_id)` | Yes | CASCADE |
| `transfer_dispatch_line(transfer_line_id)` | `transfer_order_line(transfer_line_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_dispatch_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_dispatch(carrier_partner_id)` | `business_partner(partner_id)` | No | RESTRICT/NO ACTION |
| `transfer_dispatch(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_dispatch(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_dispatch(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_dispatch(transfer_id)` | `transfer_order(transfer_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order_line(item_id)` | `item(item_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order_line(requested_lot_id)` | `inventory_lot(lot_id)` | No | RESTRICT/NO ACTION |
| `transfer_order_line(transfer_id)` | `transfer_order(transfer_id)` | Yes | CASCADE |
| `transfer_order_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(approved_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `transfer_order(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(owner_id, source_warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(owner_id, target_warehouse_id)` | `warehouse_owner(owner_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(source_warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(target_warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_order(updated_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt_line(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt_line(destination_balance_id)` | `inventory_balance(balance_id)` | No | RESTRICT/NO ACTION |
| `transfer_receipt_line(movement_id)` | `inventory_movement(movement_id)` | No | RESTRICT/NO ACTION |
| `transfer_receipt_line(posted_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `transfer_receipt_line(target_location_id)` | `warehouse_location(location_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt_line(transfer_dispatch_line_id)` | `transfer_dispatch_line(transfer_dispatch_line_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt_line(transfer_receipt_id)` | `transfer_receipt(transfer_receipt_id)` | Yes | CASCADE |
| `transfer_receipt_line(uom_id)` | `uom(uom_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt(created_by)` | `app_account(account_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt(document_type_id)` | `document_type(document_type_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt(document_type_id, status_id)` | `document_status(document_type_id, status_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt(transfer_id)` | `transfer_order(transfer_id)` | Yes | RESTRICT/NO ACTION |
| `transfer_receipt(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `warehouse_location(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `warehouse_location(location_type_id)` | `location_type(location_type_id)` | Yes | RESTRICT/NO ACTION |
| `warehouse_location(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `warehouse_location(zone_id, warehouse_id)` | `warehouse_zone(zone_id, warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `warehouse_owner(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `warehouse_owner(owner_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `warehouse_owner(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `warehouse_zone(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `warehouse_zone(warehouse_id)` | `warehouse(warehouse_id)` | Yes | RESTRICT/NO ACTION |
| `warehouse(created_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
| `warehouse(operator_id)` | `organization(organization_id)` | Yes | RESTRICT/NO ACTION |
| `warehouse(updated_by)` | `app_account(account_id)` | No | RESTRICT/NO ACTION |
