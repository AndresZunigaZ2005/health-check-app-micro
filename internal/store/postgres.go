package store

import (
	"database/sql"
	"embed"
	"time"

	"github.com/example/monitor/internal/models"
	_ "github.com/lib/pq"
)

//go:embed ../migrations/*.sql
var migrationsFS embed.FS

type Store struct {
	db *sql.DB
}

func New(postgresURL string) (*Store, error) {
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.runMigrations(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) runMigrations() error {
	// carga todos los .sql en migrations y los ejecuta en orden (aquí solo uno)
	b, err := migrationsFS.ReadFile(".. /migrations/001_create_services.sql") // not valid path; use correct path below
	if err != nil {
		// alternativa: try direct relative path
		b, err = migrationsFS.ReadFile("001_create_services.sql")
		if err != nil {
			return err
		}
	}
	_, err = s.db.Exec(string(b))
	return err
}

// AddOrUpdate inserta o actualiza
func (s *Store) AddOrUpdate(m *models.MonitoredService) error {
	_, err := s.db.Exec(
		`INSERT INTO monitored_service (name, endpoint, frequency_seconds, emails, last_status, last_checked)
         VALUES ($1,$2,$3,$4,$5,$6)
         ON CONFLICT (name) DO UPDATE
           SET endpoint = EXCLUDED.endpoint,
               frequency_seconds = EXCLUDED.frequency_seconds,
               emails = EXCLUDED.emails`,
		m.Name, m.Endpoint, m.Frequency, pq.Array(m.Emails), string(m.LastStatus), m.LastChecked,
	)
	return err
}

func (s *Store) GetAll() ([]*models.MonitoredService, error) {
	rows, err := s.db.Query(`SELECT name, endpoint, frequency_seconds, emails, last_status, last_checked FROM monitored_service`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*models.MonitoredService{}
	for rows.Next() {
		var name, endpoint, lastStatus sql.NullString
		var freq sql.NullInt64
		var emailsArr []sql.NullString
		var lastChecked sql.NullTime

		// We will scan emails into pq.StringArray for convenience
		var emails pq.StringArray
		if err := rows.Scan(&name, &endpoint, &freq, &emails, &lastStatus, &lastChecked); err != nil {
			return nil, err
		}

		m := &models.MonitoredService{
			Name:      name.String,
			Endpoint:  endpoint.String,
			Frequency: int(freq.Int64),
			Emails:    []string(emails),
		}
		if lastStatus.Valid {
			m.LastStatus = models.ServiceStatus(lastStatus.String)
		}
		if lastChecked.Valid {
			m.LastChecked = lastChecked.Time
		}
		out = append(out, m)
	}
	return out, nil
}

func (s *Store) Get(name string) (*models.MonitoredService, error) {
	row := s.db.QueryRow(`SELECT name, endpoint, frequency_seconds, emails, last_status, last_checked FROM monitored_service WHERE name=$1`, name)
	var nameN, endpoint, lastStatus sql.NullString
	var freq sql.NullInt64
	var emails pq.StringArray
	var lastChecked sql.NullTime
	if err := row.Scan(&nameN, &endpoint, &freq, &emails, &lastStatus, &lastChecked); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	m := &models.MonitoredService{
		Name:      nameN.String,
		Endpoint:  endpoint.String,
		Frequency: int(freq.Int64),
		Emails:    []string(emails),
	}
	if lastStatus.Valid {
		m.LastStatus = models.ServiceStatus(lastStatus.String)
	}
	if lastChecked.Valid {
		m.LastChecked = lastChecked.Time
	}
	return m, nil
}

func (s *Store) UpdateStatus(name string, status models.ServiceStatus, checked time.Time) error {
	_, err := s.db.Exec(`UPDATE monitored_service SET last_status=$1, last_checked=$2 WHERE name=$3`, string(status), checked, name)
	return err
}

func (s *Store) Delete(name string) error {
	_, err := s.db.Exec(`DELETE FROM monitored_service WHERE name=$1`, name)
	return err
}
