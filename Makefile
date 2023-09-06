.DEFAULT_GOAL := default

app_name = console

default:
	go build -o $(app_name) cmd/*.go

