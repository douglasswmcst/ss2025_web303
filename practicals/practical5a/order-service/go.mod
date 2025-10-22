module order-service

go 1.23

require (
	github.com/douglasswm/student-cafe-protos v0.0.0
	github.com/go-chi/chi/v5 v5.0.11
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)
