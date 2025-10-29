module integration-tests

go 1.23

require (
	github.com/douglasswm/student-cafe-protos v0.0.0
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.59.0
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.5
)

replace github.com/douglasswm/student-cafe-protos => ../../student-cafe-protos

replace menu-service => ../../menu-service

replace order-service => ../../order-service

replace user-service => ../../user-service
