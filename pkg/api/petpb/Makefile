.PHONY: gen serve-docs

gen:
	protoc -I/usr/local/include -I. \
		-I${GOPATH}/protobuf \
		--go_out . --go_opt paths=source_relative \
		--go-grpc_out . --go-grpc_opt paths=source_relative \
		--grpc-gateway_out . --grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt logtostderr=true \
        --grpc-gateway_opt generate_unbound_methods=true \
        --grpc-gateway_opt register_func_suffix=GW \
        --grpc-gateway_opt allow_delete_body=true \
        --openapiv2_out . --openapiv2_opt logtostderr=true \
		pet.proto

serve-docs:
	docker run -it --rm -p 80:80 \
      -v $$(pwd)/pet.swagger.json:/usr/share/nginx/html/swagger.yaml \
      -e SPEC_URL=swagger.yaml redocly/redoc