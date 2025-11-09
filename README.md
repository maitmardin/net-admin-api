# Network Administration API

A sketch of a RESTful network administration API implemented in Go.

## Building and Running

With Docker:
```
docker compose up --build
```

With Go (requires go 1.23+):
```
go run cmd/api/main.go
```

This starts the API server at its default location [http://localhost:8080](http://localhost:8080).

There are two environment variables that can be configured:

- `PORT` - the port that the API server should listen on (default `8080`)
- `VLAN_STORE_PATH` - the path to a json file where VLANs are stored (default `vlans.json`). Directories are not automatically created.

## Test Strategy

Since there are no external dependencies, it is easy to set up integration tests that cover most of the functionality
without unit tests and mocking. A success scenario that covers all of the `api/v1/vlans` endpoints is performed as a
single test in `internal/server/handle_vlans_test.go`. And then there are separate tests to cover the error scenarios
of each endpoint.

## Misc Comments and Open Questions

- Used UUID as the identifier for VLANs to guarantee uniqueness. Could VLAN ID (VID) be used as a unique identifier
  instead? Having two identifiers, ID and VID, is a bit confusing.
- VLAN status is currently a random string, managed by the API user. Could this be an enum, perhaps also read-only and
  determined by the service itself?
- `PUT /api/v1/vlans/{id}` endpoint could be improved by not having the ID in URL. It is duplicating the ID in request
  body and is a source for errors. Alternatively, a PATCH method could be used without having the ID in the body.
- Kubernetes deployment assumes a stateless app, which it is not. An actual deployment would be more complex.
- Added coverage.html manually to the repo. Would be better to publish it somewhere (like GitHub Pages) as a CI step.
