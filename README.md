# testnet-server
 Server for the testnet fund
To run the server create a file named `.tokens` in the same folder as main with the content
```
Main funder account seed
Bearer easyship auth
api token of Igg
acces token of igg
```
Create two files:
- `userDetails.txt`: which holds the information of users who already joined. Initial an empty file by `touch userDetails.txt`. The format of the saved file is a simple txt with below information:

`Date and Time of getting funded`, `contribution ID`, `Aura account`

- `contributions.csv`: which holds the details of contributions. You can export it from Indiegogo or create it manually. When contributing to `contributions.csv`, please ensure your file includes the following fields:

| Order No. | Pledge ID | Referrer ID | Fulfillment Status | Funding Date | Payment Method | Appearance | Name | Email | Amount | Shipping Fees | Platform Fee | Transaction Fee | Perk ID | Perk | Shipping Name | Shipping Phone Number | Shipping Address | Shipping Address 2 | Shipping City | Shipping State/Province | Shipping Zip/Postal Code | Shipping Country | Carrier | Tracking Number | Item Name | Item SKU | Option 1 | Option 2 | Option 3 | Item Name | Item SKU | Option 1 | Option 2 | Option 3 | Item Name | Item SKU | Option 1 | Option 2 | Option 3 | Item Name | Item SKU | Option 1 | Option 2 | Option 3 |
|-----------|-----------|-------------|---------------------|--------------|----------------|------------|------|-------|--------|---------------|--------------|-----------------|---------|------|----------------|-----------------------|-------------------|---------------------|---------------|-------------------------|--------------------------|------------------|---------|----------------|-----------|----------|----------|----------|----------|-----------|----------|----------|----------|----------|-----------|----------|----------|----------|----------|-----------|----------|----------|----------|----------|

Please make sure each entry is correctly placed under the corresponding column header.

In the same folder and then you can build or run it with go
```go
go build -o testnet-server .
testnet-server
```
