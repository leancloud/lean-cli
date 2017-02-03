RESOURCES=$(wildcard resources/*)

resources.go: $(RESOURCES)
	./generate_resource.py | gofmt > resources.go
