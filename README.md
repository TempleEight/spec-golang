# spec-golang

This repository holds the example microservice backend written for the development of [Temple](https://github.com/TempleEight/temple), and is therefore written with the goal of being as general as possible.

This example models the backend for a Tinder-like application, with services for users and matches, as well as a separate service for authentication.
* PostgreSQL is used as the backing datastore
* Kong acts as the API gateway and load balancer
* Prometheus and Grafana are used for gathering and displaying metrics
* OpenAPI is used to generate frontend APIs
* Scripts are supplied for orchestration with either Kubernetes or Docker Compose

For more information on the architecture we settled on, check out the docs:
* [System Architecture](https://templeeight.github.io/temple-docs/docs/arch/system)
* [Service Architecture](https://templeeight.github.io/temple-docs/docs/arch/service)
