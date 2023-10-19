.PHONY: openapi
openapi:
	oapi-codegen -generate types -o internal/trainings/openapi_types.gen.go -package main api/openapi/trainings.yml
	oapi-codegen -generate chi-server -o internal/trainings/openapi_api.gen.go -package main api/openapi/trainings.yml

	oapi-codegen -generate types -o internal/trainer/openapi_types/openapi_types.gen.go -package openapi_types api/openapi/trainer.yml
	oapi-codegen -generate chi-server -o internal/trainer/openapi_types/openapi_api.gen.go -package openapi_types api/openapi/trainer.yml

	oapi-codegen -generate types -o internal/users/oapi/openapi_types.gen.go -package oapi api/openapi/users.yml
	oapi-codegen -generate chi-server -o internal/users/oapi/openapi_api.gen.go -package oapi api/openapi/users.yml

.PHONY: proto
proto:
	protoc --go_out=internal/common/genproto --go-grpc_out=internal/common/genproto -I api/protobuf api/protobuf/trainer.proto
	protoc --go_out=internal/common/genproto --go-grpc_out=internal/common/genproto -I api/protobuf api/protobuf/users.proto

.PHONY: lint
lint:
	@./scripts/lint.sh trainer
	@./scripts/lint.sh trainings
	@./scripts/lint.sh users
