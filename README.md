# testnet-server
 Server for the testnet fund
To run the server create a file named .tokens in the same folder as main with the content
```
Main funder account seed
Bearer easyship auth
```
Create two files:
- userDetails.txt: which holds the information of users who already joined
- contributions.csv: which holds the details of contributions

In the same folder and then you can build or run it with go
```go
go build -o testnet-server .
```