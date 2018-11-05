# go-rest-template
Standard GO Micro-service template providing a ReSTful interface to be used inside a Docker container.  Also provides a ready and live probe interface for container orchestration and service discovery tools

This template provides a consistent interface for a ReSTful API that when used with an API contract first driven approach usig a tool such as API Builder ( https://www.apibuilder.io/ ) is very powerful.

N.B. This template is only intended for HTTP-GET which is usually (but not always) synchronous, whilst PUT and POST are asynchronous and involve some queueing mechanism, utilising the HTTP - 202 Accepted status with a location header.
