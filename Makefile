.PHONY: dev build reset-connections

dev:
	wails dev

build:
	wails build

# Deletes saved connections so .devconnections.json is picked up on next startup.
# Copy .devconnections.json.example to .devconnections.json and fill in your credentials first.
reset-connections:
	@target="$$HOME/Library/Application Support/QueryRaccoon/connections.json"; \
	if [ -f "$$target" ]; then \
		rm "$$target" && echo "Connections reset. Run 'make dev' to reseed from .devconnections.json."; \
	else \
		echo "No saved connections found."; \
	fi
