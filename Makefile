all: compile

ASSET_DIRS = $(shell find ./config/ -type d)
ASSET_FILES = $(shell find ./config/ -type f -name '*')

artifacts := \
	output/Windows.Triage.Targets.yaml \
	output/Windows.KapeFiles.Targets.yaml \
	output/Linux.Triage.UAC.yaml

templates := \
	config/Windows.KapeFiles.Targets.yaml \
	templates/Windows.Triage.Targets.yaml \
	templates/CommonExports.yaml \
	templates/CommonSources.yaml \
	templates/Linux.Triage.UAC.yaml

build:
	go build -o velotriage ./cmd/

compile: $(artifacts)
	cd output && rm -f Velociraptor_Triage_v0.1.zip && zip Velociraptor_Triage_v0.1.zip *.yaml

output/Windows.KapeFiles.Targets.yaml: $(templates) \
	config/Windows.KapeFiles.Targets.yaml
	go run ./cmd compile -v --config config/Windows.KapeFiles.Targets.yaml

output/Linux.Triage.UAC.yaml: $(templates) \
	config/Linux.Triage.UAC.yaml
	go run ./cmd compile -v --config config/Linux.Triage.UAC.yaml

output/Windows.Triage.Targets.yaml: $(templates) \
    $(ASSET_FILES) $(ASSET_DIRS) \
	config/Windows.Triage.Targets.yaml
	go run ./cmd compile -v --config config/Windows.Triage.Targets.yaml

.PHONY: clean
clean:
	rm output/*.yaml output/*.zip

test:
	go test -v ./tests -test.count 1

golden:
	cd tests && X=testEnv ./velociraptor.bin --definitions ../output -v --config test.config.yaml golden ./testcases --filter=${GOLDEN}

verify: compile ./tests/velociraptor.bin
	./tests/velociraptor.bin artifacts verify ./output/*.yaml --builtin

./tests/velociraptor.bin:
	echo "Downloading the Velociraptor binary"
	curl -o ./tests/velociraptor.bin -L https://github.com/Velocidex/velociraptor/releases/download/v0.75/velociraptor-v0.75.2-linux-amd64-musl
	chmod +x ./tests/velociraptor.bin

uac:
	go test -v ./tests -test.count 1
