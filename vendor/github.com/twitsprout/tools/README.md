![Go Coverage](https://github.com/twitsprout/tools/wiki/coverage.svg)

# tools

Go libraries for shared code. Currently supports Go version 1.15 up to 1.20.

## Testing

The tests in this repo need a Redis and Postgres container to execute properly. You can stand up these containers by running:

```
make test-prep
```

This will start up a Postgres container on port 5432 and a Redis container on port 6379 so make sure you don't have any other containers or local services occupying these ports.

Once the supporting containers are up, you can run:

```
make test
```

This will run all the tests locally. You can run `make test-clean` to spin down the test containers when you're done with them.

Before committing code, make sure you run:

```
make lint
```

so there aren't any linter errors thrown by CI.

## Releasing

To push out a new release, [draft a new release on GitHub](https://github.com/twitsprout/tools/releases/new) and provide an appropriate description of the changes. 

## Packages

### buffer

Buffer is a small library with a shared buffer pool.

### crypto

Crypto contains functions for reading random (or pseduo-random) bytes, as well 
as functions for encoding/decoding bytes.

### date

Date contains the functions for manipulating datetime to formats including:
     
 * ISO Format    
 * ISO 8601 Format    
 * Go time      
 * UTC Format

### distlock

Distlock contains an implementation of a distributed lock manager. It allows for
the automated locking, extending, and unlocking of a distributed lock.

### http

HTTP contains:

 * Client creation function
 * Reusable HTTP middleware functions
 * JSON reading/writing helpers

### lifecycle

Lifecycle contains a type that manages the execution of long-running processes.
It will keep track of all started processing, recording the first one to exit,
signal to all other processes that they should shut down, and wait until all
processes have gracefully exited (or a timeout is reached).

### postgres

Postgres contains helpers to create a database instance, as well as functions
to create placeholder queries, and an implementation of a pubsub client.

### requestid

RequestID contains helper functions to create request ID strings, and saving
them inside a context.

### slack

Slack contains a client to send messages to Slack via their API.

### zap

Zap contains a logging implementation that outputs JSON objects to the provided
io.Writer.

### maintenance

Maintenance contains a handler that can be used to decommission a specific route
if a service is down for maintenance and unavailable to serve requests.
