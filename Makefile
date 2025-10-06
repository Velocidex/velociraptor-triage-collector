all:
	go build -o velotriage ./cmd/

ASSET_DIRS = $(shell find ./config/ -type d)
ASSET_FILES = $(shell find ./config/ -type f -name '*')

artifacts := \
	output/Windows.Triage.Targets.yaml \
	output/Windows.KapeFiles.Targets.yaml \
	output/Linux.Triage.UAC.yaml

compile: $(artifacts)
	cd output && rm -f Velociraptor_Triage_v0.1.zip && zip Velociraptor_Triage_v0.1.zip *.yaml

output/%.yaml: config/%.yaml templates/%.yaml
	go run ./cmd compile -v --config $<

output/Windows.KapeFiles.Targets.yaml: \
	config/Windows.KapeFiles.Targets.yaml \
	templates/Windows.Triage.Targets.yaml
	go run ./cmd compile -v --config $<

output/Windows.Triage.Targets.yaml: \
	config/Windows.Triage.Targets.yaml \
	templates/Windows.Triage.Targets.yaml \
    $(ASSET_FILES) $(ASSET_DIRS)
	go run ./cmd compile -v --config $<

.PHONY: clean
clean:
	rm output/*.yaml output/*.zip

test:
	go test -v ./tests -test.count 1

golden:
	cd tests && X=testEnv ./velociraptor.bin --definitions ../output -v --config test.config.yaml golden ./testcases --filter=${GOLDEN}

verify: compile
	./tests/velociraptor.bin artifacts verify ./output/*.yaml
