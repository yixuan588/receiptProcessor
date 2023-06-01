## receiptProcessor

receiptProcessor is a programm used for handle receipt and calculate receipt points

## Installation

Download the whole package and install go from [go professional website](https://go.dev/doc/install)
Set go path environment.

## Usage

Open command line and move to the directory **\receiptProcessor\main**

 run this command below:

```
go run main.go
```

Go to web browser using Talend API tester to query the server.

Post the receipt by this url and get the id **

```
http://localhost:8080/receipts/process
```

Get the points by this url**

```
http://localhost:8080/receipts/){id}/points
```

