module github.com/vasiliiperfilev/ddd/internal/trainer

go 1.14

require (
	cloud.google.com/go/firestore v1.12.0
	github.com/deepmap/oapi-codegen v1.3.6
	github.com/go-chi/chi v4.1.0+incompatible
	github.com/go-chi/render v1.0.3
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	github.com/vasiliiperfilev/ddd/internal/common v0.0.0-00010101000000-000000000000
	google.golang.org/api v0.141.0
	google.golang.org/grpc v1.58.1
	google.golang.org/protobuf v1.31.0
)

replace github.com/vasiliiperfilev/ddd/internal/common => ../common/
