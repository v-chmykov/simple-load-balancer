# Simple Load Balancer

This is a simple load balancer written in Go.
It uses a round robin as a distribution algorithm.

## How to round

1. Run the `backend.go` instance for three ports:
```bash
go build -o backend backend.go
./backend -port 8081 &
./backend -port 8082 &
./backend -port 8083 &
```
OR
```bash
go run backend.go -port 8081 &
go run backend.go -port 8082 &
go run backend.go -port 8083 &
```

2. Run the load balancer:
```bash
go build -o load-balancer load-balancer.go
./load-balancer -port 8080
```
OR
```bash
go run load-balancer.go -port 8080
```

3. Make a request to the load balancer:
```bash
curl http://localhost:8080/
```
Repeat it few times
