package models

import "database/sql"

type Subject struct {
	Name string
}

func InsertSubject(db *sql.DB, subject Subject) (error) {
	_, err := db.Exec("INSERT INTO subjects (name) VALUES ($1) ON CONFLICT DO NOTHING", subject.Name)
	if err != nil {
		return err
	}
	return nil
}

func GetSubject(db *sql.DB, name string) (Subject, error) {
	var subject Subject
	err := db.QueryRow("SELECT name FROM subjects WHERE name = $1", name).Scan(&subject.Name)
	if err != nil {
		return subject, err
	}
	return subject, nil
}

func GetAllSubjects(db *sql.DB) ([]Subject, error) {
	subjects := make([]Subject, 0)
	rows, err := db.Query("SELECT name FROM subjects")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var subject Subject
		err = rows.Scan(&subject.Name)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, subject)
	}

	return subjects, nil
}
