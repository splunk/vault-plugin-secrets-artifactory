# Backlog

Some implementations haven't been fully completed in order to make initial plugin simple.

## Cached Client

COMPLETED: ~~Currently each rquest spins up a new artifactory client. We believe this will cause issues soon.~~

## Rollback

Utilize WAL(Write-Ahead Log) to rollback in case of Artifactory API failure.

## Configurable Client Timeout

Client timeout should be varied by organization/artifactory instance. Sometime we want to fail fast. Sometime, we want to wait for queue processing.

## HTTP Client Replacement

jfrog-client-go heavily wrapps http client and doesn't leave us wiggle room to tweak. Whether we'll replace only http client or entire library and create our own is to be determined.
