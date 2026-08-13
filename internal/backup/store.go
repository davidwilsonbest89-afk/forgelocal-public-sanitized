package backup

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct{ db *sql.DB }

func OpenSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}
func (s *SQLiteStore) Close() error { return s.db.Close() }
func now() string                   { return time.Now().UTC().Format(time.RFC3339Nano) }
func (s *SQLiteStore) audit(tx *sql.Tx, event, id, correlation string, details map[string]string) error {
	data, err := json.Marshal(details)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO audit_events(event_type, entity_id, correlation_id, details_json, created_at) VALUES(?,?,?,?,?)`, event, id, correlation, string(data), now())
	return err
}
func (s *SQLiteStore) BeginBackup(b Backup, correlation string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	t := now()
	_, err = tx.Exec(`INSERT INTO backup_operations(id,profile_id,state,artifact_path,key_id,correlation_id,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`, b.ID, b.ProfileID, "staging", b.ArtifactPath, b.KeyID, correlation, t, t)
	if err != nil {
		return err
	}
	if err = s.audit(tx, "backup.staging", b.ID, correlation, map[string]string{"profile_id": b.ProfileID, "key_id": b.KeyID}); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *SQLiteStore) MarkPublished(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE backup_operations SET state='published_unregistered',updated_at=? WHERE id=? AND state='staging'`, now(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return errors.New("backup operation not in staging state")
	}
	if err = s.audit(tx, "backup.published", id, "backup:"+id, map[string]string{}); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *SQLiteStore) CommitBackup(b Backup) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO backups(id,profile_id,artifact_path,key_id,sha256,created_at) VALUES(?,?,?,?,?,?)`, b.ID, b.ProfileID, b.ArtifactPath, b.KeyID, b.SHA256, b.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE backup_operations SET state='committed',updated_at=? WHERE id=? AND state='published_unregistered'`, now(), b.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return errors.New("backup operation not published")
	}
	if err = s.audit(tx, "backup.committed", b.ID, "backup:"+b.ID, map[string]string{"sha256": b.SHA256}); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *SQLiteStore) GetBackup(id string) (Backup, error) {
	var b Backup
	var created string
	err := s.db.QueryRow(`SELECT id,profile_id,artifact_path,key_id,sha256,created_at FROM backups WHERE id=?`, id).Scan(&b.ID, &b.ProfileID, &b.ArtifactPath, &b.KeyID, &b.SHA256, &created)
	if err != nil {
		return Backup{}, err
	}
	b.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	return b, nil
}
func (s *SQLiteStore) HasBackup(id string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(1) FROM backups WHERE id=?`, id).Scan(&n)
	return n > 0, err
}
func (s *SQLiteStore) RecoverBackup(b Backup, correlation string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT OR IGNORE INTO backups(id,profile_id,artifact_path,key_id,sha256,created_at) VALUES(?,?,?,?,?,?)`, b.ID, b.ProfileID, b.ArtifactPath, b.KeyID, b.SHA256, b.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE backup_operations SET state='committed',updated_at=?,error_code='' WHERE id=?`, now(), b.ID)
	if err != nil {
		return err
	}
	if err = s.audit(tx, "backup.recovered", b.ID, correlation, map[string]string{"reason": correlation}); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *SQLiteStore) Quarantine(id, code string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`UPDATE backup_operations SET state='quarantined',updated_at=?,error_code=? WHERE id=?`, now(), code, id)
	if err != nil {
		return err
	}
	if err = s.audit(tx, "backup.quarantined", id, "backup:"+id, map[string]string{"code": code}); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *SQLiteStore) BeginRestore(r Restore) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	t := now()
	_, err = tx.Exec(`INSERT INTO restore_operations(id,backup_id,source_profile_id,target_profile_id,target_path,state,correlation_id,created_at,updated_at) VALUES(?,?,?,?,?,'staging',?,?,?)`, r.ID, r.BackupID, r.SourceProfileID, r.TargetProfileID, r.TargetPath, r.CorrelationID, t, t)
	if err != nil {
		return err
	}
	if err = s.audit(tx, "restore.staging", r.ID, r.CorrelationID, map[string]string{"backup_id": r.BackupID, "target_profile_id": r.TargetProfileID}); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *SQLiteStore) CommitRestore(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`UPDATE restore_operations SET state='committed',updated_at=? WHERE id=? AND state='staging'`, now(), id)
	if err != nil {
		return err
	}
	if err = s.audit(tx, "restore.committed", id, "restore:"+id, map[string]string{}); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *SQLiteStore) FailRestore(id, code string) error {
	_, err := s.db.Exec(`UPDATE restore_operations SET state='failed',updated_at=?,error_code=? WHERE id=?`, now(), code, id)
	return err
}
func (s *SQLiteStore) DB() *sql.DB    { return s.db }
func (s *SQLiteStore) String() string { return fmt.Sprintf("SQLiteStore(%p)", s.db) }
