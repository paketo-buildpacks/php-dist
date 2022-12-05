.PHONY: test retrieve

id:
	@echo php

test:
	@cd test; \
	./test.sh --version $(version) --tarballPath $(tarballPath)

retrieve:
	@cd retrieval; \
	go run . \
		--buildpack_toml_path=$(buildpackTomlPath) \
		--output=$(output)
