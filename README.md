# Instructions to run this project

1. Install the dependencies using the command:
   ```
   go mod tidy
   ```

2. Run the server using the command:
   ```
   cd server
   go run *.go
   ```

3. Run the peer using the command:
   ```
   cd peer
   go run *.go
   ```

Make sure you create a .env file in the root directory of the project and add the following:
```
SERVER_IP_ADDRESS = <IP address of the server>
SERVER_CONNECTIONS_PORT = 7734

```



