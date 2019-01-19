# Sipclient

A Simple SIP2 client for testing Self Checkout equipment.

Implements a simple Socket reader repl for interacting with a SIP server

# Usage

```
Usage of sipclient:
  -adr string
    	sip server address
  -inst string
    	sip institution id
  -pass string
    	sip login pass
  -user string
    	sip login user
```   	

State commands:

```
branch <branchid>
patron <patronid>
barcode <item id>
```

Action commands:

```
checkin
checkout
renew
```