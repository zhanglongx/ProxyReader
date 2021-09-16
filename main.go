package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
)

type FileStream struct {
	path string
	f    *os.File
}

func NewFileStream(path string) *FileStream {
	return &FileStream{path, nil}
}

func (fs *FileStream) Write(b []byte) (nr int, err error) {
	if fs.f == nil {
		fs.f, err = os.Create(fs.path)
		if err != nil {
			return 0, err
		}
	}
	return fs.f.Write(b)
}

func (fs *FileStream) Close() error {
	fmt.Println("Close", fs.path)
	// if fs.f == nil {
	// 	return errors.New("FileStream was never written into")
	// }
	return fs.f.Close()
}

type Meta struct {
	req  *http.Request
	resp *http.Response
	err  error
	t    time.Time
	sess int64
	from string
}

func fprintf(nr *int64, err *error, w io.Writer, pat string, a ...interface{}) {
	if *err != nil {
		return
	}
	var n int
	n, *err = fmt.Fprintf(w, pat, a...)
	*nr += int64(n)
}

func write(nr *int64, err *error, w io.Writer, b []byte) {
	if *err != nil {
		return
	}
	var n int
	n, *err = w.Write(b)
	*nr += int64(n)
}

func (m *Meta) WriteTo(w io.Writer) (nr int64, err error) {
	if m.req != nil {
		fprintf(&nr, &err, w, "Type: request\r\n")
	} else if m.resp != nil {
		fprintf(&nr, &err, w, "Type: response\r\n")
	}
	fprintf(&nr, &err, w, "ReceivedAt: %v\r\n", m.t)
	fprintf(&nr, &err, w, "Session: %d\r\n", m.sess)
	fprintf(&nr, &err, w, "From: %v\r\n", m.from)
	if m.err != nil {
		// note the empty response
		fprintf(&nr, &err, w, "Error: %v\r\n\r\n\r\n\r\n", m.err)
	} else if m.req != nil {
		fprintf(&nr, &err, w, "\r\n")
		buf, err2 := httputil.DumpRequest(m.req, false)
		if err2 != nil {
			return nr, err2
		}
		write(&nr, &err, w, buf)
	} else if m.resp != nil {
		fprintf(&nr, &err, w, "\r\n")
		buf, err2 := httputil.DumpResponse(m.resp, false)
		if err2 != nil {
			return nr, err2
		}
		write(&nr, &err, w, buf)
	}
	return
}

// HttpLogger is an asynchronous HTTP request/response logger. It traces
// requests and responses headers in a "log" file in logger directory and dumps
// their bodies in files prefixed with the session identifiers.
// Close it to ensure pending items are correctly logged.
type HttpLogger struct {
	path  string
	c     chan *Meta
	errch chan error
}

func NewLogger(basepath string) (*HttpLogger, error) {
	f, err := os.Create(path.Join(basepath, "log"))
	if err != nil {
		return nil, err
	}
	logger := &HttpLogger{basepath, make(chan *Meta), make(chan error)}
	go func() {
		for m := range logger.c {
			if _, err := m.WriteTo(f); err != nil {
				log.Println("Can't write meta", err)
			}
		}
		logger.errch <- f.Close()
	}()
	return logger, nil
}

func (logger *HttpLogger) LogResp(resp *http.Response, ctx *goproxy.ProxyCtx) {
	body := path.Join(logger.path, fmt.Sprintf("%d_resp", ctx.Session))
	from := ""
	if ctx.UserData != nil {
		from = ctx.UserData.(*transport.RoundTripDetails).TCPAddr.String()
	}
	if resp == nil {
		resp = emptyResp
	} else {
		resp.Body = NewTeeReadCloser(resp.Body, NewFileStream(body))
	}
	logger.LogMeta(&Meta{
		resp: resp,
		err:  ctx.Error,
		t:    time.Now(),
		sess: ctx.Session,
		from: from})
}

var emptyResp = &http.Response{}
var emptyReq = &http.Request{}

func (logger *HttpLogger) LogReq(req *http.Request, ctx *goproxy.ProxyCtx) {
	body := path.Join(logger.path, fmt.Sprintf("%d_req", ctx.Session))
	if req == nil {
		req = emptyReq
	} else {
		req.Body = NewTeeReadCloser(req.Body, NewFileStream(body))
	}
	logger.LogMeta(&Meta{
		req:  req,
		err:  ctx.Error,
		t:    time.Now(),
		sess: ctx.Session,
		from: req.RemoteAddr})
}

func (logger *HttpLogger) LogMeta(m *Meta) {
	logger.c <- m
}

func (logger *HttpLogger) Close() error {
	close(logger.c)
	return <-logger.errch
}

// TeeReadCloser extends io.TeeReader by allowing reader and writer to be
// closed.
type TeeReadCloser struct {
	r io.Reader
	w io.WriteCloser
	c io.Closer
}

func NewTeeReadCloser(r io.ReadCloser, w io.WriteCloser) io.ReadCloser {
	return &TeeReadCloser{io.TeeReader(r, w), w, r}
}

func (t *TeeReadCloser) Read(b []byte) (int, error) {
	return t.r.Read(b)
}

// Close attempts to close the reader and write. It returns an error if both
// failed to Close.
func (t *TeeReadCloser) Close() error {
	err1 := t.c.Close()
	err2 := t.w.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// stoppableListener serves stoppableConn and tracks their lifetime to notify
// when it is safe to terminate the application.
type stoppableListener struct {
	net.Listener
	sync.WaitGroup
}

type stoppableConn struct {
	net.Conn
	wg *sync.WaitGroup
}

func newStoppableListener(l net.Listener) *stoppableListener {
	return &stoppableListener{l, sync.WaitGroup{}}
}

func (sl *stoppableListener) Accept() (net.Conn, error) {
	c, err := sl.Listener.Accept()
	if err != nil {
		return c, err
	}
	sl.Add(1)
	return &stoppableConn{c, &sl.WaitGroup}, nil
}

func (sc *stoppableConn) Close() error {
	sc.wg.Done()
	return sc.Conn.Close()
}

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("l", ":2081", "on which address should the proxy listen")
	flag.Parse()
	setCA(caCert, caKey)

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.Verbose = *verbose
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatal("Can't create dir", err)
	}
	logger, err := NewLogger("db")
	if err != nil {
		log.Fatal("can't open log file", err)
	}
	// tr := transport.Transport{Proxy: transport.ProxyFromEnvironment}
	// // For incoming request on domain, override the RoundTripper to extract
	// // connection information. Store it is session context log it after
	// // handling the response.
	// // proxy.OnRequest(goproxy.DstHostIs("www.baidu.com")).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	// proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	// 	ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (resp *http.Response, err error) {
	// 		ctx.UserData, resp, err = tr.DetailedRoundTrip(req)
	// 		return
	// 	})
	// 	logger.LogReq(req, ctx)
	// 	return req, nil
	// })
	proxy.OnResponse(goproxy.DstHostIs("www.baidu.com:443")).DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		logger.LogResp(resp, ctx)
		return resp
	})
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal("listen:", err)
	}
	sl := newStoppableListener(l)
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Println("Got SIGINT exiting")
		sl.Add(1)
		sl.Close()
		logger.Close()
		sl.Done()
	}()
	log.Println("Starting Proxy")
	http.Serve(sl, proxy)
	sl.Wait()
	log.Println("All connections closed - exit")
}

var caCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIJAKe/ZGdfcHdPMA0GCSqGSIb3DQEBCwUAMGAxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQxGTAXBgNVBAMMEGRlbW8gZm9yIGdvcHJveHkwHhcNMTYw
OTI3MTQzNzQ3WhcNMTkwOTI3MTQzNzQ3WjBgMQswCQYDVQQGEwJBVTETMBEGA1UE
CAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRk
MRkwFwYDVQQDDBBkZW1vIGZvciBnb3Byb3h5MIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEA2+W48YZoch72zj0a+ZlyFVY2q2MWmqsEY9f/u53fAeTxvPE6
1/DnqsydnA3FnGvxw9Dz0oZO6xG+PZvp+lhN07NZbuXK1nie8IpxCa342axpu4C0
69lZwxikpGyJO4IL5ywp/qfb5a2DxPTAyQOQ8ROAaydoEmktRp25yicnQ2yeZW//
1SIQxt7gRxQIGmuOQ/Gqr/XN/z2cZdbGJVRUvQXk7N6NhQiCX1zlmp1hzUW9jwC+
JEKKF1XVpQbc94Bo5supxhkKJ70CREPy8TH9mAUcQUZQRohnPvvt/lKneYAGhjHK
vhpajwlbMMSocVXFvY7o/IqIE/+ZUeQTs1SUwQIDAQABo1AwTjAdBgNVHQ4EFgQU
GnlWcIbfsWJW7GId+6xZIK8YlFEwHwYDVR0jBBgwFoAUGnlWcIbfsWJW7GId+6xZ
IK8YlFEwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAoFUjSD15rKlY
xudzyVlr6n0fRNhITkiZMX3JlFOvtHNYif8RfK4TH/oHNBTmle69AgixjMgy8GGd
H90prytGQ5zCs1tKcCFsN5gRSgdAkc2PpRFOK6u8HwOITV5lV7sjucsddXJcOJbQ
4fyVe47V9TTxI+A7lRnUP2HYTR1Bd0R/IgRAH57d1ZHs7omHIuQ+Ea8ph2ppXMnP
DXVOlZ9zfczSnPnQoomqULOU9Fq2ycyi8Y/ROtAHP6O7wCFbYHXhxojdaHSdhkcd
troTflFMD2/4O6MtBKbHxSmEG6H0FBYz5xUZhZq7WUH24V3xYsfge29/lOCd5/Xf
A+j0RJc/lQ==
-----END CERTIFICATE-----`)

var caKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2+W48YZoch72zj0a+ZlyFVY2q2MWmqsEY9f/u53fAeTxvPE6
1/DnqsydnA3FnGvxw9Dz0oZO6xG+PZvp+lhN07NZbuXK1nie8IpxCa342axpu4C0
69lZwxikpGyJO4IL5ywp/qfb5a2DxPTAyQOQ8ROAaydoEmktRp25yicnQ2yeZW//
1SIQxt7gRxQIGmuOQ/Gqr/XN/z2cZdbGJVRUvQXk7N6NhQiCX1zlmp1hzUW9jwC+
JEKKF1XVpQbc94Bo5supxhkKJ70CREPy8TH9mAUcQUZQRohnPvvt/lKneYAGhjHK
vhpajwlbMMSocVXFvY7o/IqIE/+ZUeQTs1SUwQIDAQABAoIBAHK94ww8W0G5QIWL
Qwkc9XeGvg4eLUxVknva2Ll4fkZJxY4WveKx9OCd1lv4n7WoacYIwUGIDaQBZShW
s/eKnkmqGy+PvpC87gqL4sHvQpuqqJ1LYpxylLEFqduWOuGPUVC2Lc+QnWCycsCS
CgqZzsbMq0S+kkKRGSvw32JJneZCzqLgLNssQNVk+Gm6SI3s4jJsGPesjhnvoPaa
xZK14uFpltaA05GSTDaQeZJFEdnnb3f/eNPc2xMEfi0S2ZlJ6Q92WJEOepAetDlR
cRFi004bNyTb4Bphg8s4+9Cti5is199aFkGCRDWxeqEnc6aMY3Ezu9Qg3uttLVUd
uy830GUCgYEA7qS0X+9UH1R02L3aoANyADVbFt2ZpUwQGauw9WM92pH52xeHAw1S
ohus6FI3OC8xQq2CN525tGLUbFDZnNZ3YQHqFsfgevfnTs1//gbKXomitev0oFKh
VT+WYS4lkgYtPlXzhdGuk32q99T/wIocAguvCUY3PiA7yBz93ReyausCgYEA6+P8
bugMqT8qjoiz1q/YCfxsw9bAGWjlVqme2xmp256AKtxvCf1BPsToAaJU3nFi3vkw
ICLxUWAYoMBODJ3YnbOsIZOavdXZwYHv54JqwqFealC3DG0Du6fZYZdiY8pK+E6m
3fiYzP1WoVK5tU4bH8ibuIQvpcI8j7Gy0cV6/AMCgYBHl7fZNAZro72uLD7DVGVF
9LvP/0kR0uDdoqli5JPw12w6szM40i1hHqZfyBJy042WsFDpeHL2z9Nkb1jpeVm1
C4r7rJkGqwqElJf6UHUzqVzb8N6hnkhyN7JYkyyIQzwdgFGfaslRzBiXYxoa3BQM
9Q5c3OjDxY3JuhDa3DoVYwKBgDNqrWJLSD832oHZAEIicBe1IswJKjQfriWWsV6W
mHSbdtpg0/88aZVR/DQm+xLFakSp0jifBTS0momngRu06Dtvp2xmLQuF6oIIXY97
2ON1owvPbibSOEcWDgb8pWCU/oRjOHIXts6vxctCKeKAFN93raGphm0+Ck9T72NU
BTubAoGBAMEhI/Wy9wAETuXwN84AhmPdQsyCyp37YKt2ZKaqu37x9v2iL8JTbPEz
pdBzkA2Gc0Wdb6ekIzRrTsJQl+c/0m9byFHsRsxXW2HnezfOFX1H4qAmF6KWP0ub
M8aIn6Rab4sNPSrvKGrU6rFpv/6M33eegzldVnV9ku6uPJI1fFTC
-----END RSA PRIVATE KEY-----`)

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}
