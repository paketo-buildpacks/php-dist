.PHONY: retrieve

retrieve:
	@cd retrieval; \
	go test -v ./... && cd run && BP_LOG_LEVEL=debug go run . \
		--buildpack_toml_path=$(buildpackTomlPath) \
		--output=$(output)
