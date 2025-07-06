---
title: "OpenAPI in Practice: Go Server + Python Client from Spec"
description: "A practical walkthrough of using OpenAPI to generate cross-language interfaces"
date: 2025-07-04T22:43:09-07:00
draft: false
searchHidden: false
showtoc: true
categories: [API, Go, Python]
---

OpenAPI is a specification for documenting HTTP APIs for both humans and machines to consume.  As OpenAPI is a specification,
it is language agnostic. OpenAPI relies on generators for translating the specification.  There's more 
than just documentation that's generated. Generators also create language-specific interfaces, tooling, and contracts.  In some 
ways the OpenAPI pattern reminds me of either protobuf with gRPC or ORM schema-first design.  As a result, a declarative API is 
created by the tooling.  

By the end of this post you'll have:

* A working Go http server generated from an OpenAPI specification.
* A Python http client generated from the same specification and authenticates with basic auth.
* Insight into common OpenAPI pitfalls and how to avoid them.

```ascii
[openapi.yaml]
     ↓
+--------------+
| oapi-codegen | ---> [Go Server]
+--------------+
     ↓
+-----------------------+
| openapi-python-client | ---> [Python Client]
+-----------------------+
```

If you would like to follow along, a complete code example can be [**downloaded**](./assets/hello_openapi.tar.gz) and extracted 
into a temporary working directory.

## Generators

Because generators are consuming the specification, the OpenAPI version is determined by what the generators support.

For example, a popular Go generator is [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) and supports 
OpenAPI 3.0.  Where a popular Python generator named
[openapi-python-client](https://github.com/openapi-generators/openapi-python-client) can support both OpenAPI 3.0 and 3.1 specifications.

Generators can be downloaded and managed as part of the languages tooling.  For Go, the oapi-codegen generator is managed with Go 
modules and invoked with `go tool oapi-codegen`.  With Python, creating a virtual environment, using 
pip install openapi-python-client, and pip freeze > requirements.txt will work nicely.

## OpenAPI Schema

At first it wasn't clear to me on how to get started with OpenAPI or what the benefits were.  This is even after reviewing the 
OpenAPI [schema documentation](https://spec.openapis.org/oas/v3.0.3.html) for 3.0.3.  

To get started one needs to create a specification.  A very minimal specification meeting the 3.0.x requirements is listed below. 
It's not a very interesting example as endpoints in the application server aren't defined, but it shows how minimal a 
specification can be that meets schema requirements.

```yaml
openapi: "3.0.3"
info:
  version: 1.0.0
  title: My Contrived Server
paths:
```

Let's get started by extending the simple example defining a path named /status. It will return a 200 response code with a JSON resonse.

```yaml
paths:
  /status:
    get:
      responses:
        '200':
          description: Get status of the application server
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/status'
```

The JSON response is documented in a separate YAML block named components. It defines the response containing a JSON
map containing the keys "state" and "message", both of which have a string value.

```yaml
components: 
    schemas:
      status:
        type: object
        properties:
          state:
            type: string
            example: "GOOD"
          message:
            type: string
            example: "App running within parameters"
```

OpenAPI supports tags, which let you group related endpoints. This example creates a data grouping and puts create_bucket in the group.

```yaml
tags:
  - name: data
    description:  data manipulation endpoints

paths:
    /create_bucket:
      post:
        tags:
            - data
        requestBody:
            required: true
            content:
            application/json:
                schema:
                $ref: '#/components/schemas/create_bucket'
        responses:
            '200':
            description: Create a storage object

```

The OpenAPI specification also provides a definition for authentication to the web application.

```yaml
components:
    securitySchemes:
      basicAuth:
        type: http
        scheme: basic
        description: Endpoints protected by basic auth base64 encoded credentials.

paths:
    /status:
        get:
            security:
                - basicAuth: []
        responses:
            '200':
            description: Get status of the application server
            content:
                application/json:
                schema:
                    $ref: '#/components/schemas/status'
```

Earlier I mentioned the generators will create interface files. Declarations which are considered middleware like 
authentication or logging are out of scope for OpenAPI.
In this example, the security entries are there to document that the endpoints require basic authentication.

## Generate Server Interfaces (Go)

The server walkthrough presumes one has both Make and Go installed, and the [example code](./assets/hello_openapi.tar.gz) (tar.gz file) 
has been downloaded and extracted into a temp/work directory. 

* Download the Go dependencies, including oapi-codegen, by running `make tidy`.
* Generate the server interfaces by running `make server-codegen`, which calls `go tool oapi-codegen`.

Feel free to inspect the api/http.gen.go file before proceeding. You'll see it contains an interface named ServerInterface,
which has the GetStatus or PostStatus endpoints from the OpenAPI specification.  http.gen.go also contains a struct named Status
that was defined from components -> schema -> status.

```go
type Status struct {
	Message string `json:"message"`
	State   string `json:"state"`
}
```

To see the working application server, run `make server-run`.  

The server has Basic Auth enabled with hardcoded credentials. The user is "alice" and the password "mySecretPW".  Curl can be
used to see the response.

```shell
% curl --basic -u alice:mySecretPW  http://localhost:8080/status
{"message":"Initializing","state":"Unknown"}
```

## Generate Client Interfaces (Python)

This is where OpenAPI really shines.  I was able to use a generator to create Python libraries
to be used by the client implementation code.  The walkthrough presumes a recent version of Python3 and pip are installed.

First, create a virtual environment and install the openapi-python-client dependencies.  This shell snippet 
presumes the current working directory is already hello_openapi.

```shell
% python3 -mvenv $PWD/.venv
% source $PWD/.venv/bin/activate
% pip install -r requirements.txt
```

Then run `make client-codegen` to build the Python client libraries located in cmd/client/my_contrived_server.

Generating the client was easy, but figuring out how to pass authentication took some trial and error. I eventually 
realized that the `token` is just a base64-encoded `username:password` string, and the `prefix` should be set to `Basic`.

```
client = AuthenticatedClient(
    base_url="http://127.0.0.1:8080",
    headers={"Content-Type": "application/json", "Accept": "application/json"},
    token="YWxpY2U6bXlTZWNyZXRQVw==",  # Token string is a base64 string containing alice:mySecretPW
    prefix="Basic"
)
```

To see the client in action, run `make client-run`.  Also take a look at cmd/client/client.py.  It 
only took a few lines of python code to implement what the openapi-python-client generator had created.

## Gotchas & Lessons Learned

One issue I have with OpenAPI is the illusion of simplicty. When I first started working with OpenAPI, I noticed the Status struct
had keys referencing a pointer of strings which wasn't ideal.

```go
type Status struct {
	Message *string `json:"message"`
	State   *string `json:"state"`
}
```

It took some fiddling with the OpenAPI specification to make the generator use strings instead of pointers to strings.
Adding 'required' to the schema made the generator do what I wanted.

```yaml
components:
      status:
        type: object
        properties:
          state:
            type: string
            example: "GOOD"
          message:
            type: string
            example: "App within parameters"
        required:
          - state
          - message
```

Another issue was not knowing that in Paths, GETs should have a **responses** entry and POSTS should have a **RequestBody** entry.
It makes sense, but it wasn't obvious to me when stumbling through hello-world.

The main takeaway? Always inspect the generated code. If something doesn’t look right, like unexpected pointers or missing method args,
chances are your spec needs tweaking.


## Wrapping Up

Even though I hit some issues with a fairly simple example, I'm going to continue using OpenAPI specifcations.  Being able to easily generate
client code in a different language was a real win. And let's not forget the free API documentation and contract definitions which comes with OpenAPI. 
I have a more complex OpenAPI project coming up. I'm sure I'll have more notes (and probably more gotchas) to share.  Stay tuned.

If you've had similar struggles with OpenAPI or tips for improving schema design, I’d love to hear them on [Bluesky Social](https://bsky.app/profile/af9.us). 
