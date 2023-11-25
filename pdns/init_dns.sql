ALTER TABLE records ADD INDEX idx_name_domain_id_disabled (name, domain_id, disabled);
ALTER TABLE records ADD INDEX idx_type_name_disabled (type, name, disabled);