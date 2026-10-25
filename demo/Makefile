include demo.env
export

.PHONY: all deps reset cli wait ui stitch

all: reset cli wait ui stitch

deps:
	npm install
	npx playwright install --with-deps chromium

reset:
	./reset.sh

# render the tape from its template, then record the terminal scene.
# NOTE: this leaves AirCompose running (the UI scene needs it).
cli:
	mkdir -p out
	envsubst '$${DEMO_BASE}' < cli.tape.tmpl > cli.tape
	vhs cli.tape

wait:
	./wait-for.sh

ui:
	npm run record:ui

stitch:
	./stitch.sh
