[doc("Running golang application with vendoring")]
@run:
    go run -mod=vendor cmd/app/main.go
[doc("Counting lines of code in git")]
@lines:
    cloc --vcs=git .
@fmt:
    golangci-lint fmt