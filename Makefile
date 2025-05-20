run:
	air -c .air.toml

templ:
	templ generate --watch --proxy="http://localhost:5173" --open-browser=false -v

css:
	tailwindcss -i ./assets/css/tailwind.css -o ./assets/css/styles.css --watch
