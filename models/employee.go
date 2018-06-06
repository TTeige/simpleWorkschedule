package models

import "database/sql"

type Employee struct {
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	Affiliation  string
	Username     string
	Admin        bool
}

func InsertEmployee(db *sql.DB, employee Employee) (error) {
	_, err := db.Exec("INSERT INTO employee (first_name, last_name, e_mail, password_hash, affiliation, username) VALUES ($1, $2, $3, $4, $5, $6)ON CONFLICT DO NOTHING",
		&employee.FirstName, &employee.LastName, &employee.Email, &employee.PasswordHash, &employee.Affiliation, &employee.Username)
	if err != nil {
		return err
	}
	return nil
}

func GetEmployee(db *sql.DB, email string) (Employee, error) {
	var employee Employee
	err := db.QueryRow("SELECT first_name, last_name, e_mail, password_hash, affiliation, username, admin FROM employee WHERE e_mail = $1", email).Scan(&employee.FirstName, &employee.LastName, &employee.Email, &employee.PasswordHash, &employee.Affiliation, &employee.Username, &employee.Admin)
	if err != nil {
		return employee, err
	}
	return employee, nil
}

func GetAllEmployees(db *sql.DB) ([]Employee, error) {
	employees := make([]Employee, 0)
	rows, err := db.Query("SELECT first_name, last_name, e_mail, password_hash, affiliation, username, admin FROM employee")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var employee Employee
		err = rows.Scan(&employee.FirstName, &employee.LastName, &employee.Email, &employee.PasswordHash, &employee.Affiliation, &employee.Username, &employee.Admin)
		if err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}

	return employees, nil
}
