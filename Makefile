THEME_DIR := themes/PaperMod
PATCH_DIR := patches/PaperMod

.PHONY: submodule-check submodule-patch submodule-update submodule-reset hugo-serve hugo-publish targets

# List targets within Makefile
targets:
	@echo "targets:          List available targets in this Makefile"
	@echo "submodule-update: Fetches new upstream changes"
	@echo "submodule-patch:  Applies local patches to the submodule"
	@echo "submodule-reset:  Reverts/removes local patches made to the submodule"
	@echo "submodule-check:  Confirms that the submodule does not contain a .github directory"
	@echo "hugo-serve:       Use HUGO to generate HTML for development and serve on localhost"
	@echo "hugo-publish:     Use Hugo to generate HTML for Production"

# Fail if the PaperMod submodule contains a .github directory before patching
submodule-check:
	@echo "Checking for unexpected .github content in $(THEME_DIR)..."
	@if [ -d "$(THEME_DIR)/.github" ]; then \
		echo "ERROR: $(THEME_DIR)/.github exists before patching."; \
		echo "       PaperMod upstream may have added GitHub workflows."; \
		echo "       Review and update your patches before continuing."; \
		exit 1; \
	fi
	@echo "OK: No unexpected .github directory."

# Apply patches from patch directory to PaperMod theme
submodule-patch: submodule-reset
	@echo "Applying patches to $(THEME_DIR)..."
	@cd $(THEME_DIR) && \
	for p in ../../$(PATCH_DIR)/*.patch; do \
		echo "  Applying $$p"; \
		git apply $$p || exit 1; \
	done
	@echo "Patches applied."

# Pull latest upstream theme updates
submodule-update:
	@echo "Updating PaperMod submodule..."
	git submodule update --init --recursive --remote --rebase

# Reset theme submodule if needed.
submodule-reset:
	@echo "Resetting submodule to a clean state..."
	cd $(THEME_DIR) && git reset --hard HEAD

# Run Hugo for local development
hugo-serve: submodule-check
	@echo "Generating HTML for development and serving on localhost"
	go tool hugo serve

# Run Hugo for production
hugo-publish: submodule-check
	@echo "Generating HTML for Production"
	go tool hugo

