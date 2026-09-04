package master

import (
	"gorm.io/gorm"
	model "wms-api/models/master"
)

func MigrateOperational(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AppModule{}, &model.AppPermission{}, &model.DocumentType{}, &model.DocumentStatus{}, &model.DocumentStatusTransition{}, &model.DocumentNumberRule{}, &model.DocumentDailyCounter{}, &model.TaskType{}, &model.TaskStatus{}, &model.TaskStatusTransition{}, &model.TaskPriority{}, &model.PickingSortMethod{}, &model.PickingStrategy{}, &model.PickingStrategyRule{}, &model.PutawayStrategy{}, &model.PutawayStrategyRule{}); err != nil {
			return err
		}
		statements := []string{
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_app_module_code ON app_module (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_app_permission_code ON app_permission (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_document_type_code ON document_type (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_document_status_document_type_id_code ON document_status (document_type_id,code)`,
			`DO $$ BEGIN ALTER TABLE document_status ADD CONSTRAINT fk_ops_document_status_document_type_id FOREIGN KEY (document_type_id) REFERENCES document_type (document_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_status_transition ADD CONSTRAINT fk_ops_document_status_transition_document_type_id FOREIGN KEY (document_type_id) REFERENCES document_type (document_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_number_rule ADD CONSTRAINT fk_ops_document_number_rule_document_type_id FOREIGN KEY (document_type_id) REFERENCES document_type (document_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_task_type_code ON task_type (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_task_status_code ON task_status (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_task_priority_code ON task_priority (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_picking_sort_method_code ON picking_sort_method (code)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_picking_strategy_owner_id_warehouse_id_code ON picking_strategy (owner_id,warehouse_id,code) NULLS NOT DISTINCT`,
			`DO $$ BEGIN ALTER TABLE picking_strategy_rule ADD CONSTRAINT fk_ops_picking_strategy_rule_picking_strategy_id FOREIGN KEY (picking_strategy_id) REFERENCES picking_strategy (picking_strategy_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_putaway_strategy_owner_id_warehouse_id_code ON putaway_strategy (owner_id,warehouse_id,code) NULLS NOT DISTINCT`,
			`DO $$ BEGIN ALTER TABLE putaway_strategy_rule ADD CONSTRAINT fk_ops_putaway_strategy_rule_putaway_strategy_id FOREIGN KEY (putaway_strategy_id) REFERENCES putaway_strategy (putaway_strategy_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_document_status_document_type_id_status_id ON document_status (document_type_id,status_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_document_status_initial ON document_status (document_type_id) WHERE is_initial`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_task_status_initial ON task_status (is_initial) WHERE is_initial`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_active_document_number_rule ON document_number_rule (document_type_id) WHERE is_active`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_document_status_transition_document_type_id_from_status_id_to_status_id ON document_status_transition (document_type_id,from_status_id,to_status_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_task_status_transition_from_status_id_to_status_id ON task_status_transition (from_status_id,to_status_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_task_priority_priority_value ON task_priority (priority_value)`,
			`DO $$ BEGIN ALTER TABLE document_type ADD CONSTRAINT fk_ops_document_type_module_code FOREIGN KEY (module_code) REFERENCES app_module (code); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE app_permission ADD CONSTRAINT fk_ops_app_permission_module_code FOREIGN KEY (module_code) REFERENCES app_module (code); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_status_transition ADD CONSTRAINT fk_ops_document_status_transition_document_type_id_from_status_id FOREIGN KEY (document_type_id,from_status_id) REFERENCES document_status (document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE task_status_transition ADD CONSTRAINT fk_ops_task_status_transition_from_status_id FOREIGN KEY (from_status_id) REFERENCES task_status (task_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_status_transition ADD CONSTRAINT fk_ops_document_status_transition_document_type_id_to_status_id FOREIGN KEY (document_type_id,to_status_id) REFERENCES document_status (document_type_id,status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE task_status_transition ADD CONSTRAINT fk_ops_task_status_transition_to_status_id FOREIGN KEY (to_status_id) REFERENCES task_status (task_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_status_transition ADD CONSTRAINT fk_ops_document_status_transition_required_permission_id FOREIGN KEY (required_permission_id) REFERENCES app_permission (permission_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_status_transition ADD CONSTRAINT ck_ops_document_status_transition_different CHECK (from_status_id <> to_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE task_status_transition ADD CONSTRAINT fk_ops_task_status_transition_required_permission_id FOREIGN KEY (required_permission_id) REFERENCES app_permission (permission_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE task_status_transition ADD CONSTRAINT ck_ops_task_status_transition_different CHECK (from_status_id <> to_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_number_rule ADD CONSTRAINT fk_ops_document_number_rule_created_by FOREIGN KEY (created_by) REFERENCES app_account (account_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_daily_counter ADD CONSTRAINT fk_ops_document_daily_counter_document_type_id FOREIGN KEY (document_type_id) REFERENCES document_type (document_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE picking_strategy ADD CONSTRAINT fk_ops_picking_strategy_owner_id FOREIGN KEY (owner_id) REFERENCES organization (organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE picking_strategy ADD CONSTRAINT fk_ops_picking_strategy_warehouse_id FOREIGN KEY (warehouse_id) REFERENCES warehouse (warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_picking_strategy_rule_picking_strategy_id_sequence_no ON picking_strategy_rule (picking_strategy_id,sequence_no)`,
			`DO $$ BEGIN ALTER TABLE picking_strategy_rule ADD CONSTRAINT fk_ops_picking_strategy_rule_zone_id FOREIGN KEY (zone_id) REFERENCES warehouse_zone (zone_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE picking_strategy_rule ADD CONSTRAINT ck_ops_picking_strategy_sequence CHECK (sequence_no > 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE putaway_strategy ADD CONSTRAINT fk_ops_putaway_strategy_owner_id FOREIGN KEY (owner_id) REFERENCES organization (organization_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE putaway_strategy ADD CONSTRAINT fk_ops_putaway_strategy_warehouse_id FOREIGN KEY (warehouse_id) REFERENCES warehouse (warehouse_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_ops_putaway_strategy_rule_putaway_strategy_id_sequence_no ON putaway_strategy_rule (putaway_strategy_id,sequence_no)`,
			`DO $$ BEGIN ALTER TABLE putaway_strategy_rule ADD CONSTRAINT fk_ops_putaway_strategy_rule_zone_id FOREIGN KEY (zone_id) REFERENCES warehouse_zone (zone_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE putaway_strategy_rule ADD CONSTRAINT ck_ops_putaway_strategy_sequence CHECK (sequence_no > 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE picking_strategy_rule ADD CONSTRAINT fk_ops_picking_strategy_rule_inventory_status_id FOREIGN KEY (inventory_status_id) REFERENCES inventory_status (inventory_status_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE picking_strategy_rule ADD CONSTRAINT fk_ops_picking_strategy_rule_picking_sort_method_id FOREIGN KEY (picking_sort_method_id) REFERENCES picking_sort_method (picking_sort_method_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE putaway_strategy_rule ADD CONSTRAINT fk_ops_putaway_strategy_rule_category_id FOREIGN KEY (category_id) REFERENCES item_category (category_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE putaway_strategy_rule ADD CONSTRAINT fk_ops_putaway_strategy_rule_location_type_id FOREIGN KEY (location_type_id) REFERENCES location_type (location_type_id); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE putaway_strategy_rule ADD CONSTRAINT ck_ops_putaway_percent CHECK (minimum_empty_percent IS NULL OR minimum_empty_percent BETWEEN 0 AND 100); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_number_rule ADD CONSTRAINT ck_ops_number_length CHECK (sequence_length BETWEEN 3 AND 18); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_number_rule ADD CONSTRAINT ck_ops_number_period CHECK (effective_until IS NULL OR effective_until >= effective_from); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_daily_counter ADD CONSTRAINT ck_ops_counter_positive CHECK (last_number > 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE task_priority ADD CONSTRAINT ck_ops_priority_value CHECK (priority_value >= 0); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE task_status ADD CONSTRAINT ck_ops_task_status_flags CHECK ((NOT is_initial OR (NOT is_final AND NOT is_cancelled)) AND (NOT is_cancelled OR is_final)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
			`DO $$ BEGIN ALTER TABLE document_status ADD CONSTRAINT ck_ops_document_status_flags CHECK ((NOT is_initial OR (NOT is_final AND NOT is_cancelled)) AND (NOT is_cancelled OR is_final)); EXCEPTION WHEN duplicate_object THEN NULL; END $$`,
		}
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
