package users

import (
	"errors"
	"fmt"
	"github.com/AlessandroArosio/bookstore-utils-go/rest_errors"
	"github.com/alessandroarosio/bookstore-utils-go/logger"
	"github.com/alessandroarosio/bookstore_users-API/datasources/mysql/users_db"
	"github.com/alessandroarosio/bookstore_users-API/utils/mysql_utils"
	"strings"
)

const (
	queryInsertUser             = "INSERT INTO users(first_name, last_name, email, date_created, status, password) VALUES (?, ?, ?, ?, ?, ?);"
	queryGetUser                = "SELECT id, first_name, last_name, email, date_created, status FROM users WHERE id = ?;"
	queryUpdateUser             = "UPDATE users SET first_name=?, last_name=?, email=? WHERE id=?;"
	queryDeleteUser             = "DELETE FROM users WHERE id=?;"
	queryFindByStatus           = "SELECT id, first_name, last_name, email, date_created, status FROM users WHERE status=?;"
	queryFindByEmailAndPassword = "SELECT id, first_name, last_name, email, date_created, status FROM users WHERE email=? AND password=? AND status=?;"
)

func (user *User) Get() *rest_errors.RestErr {
	stmt, err := users_db.Client.Prepare(queryGetUser)
	if err != nil {
		logger.Error("error when trying to prepare the get user statement", err)
		return rest_errors.NewInternalServerError("error when trying to get user", errors.New("database error"))
	}
	defer stmt.Close()

	result := stmt.QueryRow(user.Id)
	if getErr := result.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.DateCreated, &user.Status); getErr != nil {
		logger.Error("error when trying to get user by ID", err)
		return rest_errors.NewInternalServerError("error when trying to get user by ID", errors.New("database error"))
	}
	return nil
}

func (user *User) Save() *rest_errors.RestErr {
	stmt, err := users_db.Client.Prepare(queryInsertUser)
	if err != nil {
		logger.Error("error when trying to prepare SaveUser statement", err)
		return rest_errors.NewInternalServerError("error when trying to get user by ID", errors.New("database error"))
	}
	defer stmt.Close()

	insertResult, saveErr := stmt.Exec(user.FirstName, user.LastName, user.Email, user.DateCreated, user.Status, user.Password)
	if saveErr != nil {
		logger.Error("error when trying to save user", err)
		return rest_errors.NewInternalServerError("error when trying save user", errors.New("database error"))
	}

	userId, saveErr := insertResult.LastInsertId()
	if saveErr != nil {
		logger.Error("error when trying to get last insert after creating a new user", err)
		rest_errors.NewInternalServerError("error when trying to get last insert", errors.New("database error"))
	}

	user.Id = userId
	return nil
}

func (user *User) Update() *rest_errors.RestErr {
	stmt, err := users_db.Client.Prepare(queryUpdateUser)
	if err != nil {
		logger.Error("error when trying to prepare UpdateUser statement", err)
		return rest_errors.NewInternalServerError("error when trying to prepare statement updateUser",
			errors.New("database error"))
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.FirstName, user.LastName, user.Email, user.Id)
	if err != nil {
		logger.Error("error when trying to update user", err)
		return rest_errors.NewInternalServerError("error when trying to update user ", errors.New("database error"))
	}
	return nil
}

func (user *User) Delete() *rest_errors.RestErr {
	stmt, err := users_db.Client.Prepare(queryDeleteUser)
	if err != nil {
		logger.Error("error when trying to prepare DeleteUser statement", err)
		return rest_errors.NewInternalServerError("error when trying to prepare delete statement",
			errors.New("database error"))
	}
	defer stmt.Close()

	if _, err := stmt.Exec(user.Id); err != nil {
		logger.Error("error when trying to delete a user", err)
		return rest_errors.NewInternalServerError("error when trying to delete user", errors.New("database error"))
	}
	return nil
}

func (user *User) FindByStatus(status string) ([]User, *rest_errors.RestErr) {
	stmt, err := users_db.Client.Prepare(queryFindByStatus)
	if err != nil {
		logger.Error("error when trying to prepare FindByStatus statement", err)
		return nil, rest_errors.NewInternalServerError("error when trying to prepare find by status statement",
			errors.New("database error"))
	}
	defer stmt.Close()

	rows, err := stmt.Query(status)
	if err != nil {
		logger.Error("error when trying to find users by status", err)
		return nil, rest_errors.NewInternalServerError("error when trying to find users by status",
			errors.New("database error"))
	}
	defer rows.Close()

	results := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.DateCreated, &user.Status); err != nil {
			logger.Error("error when trying to scan user's row into struct", err)
			return nil, rest_errors.NewInternalServerError("error when trying to scan user's row into struct",
				errors.New("database error"))
		}
		results = append(results, user)
	}

	if len(results) == 0 {
		return nil, rest_errors.NewNotFoundError(fmt.Sprintf("no users matching status %s", status))
	}
	return results, nil
}

func (user *User) FindByEmailAndPassword() *rest_errors.RestErr {
	stmt, err := users_db.Client.Prepare(queryFindByEmailAndPassword)
	if err != nil {
		logger.Error("error when trying to prepare the get user by email and password statement", err)
		return rest_errors.NewInternalServerError("error when trying to prepare statement get user by and password",
			errors.New("database error"))
	}
	defer stmt.Close()

	result := stmt.QueryRow(user.Email, user.Password, StatusActive)
	if getErr := result.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.DateCreated, &user.Status); getErr != nil {
		if strings.Contains(getErr.Error(), mysql_utils.ErrorNoRows) {
			return rest_errors.NewNotFoundError("invalid user credentials")
		}
		logger.Error("error when trying to get user by email and password", err)
		return rest_errors.NewInternalServerError("error when trying to get user by email and pass",
			errors.New("database error"))
	}
	return nil
}
