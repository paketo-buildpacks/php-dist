.PHONY: test retrieve

id:
	@echo php

test:
	@cd test; \
	./test.sh --version $(version) --tarballPath $(tarballPath) --os $(os) --arch $(arch)

retrieve:
	@cd retrieval; \
	PHP_RELEASES_DIR="$(abspath ../php-releases)" \
	PHP_RELEASES_REFRESH=true \
	go run . \
		--buildpack_toml_path=$(buildpackTomlPath) \
		--output=$(output)
