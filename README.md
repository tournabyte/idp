# Package `github.com/tournabyte/idp` - Identity plugin for the Tournabyte platform

## Overview

This Go module is designed to handle the identity needs of the Tournabyte platform. It provides a plugin-based interface to be used with Go HTTP servers and routers. 

## Getting Started *(coming soon)*

### Building the server

This server package can be build using standard Go tooling. The entry point of the package is defined in `app/main.go` Specify this source file for the build command as follows

```bash
$ go build app/main.go
```

Use the `-o` flag to specify where the resulting binary should be placed

```bash
$ go build -o bin/idp app/main.go
```

The  `idp` executable will be place in the `bin` directory of the current working directory. It will be created if it does not exist

### Starting the service

The service can be started by launching the resulting executable. The application's CLI provides one subcommand: `serve`. Invoke the subcommand explicitly to start the service. You can use the `-p` to specify the port the service will listen for requests on.

### Resources

This go http server application offers the following endpoints for clients

#### `POST /acounts`

This exposes an endpoint for the `accounts` resources to be created. Along with the request, the following body structure is expected

```json
{"email": "testuser@example.com"}
```

The endpoint will respond with the resulting created resource upon success:

```json
{
  "Id": "69165d0e27087f8ed0d2275b",
  "Email": "testuser@example.io"
}
```

The endpoint will respond with an error message if any of the following occur

- The expected body was not present
- The body could not be parsed as JSON
- The request to create a resource conflicts with an existing one
- Any upstream errors that may occur

#### GET /accounts/{id}

This exposes an endpoint to search the `accounts` resources by a given identifier. The request path include a parameter which is the hex ID of the account to look for. The endpoint will respond will the found resource upon success:

```json
{
  "Id": "69165d0e27087f8ed0d2275b",
  "Email": "testuser@example.io"
}
```

The endpoint will respond with an error message if any of the following occur

- The ID path parameter was not provided
- The ID path parameter is not a valid hex ID
- The requested resource does not exist
- Any upstream errors that may occur
