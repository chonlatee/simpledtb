# simpledtb
## I just want to try to code about channel, networking on golang.
### FYI: start worker first and then start server

- terminal tab1 ```go run cmd/worker/main.go -port=:3331 -server=:3000```

- terminal tab2 ```go run cmd/worker/main.go -port=:3332 -server=:3000```

- terminal tab3 ```go run cmd/server/main.go -port=:3000```