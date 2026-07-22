package migrations

// migrate_smtp_config_to_sys_config is retained as a no-op for databases that
// still have the historical migration version in their upgrade sequence.
func migrate_smtp_config_to_sys_config() error {
	return nil
}
