tailwind:
	npx @tailwindcss/cli -i ../client/src/input.css -o ../client/css/style.css --watch

tailwind-build:
	npx @tailwindcss/cli -i ../client/src/input.css -o ../client/css/style.css --minify

air:
	air

dev:
	npx concurrently "make tailwind" "make air"
