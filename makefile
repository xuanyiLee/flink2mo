all: buildcheck gen_data

buildcheck:
	cd ./cmd/check && go build && mv check ../../

gen_data:
	cd ./cmd/gen_data && go build && mv gen_data ../../

grant_priv:
	chmod 0755 gen_data check