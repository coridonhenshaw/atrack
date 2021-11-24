# Atrack

Atrack is a dynamic DNS update receiver that executes predefined commands in response to dhclient-style dynamic IP DNS updates sent over HTTP[S].

The intended use case is to relay the external IP of a residential/SOHO CPE/router to services inside the LAN.

## Security

Atrack is intended to be secure against remote code execution attacks (barring vulnerabilities in the Golang runtime) but is not designed to defend against all possible attack vectors. As such:

**Do not expose Atrack to the open Internet directly or via a reverse proxy.**

If it is necessary to communicate with an Atrack instance over the Internet, use a VPN.

## Command line

`atrack`

No command line options are accepted. The process does not daemonize. Use a service manager (such as systemd) if background operation is desired.

A PID file will be created.

In response to SIGHUP, Atrack will reload access credentials and execution commands from the configuration file.

## Configuration

All configuration is stored in `atrack.xml`. Example:

```
<?xml version="1.0" encoding="UTF-8"?>
<Atrack>
  <ListenAddress Value="172.16.0.10:8003"/>
  <Credentials UserID="testuser" Password="testpassword"/>
  <IPv4Commands Exec="foo $IPv4" Timeout = "1000"/>
  <IPv6Commands Exec="bar $IPv6"/>
<!-- TLSCredentials is optional -->
  <TLSCredentials Fullchain="/srv/cert" PrivateKey="/srv/key"/>
<!-- PIDFile is optional -->
  <PIDFile Value="atrack.pid"/>
</Atrack>
```

`<ListenAddress Value="172.16.0.10:8003"/>`

Listen on port 8003 on address 172.16.0.10. Do not use an address that is publicly reachable.

`<Credentials UserID="testuser" Password="testpassword"/>`

UserID and password credentials to require from the client.

`<IPv4Commands Exec="foo $IPv4" Timeout = "1000"/>`

Command to execute when the client indicates that its IPv4 address has changed. The field `$IPv4` will be replaced with the IPv4 address provided by the client. The `Timeout` field, if specified, directs Atrack to terminate the command if it runs for more than the specified number of seconds.

If multiple IPv4Commands entries are provided, each command will be executed in the order established in the configuration file.

`<IPv6Commands Exec="bar $IPv6"/>`

As with `IPv4Commands`, but for IPv6 addresses.

`<TLSCredentials Fullchain = "/srv/cert" PrivateKey = "/srv/key" />`

Specifies the filenames for the certificate chain and private key to enable HTTPS TLS. Omit `TLSCredentials` to disable TLS.

Unencrypted HTTP is disabled if TLS credentials are provided.

The intended use of this option is to interoperate with DDNS clients that require HTTPS connections. It is not intended to provide any security.

`<PIDFile Value="atrack.pid"/>`

Specifies a PID file to control Atrack while it is running. If this parameter is not specified, `atrack.pid` will be created in the current working directory.

## Network API Endpoints

### `/GetIP`

In response to HTTP GET, this API returns a two-line response consisting of the external IP IPv4 address on line one and the and the external IPv6 address on line two. If an address is not known, `<nil>` will be returned for the relevant line.

*Atrack uses no persistent storage and, after startup, will return `<nil>` for both addresses until a client has called the Update endpoint.*

#### Example:

```
> curl http://172.16.0.10:8003/GetIP
198.51.100.1
<nil>
```

### `/Update`

In response to HTTP GET, this API updates the current IPv4 or IPv6 address retained by Atrack.

#### Parameters:

`IPv4=<IPv4 address>` \
`IPv6=<IPv6 address>`

Updates the stored IPv4 or IPv6 addresses, respectively. Both parameters must not be specified in the same API call; to update both IPv4 and IPv6 addresses, make one API call for each address family.

`UserID=<Userid>` \
`Password=<Password>`

Provides authentication credentials. Both parameters are required.


#### Example:

```
> curl http://172.16.0.10:8003/Update?IPv4=198.51.100.1&UserID=testuser&Password=testpassword
```

Updates the stored IPv4 address to 198.51.100.1 using the UserID `testuser` and the password `testpassword`.

## Bugs

Probably.

## Release History

#### 0

Initial Release

## License

Copyright 2021 Coridon Henshaw

Permission is granted to all natural persons to execute, distribute, and/or modify this software (including its documentation) subject to the following terms:

1. Subject to point \#2, below, **all commercial use and distribution is prohibited.** This software has been released for personal and academic use for the betterment of society through any purpose that does not create income or revenue. *It has not been made available for businesses to profit from unpaid labor.*

2. Re-distribution of this software on for-profit, public use, repository hosting sites (for example: Github) is permitted provided no fees are charged specifically to access this software.

3. **This software is provided on an as-is basis and may only be used at your own risk.** This software is the product of a single individual's recreational project. The author does not have the resources to perform the degree of code review, testing, or other verification required to extend any assurances that this software is suitable for any purpose, or to offer any assurances that it is safe to execute without causing data loss or other damage.

4. **This software is intended for experimental use in situations where data loss (or any other undesired behavior) will not cause unacceptable harm.** Users with critical data safety needs must not use this software and, instead, should use equivalent tools that have a proven track record.

5. If this software is redistributed, this copyright notice and license text must be included without modification.

6. Distribution of modified copies of this software is discouraged but is not prohibited. It is strongly encouraged that fixes, modifications, and additions be submitted for inclusion into the main release rather than distributed independently.

7. This software reverts to the public domain 10 years after its final update or immediately upon the death of its author, whichever happens first.
