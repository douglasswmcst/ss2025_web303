package grpc

import (
	"context"
	"order-service/database"
	"order-service/models"
	"testing"
	"time"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockUserServiceClient is a mock for UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) CreateUser(ctx context.Context, req *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userv1.CreateUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, req *userv1.GetUserRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userv1.GetUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetUsers(ctx context.Context, req *userv1.GetUsersRequest, opts ...grpc.CallOption) (*userv1.GetUsersResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userv1.GetUsersResponse), args.Error(1)
}

// MockMenuServiceClient is a mock for MenuServiceClient
type MockMenuServiceClient struct {
	mock.Mock
}

func (m *MockMenuServiceClient) GetMenuItem(ctx context.Context, req *menuv1.GetMenuItemRequest, opts ...grpc.CallOption) (*menuv1.GetMenuItemResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menuv1.GetMenuItemResponse), args.Error(1)
}

func (m *MockMenuServiceClient) GetMenu(ctx context.Context, req *menuv1.GetMenuRequest, opts ...grpc.CallOption) (*menuv1.GetMenuResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menuv1.GetMenuResponse), args.Error(1)
}

func (m *MockMenuServiceClient) CreateMenuItem(ctx context.Context, req *menuv1.CreateMenuItemRequest, opts ...grpc.CallOption) (*menuv1.CreateMenuItemResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*menuv1.CreateMenuItemResponse), args.Error(1)
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	// Auto-migrate the Order and OrderItem models
	err = db.AutoMigrate(&models.Order{}, &models.OrderItem{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()
}

func TestCreateOrder_Success(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 1}).
		Return(&userv1.GetUserResponse{
			User: &userv1.User{Id: 1, Name: "Test User", Email: "test@example.com"},
		}, nil)

	// Mock menu item lookup
	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 1}).
		Return(&menuv1.GetMenuItemResponse{
			MenuItem: &menuv1.MenuItem{Id: 1, Name: "Coffee", Price: 2.50},
		}, nil)

	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 2}).
		Return(&menuv1.GetMenuItemResponse{
			MenuItem: &menuv1.MenuItem{Id: 2, Name: "Tea", Price: 2.00},
		}, nil)

	// Test
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 1,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 1, Quantity: 2},
			{MenuItemId: 2, Quantity: 1},
		},
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotZero(t, resp.Order.Id)
	assert.Equal(t, uint32(1), resp.Order.UserId)
	assert.Equal(t, "pending", resp.Order.Status)
	assert.Len(t, resp.Order.OrderItems, 2)

	// Verify first item
	assert.Equal(t, uint32(1), resp.Order.OrderItems[0].MenuItemId)
	assert.Equal(t, int32(2), resp.Order.OrderItems[0].Quantity)
	assert.InDelta(t, 2.50, resp.Order.OrderItems[0].Price, 0.001)

	// Verify second item
	assert.Equal(t, uint32(2), resp.Order.OrderItems[1].MenuItemId)
	assert.Equal(t, int32(1), resp.Order.OrderItems[1].Quantity)
	assert.InDelta(t, 2.00, resp.Order.OrderItems[1].Price, 0.001)

	mockUserClient.AssertExpectations(t)
	mockMenuClient.AssertExpectations(t)
}

func TestCreateOrder_InvalidUser(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation failure
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 999}).
		Return(nil, status.Errorf(codes.NotFound, "user not found"))

	// Test
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 999,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 1, Quantity: 1},
		},
	})

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "user not found")

	mockUserClient.AssertExpectations(t)
}

func TestCreateOrder_InvalidMenuItem(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation success
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 1}).
		Return(&userv1.GetUserResponse{
			User: &userv1.User{Id: 1, Name: "Test User"},
		}, nil)

	// Mock menu item lookup failure
	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 999}).
		Return(nil, status.Errorf(codes.NotFound, "menu item not found"))

	// Test
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 1,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 999, Quantity: 1},
		},
	})

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "menu item 999 not found")

	mockUserClient.AssertExpectations(t)
	mockMenuClient.AssertExpectations(t)
}

func TestGetOrder(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Create a test order
	testOrder := models.Order{
		UserID: 1,
		Status: "pending",
		OrderItems: []models.OrderItem{
			{MenuItemID: 1, Quantity: 2, Price: 2.50},
		},
	}
	err := db.Create(&testOrder).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		orderID     uint32
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:    "get existing order",
			orderID: uint32(testOrder.ID),
			wantErr: false,
		},
		{
			name:        "get non-existent order",
			orderID:     9999,
			wantErr:     true,
			expectedErr: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := server.GetOrder(ctx, &orderv1.GetOrderRequest{
				Id: tt.orderID,
			})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedErr, st.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.orderID, resp.Order.Id)
				assert.Equal(t, uint32(testOrder.UserID), resp.Order.UserId)
				assert.Equal(t, testOrder.Status, resp.Order.Status)
				assert.Len(t, resp.Order.OrderItems, 1)
			}
		})
	}
}

func TestGetOrders(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Test empty orders
	t.Run("empty orders", func(t *testing.T) {
		ctx := context.Background()
		resp, err := server.GetOrders(ctx, &orderv1.GetOrdersRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Orders)
	})

	// Create multiple test orders
	testOrders := []models.Order{
		{
			UserID: 1,
			Status: "pending",
			OrderItems: []models.OrderItem{
				{MenuItemID: 1, Quantity: 2, Price: 2.50},
			},
		},
		{
			UserID: 2,
			Status: "completed",
			OrderItems: []models.OrderItem{
				{MenuItemID: 2, Quantity: 1, Price: 3.00},
			},
		},
	}

	for _, order := range testOrders {
		err := db.Create(&order).Error
		require.NoError(t, err)
	}

	// Test multiple orders
	t.Run("multiple orders", func(t *testing.T) {
		ctx := context.Background()
		resp, err := server.GetOrders(ctx, &orderv1.GetOrdersRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Orders, 2)

		// Verify orders are returned
		for i, order := range resp.Orders {
			assert.Equal(t, uint32(testOrders[i].UserID), order.UserId)
			assert.Equal(t, testOrders[i].Status, order.Status)
			assert.Len(t, order.OrderItems, 1)
		}
	})
}

func TestModelToProto(t *testing.T) {
	now := time.Now()
	order := &models.Order{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		UserID: 5,
		Status: "pending",
		OrderItems: []models.OrderItem{
			{
				Model: gorm.Model{
					ID:        10,
					CreatedAt: now,
					UpdatedAt: now,
				},
				OrderID:    1,
				MenuItemID: 2,
				Quantity:   3,
				Price:      4.50,
			},
		},
	}

	protoOrder := modelToProto(order)

	assert.Equal(t, uint32(1), protoOrder.Id)
	assert.Equal(t, uint32(5), protoOrder.UserId)
	assert.Equal(t, "pending", protoOrder.Status)
	assert.Equal(t, now.Format(time.RFC3339), protoOrder.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), protoOrder.UpdatedAt)

	// Verify order items
	require.Len(t, protoOrder.OrderItems, 1)
	assert.Equal(t, uint32(10), protoOrder.OrderItems[0].Id)
	assert.Equal(t, uint32(1), protoOrder.OrderItems[0].OrderId)
	assert.Equal(t, uint32(2), protoOrder.OrderItems[0].MenuItemId)
	assert.Equal(t, int32(3), protoOrder.OrderItems[0].Quantity)
	assert.InDelta(t, 4.50, protoOrder.OrderItems[0].Price, 0.001)
}

func TestCreateOrder_PriceSnapshot(t *testing.T) {
	// This test verifies that prices are snapshotted at order creation time
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	database.DB = db

	mockUserClient := new(MockUserServiceClient)
	mockMenuClient := new(MockMenuServiceClient)

	server := &OrderServer{
		UserClient: mockUserClient,
		MenuClient: mockMenuClient,
	}

	// Mock user validation
	mockUserClient.On("GetUser", mock.Anything, &userv1.GetUserRequest{Id: 1}).
		Return(&userv1.GetUserResponse{
			User: &userv1.User{Id: 1, Name: "Test User"},
		}, nil)

	// Mock menu item with specific price
	originalPrice := 5.99
	mockMenuClient.On("GetMenuItem", mock.Anything, &menuv1.GetMenuItemRequest{Id: 1}).
		Return(&menuv1.GetMenuItemResponse{
			MenuItem: &menuv1.MenuItem{Id: 1, Name: "Special", Price: originalPrice},
		}, nil)

	// Create order
	ctx := context.Background()
	resp, err := server.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: 1,
		Items: []*orderv1.OrderItemRequest{
			{MenuItemId: 1, Quantity: 1},
		},
	})

	require.NoError(t, err)
	assert.InDelta(t, originalPrice, resp.Order.OrderItems[0].Price, 0.001)

	// Verify price is stored in database
	var dbOrder models.Order
	err = db.Preload("OrderItems").First(&dbOrder, resp.Order.Id).Error
	require.NoError(t, err)
	assert.InDelta(t, originalPrice, dbOrder.OrderItems[0].Price, 0.001)

	mockUserClient.AssertExpectations(t)
	mockMenuClient.AssertExpectations(t)
}
