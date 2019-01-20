package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/knakk/sip"
)

// Ref: https://github.com/knakk/sip/blob/master/msgdef.go

func (c *Client) login() {
	fmt.Println("LOGIN")
	req := sip.NewMessage(sip.MsgReqLogin)
	req.AddField(c.user)
	req.AddField(c.pass)
	req.AddField(sip.Field{Type: sip.FieldUIDAlgorithm, Value: "0"})
	req.AddField(sip.Field{Type: sip.FieldPWDAlgorithm, Value: "0"})

	fmt.Println(req.String())
	res := c.send(req)
	fmt.Println(res.String())
	return
}
func (c *Client) ping() {
	req := sip.NewMessage(sip.MsgReqStatus)
	req.AddField(sip.Field{Type: sip.FieldStatusCode, Value: "1"})
	req.AddField(sip.Field{Type: sip.FieldMaxPrintWidth, Value: "010"})
	req.AddField(sip.Field{Type: sip.FieldProtocolVersion, Value: "2.00"})

	fmt.Println(req.String())
	res := c.send(req)
	fmt.Println(res.String())
	return
}

func (c *Client) patronInfo() {
	if c.State["patron"] == "" {
		fmt.Println("Missing patron!")
		return
	}
	fmt.Printf("fetching info on patron %s\n", c.State["patron"])
	req := sip.NewMessage(sip.MsgReqPatronInformation)
	req.AddField(c.inst)
	t := sipTime()
	req.AddField(sip.Field{Type: sip.FieldLanguage, Value: "010"})
	req.AddField(sip.Field{Type: sip.FieldTransactionDate, Value: t})
	req.AddField(sip.Field{Type: sip.FieldSummary, Value: "YYYYYYYYYY"})
	req.AddField(sip.Field{Type: sip.FieldPatronIdentifier, Value: c.State["patron"]})

	fmt.Println(req.String())
	res := c.send(req)
	fmt.Println(res.String())
	return
}

func (c *Client) checkin() {
	if c.State["barcode"] == "" || c.State["branch"] == "" {
		fmt.Println("Missing barcode or branch!")
		return
	}
	fmt.Printf("checking in barcode %s at %s\n", c.State["barcode"], c.State["branch"])
	req := sip.NewMessage(sip.MsgReqCheckin)
	//req.AddField(c.inst)
	t := sipTime()
	req.AddField(sip.Field{Type: sip.FieldNoBlock, Value: "Y"})
	req.AddField(sip.Field{Type: sip.FieldTransactionDate, Value: t})
	req.AddField(sip.Field{Type: sip.FieldReturnDate, Value: t})
	req.AddField(sip.Field{Type: sip.FieldCurrentLocation, Value: c.State["branch"]})
	req.AddField(sip.Field{Type: sip.FieldInstitutionID, Value: c.State["branch"]})
	req.AddField(sip.Field{Type: sip.FieldTerminalLocation, Value: c.State["branch"]})
	req.AddField(sip.Field{Type: sip.FieldItemIdentifier, Value: c.State["barcode"]})
	fmt.Println(req.String())
	res := c.send(req)
	fmt.Println(res.String())
	return
}

func (c *Client) checkout() {
	if c.State["barcode"] == "" || c.State["branch"] == "" || c.State["patron"] == "" {
		fmt.Println("Missing patron, barcode or branch!")
		return
	}
	fmt.Printf("checking out barcode %s for %s at %s\n", c.State["barcode"], c.State["patron"], c.State["branch"])
	req := sip.NewMessage(sip.MsgReqCheckout)
	//req.AddField(c.inst)
	t := sipTime()
	req.AddField(sip.Field{Type: sip.FieldRenewalPolicy, Value: "Y"})
	req.AddField(sip.Field{Type: sip.FieldNoBlock, Value: "N"})
	req.AddField(sip.Field{Type: sip.FieldTransactionDate, Value: t})
	req.AddField(sip.Field{Type: sip.FieldNbDueDate, Value: t})
	req.AddField(sip.Field{Type: sip.FieldInstitutionID, Value: c.State["branch"]})
	req.AddField(sip.Field{Type: sip.FieldTerminalLocation, Value: c.State["branch"]})
	req.AddField(sip.Field{Type: sip.FieldItemIdentifier, Value: c.State["barcode"]})
	req.AddField(sip.Field{Type: sip.FieldPatronIdentifier, Value: c.State["patron"]})

	fmt.Println(req.String())
	res := c.send(req)
	fmt.Println(res.String())
	return
}

func (c *Client) renew() {
	if c.State["barcode"] == "" || c.State["branch"] == "" || c.State["patron"] == "" {
		fmt.Println("Missing patron, barcode or branch!")
		return
	}
	fmt.Printf("renewing barcode %s for %s at %s\n", c.State["barcode"], c.State["patron"], c.State["branch"])
	req := sip.NewMessage(sip.MsgReqRenew)
	//req.AddField(c.inst)
	t := sipTime()
	req.AddField(sip.Field{Type: sip.FieldThirdPartyAllowd, Value: "Y"})
	req.AddField(sip.Field{Type: sip.FieldNoBlock, Value: "Y"})
	req.AddField(sip.Field{Type: sip.FieldTransactionDate, Value: t})
	req.AddField(sip.Field{Type: sip.FieldNbDueDate, Value: t})
	req.AddField(sip.Field{Type: sip.FieldPatronIdentifier, Value: c.State["patron"]})
	req.AddField(sip.Field{Type: sip.FieldInstitutionID, Value: c.State["branch"]})
	req.AddField(sip.Field{Type: sip.FieldTerminalLocation, Value: c.State["branch"]})
	req.AddField(sip.Field{Type: sip.FieldItemIdentifier, Value: c.State["barcode"]})

	res := c.send(req)
	fmt.Println(res.String())
	return
}

func (c *Client) send(req sip.Message) sip.Message {
	if _, err := req.Encode(c.Conn); err != nil {
		log.Println("error writing SIP request to remote: " + err.Error())
		return sip.Message{}
	}
	rBuf := make([]byte, 16*1024) // read buffer
	n, err := c.Conn.Read(rBuf)
	if err != nil {
		if err != io.EOF {
			log.Println("error reading SIP server response: " + err.Error())
		}
		return sip.Message{}
	}
	res, err := sip.Decode(rBuf[:n])
	if err != nil {
		log.Println("error decoding SIP response: " + err.Error())
		return sip.Message{}
	}
	return res
}

func sipTime() string {
	t := time.Now()
	return fmt.Sprintf("%s", t.Format("20060102    150405"))
}

func usage() string {
	return `
state    print state object
ping     keepalive ping

state setters:
barcode  <barcode>
branch   <branchcode>
patron   <cardnumber>

methods:
checkin
checkout
renew
patronInfo
`
}

func readStdin(c *Client) {
	for {
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		cmd := strings.Fields(line)
		switch cmd[0] {
		case "", "?", "h":
			fmt.Println(usage())
		/* --- state commands --- */
		case "state":
			fmt.Printf("status:\n----\n%v\n", c.State)
		case "ping":
			c.ping()
		case "branch":
			c.State["branch"] = cmd[1]
		case "barcode":
			c.State["barcode"] = cmd[1]
		case "patron":
			c.State["patron"] = cmd[1]
		/* --- action commands --- */
		case "checkin":
			c.checkin()
		case "checkout":
			c.checkout()
		case "renew":
			c.renew()
		case "patronInfo":
			c.patronInfo()
		default:
			fmt.Println("UNKNOWN COMMAND")
		}
	}
	//c.Conn.Write([]byte(msg))
}

type Client struct {
	user  sip.Field
	pass  sip.Field
	inst  sip.Field
	Conn  net.Conn
	State map[string]string // simplest possible state map for patron, branch, barcode etc
}

func newClient(adr string, u string, p string, i string) *Client {
	// Clientect to remote.
	c, err := net.Dial("tcp", adr)
	if err != nil {
		log.Println("error connecting to remote: " + err.Error())
		return nil
	}
	return &Client{
		user:  sip.Field{Type: sip.FieldLoginUserID, Value: u},
		pass:  sip.Field{Type: sip.FieldLoginPassword, Value: p},
		inst:  sip.Field{Type: sip.FieldInstitutionID, Value: i},
		Conn:  c,
		State: make(map[string]string),
	}
}

func main() {
	http.DefaultClient.Timeout = 5 * time.Second
	var (
		adr  = flag.String("adr", "", "sip server address")
		user = flag.String("user", "", "sip login user")
		pass = flag.String("pass", "", "sip login pass")
		inst = flag.String("inst", "", "sip institution id")
	)
	flag.Parse()

	if *adr == "" || *user == "" || *pass == "" {
		flag.Usage()
		os.Exit(2)
	}

	c := newClient(*adr, *user, *pass, *inst)
	fmt.Printf("connected to to: %s\n", c.Conn.RemoteAddr().String())
	c.login()
	defer c.Conn.Close()
	readStdin(c)
}
