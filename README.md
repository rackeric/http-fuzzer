# http-fuzzer

A simple API server that accepts jobs to execute http fuzzing on a target. A UI is also provided to interact with the API to upload worklists, submit jobs, and see status of jobs.

## Quickstart

To start the API server:
```
make run
```

Then, open a web browser to `http://localhost:8080/`.

More options are found with:
```
$ make help

Usage:
  make <target>

Targets:
  build                Build the application
  build-unix           Build for unix
  run                  Run the application
  clean                Clean build directory
  test                 Run tests
  coverage             Run tests with coverage
  deps                 Download dependencies
  install-lint         Install golangci-lint
  lint                 Run linter
  vet                  Run go vet
  fmt                  Format code
  help                 Show help
```