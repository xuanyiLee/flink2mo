all: buildcheck gen_data

buildcheck:
	@echo "Building check..."
	@cd ./cmd/check && go build && mv check ../../

gen_data:
	@echo "Building gen_data..."
	@cd ./cmd/gen_data && go build && mv gen_data ../../

grant_priv:
	@chmod 0755 gen_data check

clean:
	rm -rf gen_data check