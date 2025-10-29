package grpc

import (
	"context"
	"testing"
	"time"
	"user-service/database"
	"user-service/models"

	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	// Auto-migrate the User model
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()
}

func TestCreateUser(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	server := NewUserServer()

	tests := []struct {
		name        string
		request     *userv1.CreateUserRequest
		wantErr     bool
		expectedMsg string
	}{
		{
			name: "successful user creation",
			request: &userv1.CreateUserRequest{
				Name:        "John Doe",
				Email:       "john@example.com",
				IsCafeOwner: false,
			},
			wantErr: false,
		},
		{
			name: "create cafe owner",
			request: &userv1.CreateUserRequest{
				Name:        "Jane Owner",
				Email:       "jane@cafeshop.com",
				IsCafeOwner: true,
			},
			wantErr: false,
		},
		{
			name: "empty name should still work (validation is optional)",
			request: &userv1.CreateUserRequest{
				Name:        "",
				Email:       "test@example.com",
				IsCafeOwner: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.CreateUser(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotZero(t, resp.User.Id)
				assert.Equal(t, tt.request.Name, resp.User.Name)
				assert.Equal(t, tt.request.Email, resp.User.Email)
				assert.Equal(t, tt.request.IsCafeOwner, resp.User.IsCafeOwner)
				assert.NotEmpty(t, resp.User.CreatedAt)
				assert.NotEmpty(t, resp.User.UpdatedAt)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	server := NewUserServer()

	// Create a test user
	testUser := models.User{
		Name:        "Test User",
		Email:       "test@example.com",
		IsCafeOwner: false,
	}
	err := db.Create(&testUser).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uint32
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:    "get existing user",
			userID:  uint32(testUser.ID),
			wantErr: false,
		},
		{
			name:        "get non-existent user",
			userID:      9999,
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
		{
			name:        "get user with ID 0",
			userID:      0,
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.GetUser(ctx, &userv1.GetUserRequest{
				Id: tt.userID,
			})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.userID, resp.User.Id)
				assert.Equal(t, testUser.Name, resp.User.Name)
				assert.Equal(t, testUser.Email, resp.User.Email)
			}
		})
	}
}

func TestGetUsers(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	server := NewUserServer()

	// Test empty database
	t.Run("empty database", func(t *testing.T) {
		ctx := context.Background()
		resp, err := server.GetUsers(ctx, &userv1.GetUsersRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Users)
	})

	// Create multiple test users
	testUsers := []models.User{
		{Name: "User 1", Email: "user1@example.com", IsCafeOwner: false},
		{Name: "User 2", Email: "user2@example.com", IsCafeOwner: true},
		{Name: "User 3", Email: "user3@example.com", IsCafeOwner: false},
	}

	for _, user := range testUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Test multiple users
	t.Run("multiple users", func(t *testing.T) {
		ctx := context.Background()
		resp, err := server.GetUsers(ctx, &userv1.GetUsersRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Users, 3)

		// Verify all users are returned
		for i, user := range resp.Users {
			assert.Equal(t, testUsers[i].Name, user.Name)
			assert.Equal(t, testUsers[i].Email, user.Email)
			assert.Equal(t, testUsers[i].IsCafeOwner, user.IsCafeOwner)
		}
	})
}

func TestModelToProto(t *testing.T) {
	now := time.Now()
	user := &models.User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        "Test User",
		Email:       "test@example.com",
		IsCafeOwner: true,
	}

	protoUser := modelToProto(user)

	assert.Equal(t, uint32(1), protoUser.Id)
	assert.Equal(t, "Test User", protoUser.Name)
	assert.Equal(t, "test@example.com", protoUser.Email)
	assert.Equal(t, true, protoUser.IsCafeOwner)
	assert.Equal(t, now.Format(time.RFC3339), protoUser.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), protoUser.UpdatedAt)
}
