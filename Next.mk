# Copyright 2021 Tetrate
# Licensed under the Apache License, Version 2.0 (the "License")
.PHONY: test

# ANSI escape codes. f_ means foreground, b_ background.
# See https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
b_black  = $(shell printf "\33[40m")
f_white  = $(shell printf "\33[97m")
f_gray  = $(shell printf "\33[37m")
f_dark_gray  = $(shell printf "\33[90m")
f_bright_magenta  = $(shell printf "\33[95m")
b_bright_magenta  = $(shell printf "\33[105m")
ansi_reset  = $(shell printf "\33[0m")
ansi_func_e = $(b_black)$(f_white)func-$(b_bright_magenta)e$(ansi_reset)
ansi_format_dark = $(f_gray)$(f_bright_magenta)%-10s$(ansi_reset) $(f_dark_gray)%s$(ansi_reset)\n
ansi_format_bright = $(f_white)$(f_bright_magenta)%-10s$(ansi_reset) $(f_white)$(b_bright_magenta)%s$(ansi_reset)\n

# This formats help statements in ANSI colors. To hide a target from help, don't comment it with a trailing '##'.
help: ## Describe how to use each target
	@printf "$(ansi_func_e)$(f_white)\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "$(ansi_format_dark)", $$1, $$2}' $(MAKEFILE_LIST)


test: ## Run all unit tests
	@printf "$(ansi_format_dark)" test "running unit tests"
	@go test . ./internal/...
	@printf "$(ansi_format_bright)" test "✔"
