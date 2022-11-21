# Reverse proxy
## Motivation

The reverse proxy sits between clients and multiple service instances. It accepts a request from a client, forwards the request to one of the service instances, and returns the results from the service that actually processed the request to the client.

To run the project check the dependencies section and run from the root of the project:
```bash
docker compose up
```
Which service instance handles the request is determined by the `Host` header, it contains the domain of the service instance. A request that will be handled by one of the hosts from the service with the domain `service1.com` looks like this for example:
```bash
curl -H "Host: service1.com" -s -o /dev/null -v locahost:8080  
```
The reverse proxy supports 2 load-balancing strategies: 
* Round Robin: the received requests are handled in a circular order by each host, without priority
* Random: the requests are handled by hosts that are selected at random, using the timestamp as seed to generate random numbers  
## Features

* Listens on HTTP requests and forwards them to one of the service instances
* Responses from the service instances are forwarded back to the client
* Supports load-balancing strategies: Round Robin, Random
* Scheduled health checks for the service instances
* Configuration via YAML file
* Supports HTTP 1.1

## Config file
The reverse proxy is configured through a `config.yaml` file in the `config` folder:  
```yaml
proxy:
  lbPolicy: "RANDOM"
  listen:
    address: "127.0.0.1"
    port: 8080
  services:
    - name: service1
      domain: service1.com
      hosts:
        - address: "http://s1_host0"
          port: 80
        - address: "http://s1_host1"
          port: 80 
        - address: "http://s1_host2"
          port: 80
        - address: "http://s1_host3" # this is a failing host for the demo
          port: 80
    - name: service2
      domain: service2.com
      hosts:
        - address: "http://s2_host0"
          port: 80
        - address: "http://s2_host1"
          port: 80 
        - address: "http://s2_host2"
          port: 80
        - address: "http://s2_host3" # this is a failing host for the demo
          port: 80
```
* proxy.lbPolicy specifies the load-balancing strategy that the reverse proxy uses to forward the incoming requests
* all the fields of proxy.listen are used for the configuration of the reverse proxy
* all the fields of proxy.services are service instances with multiple hosts that will receive the redirected requests

For the demo [docker compose](https://docs.docker.com/compose/install/) was used to build and run the container of the application and of the service instances (2 services with 3 hosts each). For the container images of the hosts [httpbin](https://httpbin.org/) was used.
## Docker compose file
```yaml Docker-compose
version: "3.9"
services:
  s1_host0:
    image: kennethreitz/httpbin
    ports: 
      - "9090:80"
  s1_host1:
    image: kennethreitz/httpbin
    ports: 
      - "9091:80"
  s1_host2:
    image: kennethreitz/httpbin
    ports: 
      - "9092:80"
  s2_host0:
    image: kennethreitz/httpbin
    ports: 
      - "9093:80"
  s2_host1:
    image: kennethreitz/httpbin
    ports: 
      - "9094:80"
  s2_host2:
    image: kennethreitz/httpbin
    ports: 
      - "9095:80"
  app:
    build: 
      context: . 
      dockerfile: ./docker/Dockerfile
    ports:
      - "8080:8080"
```
## Dependencies needed to run the project
* go 
* docker
* docker compose
* Linux machine