build:
	acorn build -t todo .
	acorn build -t todo-db ./external-db

apply: build
	kubectl apply -Rf ./manifests

dev:
	acorn run --name todo-dev --dev .
