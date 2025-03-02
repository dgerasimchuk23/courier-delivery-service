package auth

import (
	"database/sql"
	"delivery/internal/business/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)

	tests := []struct {
		user    models.User
		mock    func()
		wantErr bool
	}{
		{
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mock: func() {
				mock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
				mock.ExpectExec("INSERT INTO users").
					WithArgs("test@example.com", "password123", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mock: func() {
				mock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.mock != nil {
			tt.mock()
		}
		_, err := userStore.CreateUser(tt.user)
		if (err != nil) != tt.wantErr {
			t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)

	tests := []struct {
		email   string
		mock    func()
		wantErr bool
	}{
		{
			email: "test@example.com",
			mock: func() {
				mock.ExpectQuery("SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(1, "test@example.com", "password123"))
			},
			wantErr: false,
		},
		{
			email: "nonexistent@example.com",
			mock: func() {
				mock.ExpectQuery("SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1").
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.mock != nil {
			tt.mock()
		}
		_, err := userStore.GetUserByEmail(tt.email)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestGetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)

	tests := []struct {
		id      int
		mock    func()
		wantErr bool
	}{
		{
			id: 1,
			mock: func() {
				mock.ExpectQuery("SELECT id, email, password, created_at, updated_at FROM users WHERE id = $1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(1, "test@example.com", "password123"))
			},
			wantErr: false,
		},
		{
			id: 2,
			mock: func() {
				mock.ExpectQuery("SELECT id, email, password, created_at, updated_at FROM users WHERE id = $1").
					WithArgs(2).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.mock != nil {
			tt.mock()
		}
		_, err := userStore.GetUserByID(tt.id)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}
