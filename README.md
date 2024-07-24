# testnet-server
 Server for the testnet fund
To run the server create a file named `.tokens` in the same folder as main with the content
```
Main funder account seed
Bearer easyship auth
api token of Igg
acces token of igg
```
Create four files:
- `userDetails.txt`: which holds the information of users who already joined. Initial an empty file by `touch userDetails.txt`. The format of the saved file is a simple txt with below information:

`Date and Time of getting funded`, `contribution ID`, `Aura account`

- `contributions-masked.csv`: which holds the details of contributions. You can export it from Indiegogo or create it manually. When contributing to `contributions.csv`, please ensure your file includes the following fields:

| Order No. | Email | Amount | Shipping Phone Number (Masked to the last 4 digist only for security) |
|-----------|-----------|-------------|---------------------|

- `.tokens` : the first line of this file holds the seed to an account with enough funds to execute join requests and fund them with gas token
- `brevo.key` this contains the API key for email server

Please make sure each entry is correctly placed under the corresponding column header.

In the same folder and then you can build or run it with go
```go
go build -o testnet-server .
testnet-server --opensea-api xxxxxx
```

and then an example service file `/etc/systemd/system/testnet-server.service` is like:

```
[Unit]
Description=Testnet Server

[Service]
TimeoutStartSec=0
Type=simple
User=root
WorkingDirectory=/home/${USER}/testnet-server
ExecStart=/home/${USER}/testnet-server/testnet-server
Restart=always
StandardOutput=file:/var/log/testnet-server.log
StandardError=file:/var/log/testnet-server.err

[Install]
WantedBy=multi-user.target
```

And then run:

```
sudo systemctl daemon-reload

systemctl enable testnet-server
```
