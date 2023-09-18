module github.com/vasiliiperfilev/ddd/internal/users

go 1.14

require (
	cloud.google.com/go/firestore v1.2.0
	firebase.google.com/go v3.12.0+incompatible
	github.com/vasiliiperfilev/ddd/internal/common v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi v4.1.0+incompatible
	github.com/go-chi/render v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.5.0
	golang.org/x/crypto v0.0.0-20191112222119-e1110fd1c708 // indirect
	google.golang.org/api v0.21.0
	google.golang.org/grpc v1.28.0
	gopkg.in/yaml.v2 v2.2.5 // indirect
)

replace github.com/vasiliiperfilev/ddd/internal/common => ../common/
