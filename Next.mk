# Copyright 2021 Tetrate
# Licensed under the Apache License, Version 2.0 (the "License")
.PHONY: test

# ANSI escape codes. See https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
GREEN  = $(shell printf "\33[32m")
WHITE  = $(shell printf "\33[37m")
YELLOW = $(shell printf "\33[33m")
RESET  = $(shell printf "\33[0m")

# This formats help statements using ANSI codes. To hide a target from help, don't comment it with a trailing '##'.
help: ## Describe how to use each target
	@echo "$(WHITE)targets:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "$(YELLOW)%-20s$(RESET) $(GREEN)%s$(RESET)\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all unit tests
	@echo "${YELLOW}Running unit tests${RESET}"
	@go test . ./internal/...
	@echo "${GREEN}✔ successfully ran unit tests${RESET}\n"
