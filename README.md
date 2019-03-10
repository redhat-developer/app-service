# A boilerplate HTTP-REST server to be built upon

This project provides the code to build a HTTP-REST server with some good defaults and features already configfured. Among those are:

 * a /status/yaml and /status/json endpoint for health-checking the server
 * graceful shutdown of HTTP server
 * Apache combined logging output
 * porwerful configuration with defaults
 * support for golden files to capture HTTP requests and responses
 * an easy to use Makefile

# To build and run the server...

```bash
# Clone the project
git clone https://github.com/redhat-developer/boilerplate-app ${GOPATH}/src/github.com/redhat-developer/boilerplate-app
# Navigate to the project
cd ${GOPATH}/src/github.com/redhat-developer/boilerplate-app
# Check that you have all the tools needed (git, go, dep, docker)
make check-tools
# Build the server
make all
# Run tests including coverage anaylsis
make test-coverage
# Run the server and confiugre an HTTP address to listen on
APP_HTTP_ADDRESS=0.0.0.0:8088 ./out/app-server
# Curl the /status/yaml endpoint
curl -v 0.0.0.0:8001/status/yaml
```

# Make targets

Here's a list of `make` targets and their purpose:

**`make check-tools`**: Checks that `make` can find these binaries: `go`, `dep`, `git`, and `docker`. 

**`make all`**: builds the app server and places it in `./out/app-server`.

**`make test`**: runs all tests in all packages without any coverage information. Upon the first error, we don't continue running any other tests.

**`make test-coverage`**: runs all tests and collects coverage information that will be stored in `./out/cover.out`.

**`make test-coverage-html`**: opens a browser to show you the annotated source code with coverage information from `./out/cover.out`. If `./out/cover.out` doesn't exist yet it will be generated.

**`make clean`**: removes all buiöd and tests artifacts from `./out` and removes the `./vendor` directory.

If you want to see what commands are executed before they are being run by `make` you can set `VERBOSE=1`, i.e. `make VERBOSE=1 test-coverage`. 