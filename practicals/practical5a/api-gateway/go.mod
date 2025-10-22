module api-gateway

go 1.23

require (
	github.com/douglasswm/student-cafe-protos v0.0.0
	github.com/go-chi/chi/v5 v5.0.11
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
)

replace github.com/douglasswm/student-cafe-protos => ../student-cafe-protos
