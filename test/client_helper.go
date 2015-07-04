package test

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/nbusy/devastator"
	"github.com/nbusy/neptulon/jsonrpc"
)

// client cert for testing
var (
	// host = client.127.0.0.1, cn = client.127.0.0.1, org = devastator
	clientCert = `-----BEGIN CERTIFICATE-----
MIIEXzCCArOgAwIBAgIQSwLu5wcVkGlY9qOW1pY4KTALBgkqhkiG9w0BAQswKTET
MBEGA1UEChMKZGV2YXN0YXRvcjESMBAGA1UEAxMJMTI3LjAuMC4xMCAXDTE1MDUw
OTA5MDk1OFoYDzIzMDUwMjI4MDkwOTU4WjAwMRMwEQYDVQQKEwpkZXZhc3RhdG9y
MRkwFwYDVQQDExBjbGllbnQuMTI3LjAuMC4xMIIBuDANBgkqhkiG9w0BAQEFAAOC
AaUAMIIBoAKCAZcAzTlkFCni6zUZza1X8UPkRLwlcMDBKpbxdQZrSQAhHqKZe7xg
ZthomSJ6Ahd0ueiXkGwDorhZLxW+1/iTQww4yDSMWmCNeflzuPp8E3maNTucYSSz
QGR4+GJx7+336spFBaT/ikGLHnbVaW6lGUAvKbRtUHFSoyfYit3Ar6x0+OnQrq+a
x4GFe+XiOvlZXqgKGQm1OWe58SFCbnvz+r0vbWIPabXk66gJMQ4yA0mjPKo7hEuE
h0XpUn2QiaSehS+NVmxjuM3j5fjPjMI4J3CXK4Ax9tM/imA2TbPig69T24CZpkHq
J/T99fgvEyO5+RcrrfHOjLHdVnyDceRXAB3avSOX8PiB3vF5kNruWN5GTLXxWaxW
IQkCtRjTWcN1APqFmkB0WEQcI+qgoE9WPSKUD2SMKzujd/HVn3xhKM3zAoSOZgBf
iXzQy6BRTxr7dE8/C/pultmVc4xS8YNGc0aIRGnht9s8yUSw7RQLb00Us4HE0RtX
FtRDEG+65V2d7yWtqN/ZQPo8832PfPSvR4XyrRUNp+Df9wIDAQABo1IwUDAOBgNV
HQ8BAf8EBAMCAKAwEwYDVR0lBAwwCgYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAb
BgNVHREEFDASghBjbGllbnQuMTI3LjAuMC4xMAsGCSqGSIb3DQEBCwOCAZcANbLw
HOQ2U84YnA30WUMhh6ETSJlwqOzSn2pyffI/EzjbBHYaQEWYWrZ7srNfF3+GctXj
rxyPodTJgzCrNFVAyE8V/Xm3DmCsxhJEGAapO/POFJ3wQNdYVK+yEox+lJDllz2Y
iffeV+WEV/6jixRsNDz5EXwljrUZiIeEXCWe+vpienOxB+Z+7Rh4JUbD+LuJCilx
XBs7uHSY8f2kCmu7lbI+5OrqO8lDhzGltBdt1cIB0T5dERaxn1JwWPuhCtyq2VQ+
1aaVD3t2E4ItlY4KNBW7tq9BLaO25cqwhIArWbuFY5ahV/hY/oJaNJI2V2D7f2YL
z6GHGVTA85rbYp9SLFKTQ9uT7OOL7LVburfKIgllgvgBPe1v+RwND7W6SWeSVZ/K
8CAQB4B8dzVCmpLbdvq0NBp0+8I2BZ0r+T9E42ddAnQqIt3x9dMj5jNX83LAuuzo
vrd2owpJtyFDW9U2uZrGqbzzRNTSAbcW5Oyf9PiCAxT8oDQ6YapEXr9dGRpMSGu/
ExyUsaeG4ZAGeYJp5v8Xda2CTw==
-----END CERTIFICATE-----`

	// RSA 3248 bits
	clientKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

	clientCertBytes = []byte(clientCert)
	clientKeyBytes  = []byte(clientKey)
)

// ClientHelper is a JSON-RPC client wrapper with built-in error logging for testing.
type ClientHelper struct {
	client    *jsonrpc.Client
	testing   *testing.T
	cert, key []byte
}

// NewClientHelper creates a new JSON-RPC client wrapper which has built-in error logging for testing.
func NewClientHelper(t *testing.T) *ClientHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	return &ClientHelper{testing: t}
}

// DefaultCert attaches default test client certificate to the connection.
func (c *ClientHelper) DefaultCert() *ClientHelper {
	c.cert = clientCertBytes
	c.key = clientKeyBytes
	return c
}

// Cert attaches given PEM encoded client certificate to the connection.
func (c *ClientHelper) Cert(cert, key []byte) *ClientHelper {
	c.cert = cert
	c.key = key
	return c
}

// Dial initiates a connection.
func (c *ClientHelper) Dial() *ClientHelper {
	addr := "127.0.0.1:" + devastator.Conf.App.Port

	// retry connect in case we're operating on a very slow machine
	for i := 0; i <= 5; i++ {
		client, err := jsonrpc.Dial(addr, caCertBytes, c.cert, c.key, false) // no need for debug mode on client conn
		if err != nil {
			if operr, ok := err.(*net.OpError); ok && operr.Op == "dial" && operr.Err.Error() == "connection refused" {
				time.Sleep(time.Millisecond * 50)
				continue
			} else if i == 5 {
				c.testing.Fatalf("Cannot connect to server address %v after 5 retries, with error: %v", addr, err)
			}
			c.testing.Fatalf("Cannot connect to server address %v with error: %v", addr, err)
		}

		if i != 0 {
			c.testing.Logf("WARNING: it took %v retries to connect to the server, which might indicate code issues or slow machine.", i)
		}

		client.SetReadDeadline(10)
		c.client = client
		return c
	}

	return nil
}

// WriteRequest writes a request to a client connection with error logging for testing.
func (c *ClientHelper) WriteRequest(method string, params interface{}) (reqID string) {
	id, err := c.client.WriteRequest(method, params)
	if err != nil {
		c.testing.Fatal("Failed to write request to client connection:", err)
	}
	return id
}

// ReadMsg reads a JSON-RPC message from a client connection with error logging for testing.
func (c *ClientHelper) ReadMsg() (req *jsonrpc.Request, res *jsonrpc.Response, not *jsonrpc.Notification) {
	req, res, not, err := c.client.ReadMsg()
	if err != nil {
		c.testing.Fatal("Failed to read message from client connection:", err)
	}

	return
}

// ReadRes reads a response object from a client connection. If incoming message is not a response, an error is logged.
func (c *ClientHelper) ReadRes() *jsonrpc.Response {
	_, res, _, err := c.client.ReadMsg()
	if err != nil {
		c.testing.Fatal("Failed to read response from client connection:", err)
	}

	return res
}

// VerifyConnClosed verifies that the connection is in closed state.
// Verification is done via reading from the channel and checking that returned error is io.EOF.
func (c *ClientHelper) VerifyConnClosed() bool {
	_, _, _, err := c.client.ReadMsg()
	if err != io.EOF {
		return false
	}

	return true
}

// Close closes a client connection.
func (c *ClientHelper) Close() {
	if err := c.client.Close(); err != nil {
		c.testing.Fatal("Failed to close client connection:", err)
	}
}
