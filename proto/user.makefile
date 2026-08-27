.PHONY: user-grpc

user-grpc:
	cd proto && protoc \
	--proto_path=. \
	--go_out=out \
	--go_opt=paths=source_relative \
	--go-grpc_out=out \
	--go-grpc_opt=paths=source_relative \
	user/v1/v1.proto

order-grpc:
	cd proto && protoc \
	--proto_path=. \
	--go_out=out \
	--go_opt=paths=source_relative \
	--go-grpc_out=out \
	--go-grpc_opt=paths=source_relative \
	order/v1/order.proto
