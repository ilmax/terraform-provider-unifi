TEST         ?= ./...
TESTARGS     ?=
TEST_COUNT   ?= 1
TEST_TIMEOUT ?= 10m

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test -count $(TEST_COUNT) $(TEST) $(TESTARGS)

.PHONY: testacc
testacc:
	TF_ACC=1 go test -count $(TEST_COUNT) -timeout $(TEST_TIMEOUT) -run TestAcc -v $(TEST) $(TESTARGS)
