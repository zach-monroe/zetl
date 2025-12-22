
tailwind:
	npx @tailwindcss/cli --config ../client/src/tailwind.config.js -i ../client/src/input.css -o ../client/css/style.css --watch

air:
	air

dev:  # <- This is your missing target
	npx concurrently "make tailwind" "make air"
