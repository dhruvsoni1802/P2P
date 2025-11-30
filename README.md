# Instructions to run this project

Make sure you have go programming language installed in your system. We have tested this using go version go1.25.1 in our MacOS system. 

1. If you are downloading the source code or using the zip, then make sure you are inside the root directory which is P2P directory.

2. Install the dependencies using the command:
   ```
   go mod tidy
   ```

3. Run the server using the command:
   ```
   cd server
   go run *.go
   ```

4. Run the peer using the command:
   ```
   cd peer
   go run *.go
   ```

Make sure you create a .env file in the root directory of the project and add the following:
```
SERVER_IP_ADDRESS = <IP address of the server>
SERVER_CONNECTIONS_PORT = 7734

```

Also look at the Sample_Commands.txt and change specific fields to run the commands according to the current running system.

We have already added 2 sample files inside peer/RFCs called 1_Hello.txt and 4_Hey.txt. These files follow the convention of RFCNumber_RFCTitle.txt. You can add more text files inside the same RFC directory for testing the system



