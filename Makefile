all:
	go build -o velotriage ./cmd/

compile: kapefiles

kapefiles:
	go run ./cmd compile -v --config config/Windows.KapeFiles.Targets.yaml
