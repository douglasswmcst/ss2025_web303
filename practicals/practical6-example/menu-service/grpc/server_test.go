package grpc

import (
	"context"
	"menu-service/database"
	"menu-service/models"
	"testing"
	"time"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
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

	// Auto-migrate the MenuItem model
	err = db.AutoMigrate(&models.MenuItem{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()
}

func TestCreateMenuItem(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	server := NewMenuServer()

	tests := []struct {
		name    string
		request *menuv1.CreateMenuItemRequest
		wantErr bool
	}{
		{
			name: "successful menu item creation",
			request: &menuv1.CreateMenuItemRequest{
				Name:        "Cappuccino",
				Description: "Espresso with steamed milk and foam",
				Price:       4.50,
			},
			wantErr: false,
		},
		{
			name: "create item with zero price",
			request: &menuv1.CreateMenuItemRequest{
				Name:        "Water",
				Description: "Free water",
				Price:       0.0,
			},
			wantErr: false,
		},
		{
			name: "create item with long description",
			request: &menuv1.CreateMenuItemRequest{
				Name:        "Special Brew",
				Description: "A very long description that describes the coffee in great detail with many words",
				Price:       5.99,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.CreateMenuItem(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotZero(t, resp.MenuItem.Id)
				assert.Equal(t, tt.request.Name, resp.MenuItem.Name)
				assert.Equal(t, tt.request.Description, resp.MenuItem.Description)
				assert.InDelta(t, tt.request.Price, resp.MenuItem.Price, 0.001)
				assert.NotEmpty(t, resp.MenuItem.CreatedAt)
				assert.NotEmpty(t, resp.MenuItem.UpdatedAt)
			}
		})
	}
}

func TestGetMenuItem(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	server := NewMenuServer()

	// Create a test menu item
	testItem := models.MenuItem{
		Name:        "Latte",
		Description: "Espresso with steamed milk",
		Price:       4.00,
	}
	err := db.Create(&testItem).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		itemID      uint32
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:    "get existing menu item",
			itemID:  uint32(testItem.ID),
			wantErr: false,
		},
		{
			name:        "get non-existent menu item",
			itemID:      9999,
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
		{
			name:        "get item with ID 0",
			itemID:      0,
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.GetMenuItem(ctx, &menuv1.GetMenuItemRequest{
				Id: tt.itemID,
			})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.itemID, resp.MenuItem.Id)
				assert.Equal(t, testItem.Name, resp.MenuItem.Name)
				assert.Equal(t, testItem.Description, resp.MenuItem.Description)
				assert.InDelta(t, testItem.Price, resp.MenuItem.Price, 0.001)
			}
		})
	}
}

func TestGetMenu(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	server := NewMenuServer()

	// Test empty menu
	t.Run("empty menu", func(t *testing.T) {
		ctx := context.Background()
		resp, err := server.GetMenu(ctx, &menuv1.GetMenuRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.MenuItems)
	})

	// Create multiple test menu items
	testItems := []models.MenuItem{
		{Name: "Coffee", Description: "Black coffee", Price: 2.50},
		{Name: "Tea", Description: "Green tea", Price: 2.00},
		{Name: "Sandwich", Description: "Ham and cheese", Price: 5.50},
	}

	for _, item := range testItems {
		err := db.Create(&item).Error
		require.NoError(t, err)
	}

	// Test multiple items
	t.Run("multiple items", func(t *testing.T) {
		ctx := context.Background()
		resp, err := server.GetMenu(ctx, &menuv1.GetMenuRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.MenuItems, 3)

		// Verify all items are returned
		for i, item := range resp.MenuItems {
			assert.Equal(t, testItems[i].Name, item.Name)
			assert.Equal(t, testItems[i].Description, item.Description)
			assert.InDelta(t, testItems[i].Price, item.Price, 0.001)
		}
	})
}

func TestModelToProto(t *testing.T) {
	now := time.Now()
	item := &models.MenuItem{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        "Test Item",
		Description: "Test Description",
		Price:       3.99,
	}

	protoItem := modelToProto(item)

	assert.Equal(t, uint32(1), protoItem.Id)
	assert.Equal(t, "Test Item", protoItem.Name)
	assert.Equal(t, "Test Description", protoItem.Description)
	assert.InDelta(t, 3.99, protoItem.Price, 0.001)
	assert.Equal(t, now.Format(time.RFC3339), protoItem.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), protoItem.UpdatedAt)
}

func TestPriceHandling(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	server := NewMenuServer()

	// Test various price formats
	testCases := []struct {
		name  string
		price float64
	}{
		{"integer price", 5.0},
		{"two decimal places", 5.99},
		{"three decimal places", 5.999},
		{"very small price", 0.01},
		{"large price", 999.99},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.CreateMenuItem(ctx, &menuv1.CreateMenuItemRequest{
				Name:        "Test Item",
				Description: "Price test",
				Price:       tc.price,
			})

			require.NoError(t, err)
			assert.InDelta(t, tc.price, resp.MenuItem.Price, 0.001)
		})
	}
}
