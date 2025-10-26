[doc("Running golang application with vendoring")]
@run:
    go run -mod=vendor cmd/app/main.go
[doc("Counting lines of code in git")]
@lines:
    cloc --vcs=git .
@fmt:
    golangci-lint fmt
@nats_init:
    nats stream add --config ./config/nats/append_stream.json --user nats --password nats

    nats stream add --config ./config/nats/debezium_stream.json --user nats --password nats