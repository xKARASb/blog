package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/xkarasb/blog/pkg/db/postgres"
	"github.com/xkarasb/blog/pkg/errors"
)

func TestPostgresRepository_AddNewUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()
	repo := &PostgresRepository{DB: &postgres.DB{sqlx.NewDb(db, "postgres")}}

	tests := []struct {
		name        string
		email       string
		setupMock   func()
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "successful insert",
			email: "test@example.com",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{
					"user_id", "email", "password_hash", "role",
					"refresh_token", "refresh_token_expiry_time",
				}).AddRow(
					uuid.New(), "test@example.com", "hashed_password", "user",
					"refresh_token", "2024-12-31T23:59:59Z",
				)

				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs("test@example.com", "password_hash", "user", "refresh_token", sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:  "duplicate user",
			email: "existing@example.com",
			setupMock: func() {
				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs("existing@example.com", "password_hash", "user", "refresh_token", sqlmock.AnyArg()).
					WillReturnError(&pq.Error{Code: "23505"})
			},
			wantErr:     true,
			expectedErr: errors.ErrorRepositoryUserAlreadyExsist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			user, err := repo.AddNewUser(tt.email, "password_hash", "user", "refresh_token")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
