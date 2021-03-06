# ProxyReader

ProxyReader acts as a normal http/https proxy, and save the mitm data to file(s).

For now, it only effects on specified url.

## Usage

Generate a Self-Signed Certification. And Install the Certification on your device.

```bash
	$ openssl genrsa -out ca.key
	$ openssl req -x509 -new -nodes -key ca.key -sha256 -days 730 -out ca.crt
```

Build and Install (GoPath Mode)

```bash
	$ go build -v github.com/zhanglongx/ProxyReader
	$ sudo go install
```

Help

```bash
	$ ProxyReader --help
	Usage of ProxyReader:
	-c string
			CA crt filepath (default "cert/ca.crt")
	-k string
			CA key filepath (default "cert/ca.key")
	-l string
			on which address should the proxy listen (default ":2081")
	-p string
			path to save kw (default "db")
	-v    should every proxy request be logged to stdout
```

Normally, use ProxyReader as Systemd Service

```ini
	[Unit]
	Description=Proxy Reader
	After=network.target

	[Install]
	WantedBy=multi-user.target

	[Service]
	Type=simple
	ExecStart=/usr/local/bin/ProxyReader -c /opt/ca.crt -k /opt/ca.key -p /opt/kw
	ExecReload=/bin/kill -HUP $MAINPID
	Restart=on-failure
```

## Thanks

[elazarl/goproxy](https://github.com/elazarl/goproxy)

## Author

zhanglongx@gmail.com
