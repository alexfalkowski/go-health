# Accepted design for testing

Part of the [accepted-design index](../accepted-design.md).

## Live network integration tests

Server tests intentionally exercise real HTTP, TCP, DNS, and the default
online checker. Preserve that integration coverage when reviewing the use of
`google.com`, the default connectivity URLs, or other outbound requests.
Revisit live network use when the task calls for removing that coverage or
when a concrete defect affects the test or the behavior it exercises.

[Server integration tests](../../../server/server_test.go)
