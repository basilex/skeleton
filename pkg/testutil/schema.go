package testutil

// DefaultSchema contains the base schema for testing.
// This creates minimal tables needed for repository tests.
const DefaultSchema = `
-- Users table
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	description TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL
);

-- User roles (many-to-many)
CREATE TABLE IF NOT EXISTS user_roles (
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, role_id)
);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
	id SERIAL PRIMARY KEY,
	name TEXT UNIQUE NOT NULL
);

-- Role permissions (many-to-many)
CREATE TABLE IF NOT EXISTS role_permissions (
	role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
	PRIMARY KEY (role_id, permission_id)
);

-- Files table
CREATE TABLE IF NOT EXISTS files (
	id TEXT PRIMARY KEY,
	owner_id TEXT,
	filename TEXT NOT NULL,
	stored_name TEXT NOT NULL,
	mime_type TEXT NOT NULL,
	size BIGINT NOT NULL,
	path TEXT NOT NULL,
	storage_provider TEXT NOT NULL,
	checksum TEXT,
	metadata JSONB,
	access_level TEXT NOT NULL,
	uploaded_at TIMESTAMP NOT NULL,
	expires_at TIMESTAMP,
	processed_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- File uploads table
CREATE TABLE IF NOT EXISTS file_uploads (
	id TEXT PRIMARY KEY,
	file_id TEXT REFERENCES files(id) ON DELETE CASCADE,
	upload_url TEXT NOT NULL,
	fields JSONB,
	status TEXT NOT NULL,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL
);

-- File processings table
CREATE TABLE IF NOT EXISTS file_processings (
	id TEXT PRIMARY KEY,
	file_id TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
	operation TEXT NOT NULL,
	options JSONB,
	status TEXT NOT NULL,
	result_file_id TEXT REFERENCES files(id) ON DELETE SET NULL,
	error TEXT,
	started_at TIMESTAMP,
	completed_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
	id TEXT PRIMARY KEY,
	user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
	channel TEXT NOT NULL,
	subject TEXT,
	content TEXT NOT NULL,
	html_content TEXT,
	status TEXT NOT NULL,
	priority TEXT NOT NULL,
	scheduled_at TIMESTAMP,
	sent_at TIMESTAMP,
	delivered_at TIMESTAMP,
	failed_at TIMESTAMP,
	failure_reason TEXT,
	attempts INTEGER NOT NULL DEFAULT 0,
	max_attempts INTEGER NOT NULL DEFAULT 3,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- Notification templates
CREATE TABLE IF NOT EXISTS notification_templates (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	channel TEXT NOT NULL,
	subject_template TEXT,
	body_template TEXT NOT NULL,
	variables JSONB,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- Notification preferences
CREATE TABLE IF NOT EXISTS notification_preferences (
	id TEXT PRIMARY KEY,
	user_id TEXT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	email_enabled BOOLEAN NOT NULL DEFAULT true,
	push_enabled BOOLEAN NOT NULL DEFAULT true,
	sms_enabled BOOLEAN NOT NULL DEFAULT false,
	in_app_enabled BOOLEAN NOT NULL DEFAULT true,
	quiet_hours_start TIME,
	quiet_hours_end TIME,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- Audit records table
CREATE TABLE IF NOT EXISTS audit_records (
	id TEXT PRIMARY KEY,
	user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	action TEXT NOT NULL,
	resource_type TEXT NOT NULL,
	resource_id TEXT,
	details JSONB,
	ip_address TEXT,
	user_agent TEXT,
	created_at TIMESTAMP NOT NULL
);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	payload JSONB NOT NULL,
	status TEXT NOT NULL,
	priority INTEGER NOT NULL DEFAULT 0,
	scheduled_at TIMESTAMP NOT NULL,
	attempts INTEGER NOT NULL DEFAULT 0,
	max_attempts INTEGER NOT NULL DEFAULT 3,
	error TEXT,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	started_at TIMESTAMP,
	completed_at TIMESTAMP
);

-- Task schedules
CREATE TABLE IF NOT EXISTS schedules (
	id TEXT PRIMARY KEY,
	task_type TEXT NOT NULL,
	cron_expression TEXT NOT NULL,
	payload JSONB,
	enabled BOOLEAN NOT NULL DEFAULT true,
	last_run TIMESTAMP,
	next_run TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- Dead letter queue
CREATE TABLE IF NOT EXISTS dead_letters (
	id TEXT PRIMARY KEY,
	original_task_id TEXT,
	type TEXT NOT NULL,
	payload JSONB NOT NULL,
	error TEXT NOT NULL,
	retries INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMP NOT NULL
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_at ON tasks(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_audit_records_user_id ON audit_records(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_records_created_at ON audit_records(created_at);
`
