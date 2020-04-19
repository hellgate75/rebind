# Re-Bind Dns Server
Go-Language API Dns multi-plex / tls protected server

# Features

This project contains following components

* [Re-Bind](/rebind/main.go) - Dns Server

* [Re-Web](/reweb/main.go) - Rest API service

# Goals

Goal is to define a resilient DNS service, with remote configuration via API service.

This service groups records per *Groups*. Any group is characterised by a dns forwarder(s) list, a domain list and the DNS Records.


# API implementation

At the moment the available API versions are:

* v1 -> Base dns components configuration


# Easy to use

The design wants this DNS to be configured via API, and automatically resources becomes active ( in milliseconds ) without any manual operation.

Call the REST API and manage your records and groups.

# V1 Rest API

Rest API in release V1 contains following endpoints:

* /v1/dns - Dns server main components list

# V1 Rest Path /v1/dns

Implemented METHODS:

* *GET* - Retrieves list of groups, with query param action-template it return templeates for available requests and methods
* *POST* - 

# Docker image

At the moment a docker image is available within following components:

* Re-Bind - Dns Server

* Re-Bind Net pipe connector - used to communicate internally with the API Rest server

* Re-Web - Rest API Server

* Re-Web  Net pipe connector - used to communicate internally with the Dns server

* log combiner - used to combine output coming from server logs and reloading the logs on rotation happening ...

The docker image is available at: 
[re-bind docker image](https://hub.docker.com/repository/docker/hellgate75/rebind)


Enjoy the experience.

## License

The library is licensed with [LGPL v. 3.0](/LICENSE) clauses, with prior authorization of author before any production or commercial use. Use of this library or any extension is prohibited due to high risk of damages due to improper use. No warranty is provided for improper or unauthorized use of this library or any implementation.

Any request can be prompted to the author [Fabrizio Torelli](https://www.linkedin.com/in/fabriziotorelli) at the following email address:

[hellgate75@gmail.com](mailto:hellgate75@gmail.com)

