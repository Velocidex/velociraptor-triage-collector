all:
	go build -o velotriage ./cmd/

compile: kapefiles uac

kapefiles: config/Windows.Triage.Targets.yaml
	go run ./cmd compile -v --config config/Windows.Triage.Targets.yaml

uac: config/Linux.Triage.UAC.yaml
	go run ./cmd compile -v --config config/Linux.Triage.UAC.yaml

test:
	go test -v ./tests -test.count 1

golden:
	cd tests && ./velociraptor.bin --definitions ../output -v --config test.config.yaml golden ./testcases
