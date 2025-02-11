# Stork DNS Zone Viewer

## Background
Stork (https://stork.isc.org) is an open-source project implementing a monitoring and management system for Kea DHCP servers and basic integration with BIND9 DNS servers. Lately, ISC has been working on the zone viewer for BIND9 servers. We have added an efficient mechanism to detect and fetch the configured zones from the monitored BIND9 servers into the central database. Stork server makes them available over the REST API so they can be listed, filtered, and viewed in the GUI.

Many administrators integrate DNS servers from different vendors in a single installation. It protects against potential issues specific to a selected DNS implementation, provides redundancy, and increases performance and flexibility.

## The Hackathon Project

During the hackathon, we'd like to extend the zone viewer to work with another DNS implementation (i.e., PowerDNS). We'd like to create a proof of concept that the Stork project can be used in environments with diverse DNS implementations and identify any roadblocks and possible portability issues.

## Setup

During the Hackathon we're going to use a development version of the [Stork project](https://gitlab.isc.org/isc-projects/stork). This version is not yet available in the Hackathon repository but will be committed here by the end of February. It will contain a Docker image with the PowerDNS, that will be a part of the [demo](https://gitlab.isc.org/isc-projects/stork/-/wikis/demo) runnable using the Stork build system. The container is mostly useful for testing the working code. For the convenience of the code development during the hackathon, it is desired to install PowerDNS on the developer's machine to ease debugging, capturing the network traffic between Stork and PowerDNS. Please see [Installing PowerDNS](https://doc.powerdns.com/authoritative/installation.html) for instructions how to run and configure PowerDNS on various operating systems.

Suppose you have the following BIND9 configuration file:

```
acl "rndc-users" {
     172.10.17.1;
};

options {
        directory "/Users/marcin/devel/pdns/";
        listen-on { localhost; };
        allow-update {
                127.0.0.1;
        };
        ixfr-from-differences yes;
        provide-ixfr yes;
        notify yes;
        allow-new-zones yes;
};

key "rndc-key" {
        algorithm hmac-sha256;
        secret "CuJ1RN0GKsyGZlvrnFk7qNh8XYathYSNYAwT2nNKAXQ=";
};

controls {
	inet * allow { "rndc-users"; } keys { "rndc-key"; };
};

statistics-channels {
        inet 127.0.0.1 port 8053 allow { 127.0.0.1; };
};


zone "example.com." {
        type primary;
        allow-transfer { any; };
        file "/Users/marcin/devel/pdns/db.example.com";
};

zone "254.168.192.in-addr.arpa." {
        type primary;
        allow-transfer { any; };
        file "/Users/marcin/devel/pdns/192.168.254.rev";
```

This file (`primary.conf`) can be used as a "configuration backend" for PowerDNS using the configuration similar to this:

```
launch=bind
bind-config=/Users/marcin/devel/pdns/primary.conf
webserver=yes
webserver-address=127.0.0.1
api=yes
api-key=changeme
dnsupdate=yes
```

Note that this configuration also contains the parameters necessary to enable REST API in PowerDNS, that the Stork agent will be able to talk to.

Then, the PowerDNS server can be started using the following command line, assuming that the `pdns.conf` file is in a custom location:

```
/opt/homebrew/opt/pdns/sbin/pdns_server --config-dir=pdns/
```

The above command worked on macOS. The binary location is obviously specific to the used OS.
Stork server requires a PostgreSQL server to run. It can be installed using the steps described in the [Stork ARM](https://stork.readthedocs.io/en/v2.0.0/install.html#preparing-the-stork-server-database).

For development purposes, it is convenient to run the Stork server and the agent on the same machine. To run the server:

```
$ cd stork/
$ rake run:server
```

Optionally, a database host can be specified. For example, on macOS:

```
$ rake run:server STORK_DATABASE_HOST=/tmp
```

To run the agent do:

```
$ cd stork/
$ rake run:agent REGISTER=true
```

Note that the `REGISTER` variable instructs the agent to try to register in the Stork server. The registration is required to connect the Stork server and the agent, and thus make the agent fetch the zone information from the agent.

To register the agent, open the brower and type: `localhost:8080`. Login to Stork using the `admin/admin` credentials. Follow the steps to change the password when prompted. Navigate to `Machines -> Unauthorized`. Click the button on the right in the table containing the machine information. Register the agent.

The agent now works next to the PowerDNS server but it neither detects the DNS server nor can communicate with it. The Hackathon task is about integrating them, similar to how BIND9 is integrated.

## Tasks

The following is the preliminary list of tasks for integrating PowerDNS with Stork.

### Detect a Running PowerDNS Instance
Stork agent is a program running on each monitored machine. It needs to be extended to detect a PowerDNS instance and report it back to the Stork server, which maintains a central database of the monitored services. We need to take into account that there might be different ways to launch the DNS servers, and different access privileges may be required. A usual way to detect a running application is to list processes and match executable names against known naming patterns. Once the DNS server instance is detected, the Stork agent must locate appropriate configuration files for this server.

Most of the changes are required in the `backend/agent/monitor.go`. The `detectApps` function already contains the relevant code to Kea and BIND9. It needs to be extended. `backend/agent/bind9.go` and `backend/agent/kea.go` contain the code specific to detecting and parsing configuration of the BIND9 and Kea apps respectively. We will presumably need a new file `backend/agent/pdns.go` for PowerDNS specific functionality.

### Fetch Configuration from PowerDNS
A usual way to fetch configuration from a server is either by looking into its configuration file (if the file location is known) or by using its configuration API. The latter is usually impossible without first looking into the configuration. It may require some basic configuration parser to locate the meaningful configuration bits.

PowerDNS uses fairly simple configuration syntax described [here](https://doc.powerdns.com/authoritative/settings.html). The configuration parser we create can probably just read the config line by line and split the key/value pairs into a map. Then, we should be able to look into the interesting configuration information (e.g., `api-key`, `webserver-address`). This configuration information will be used to setup a REST client.

The best place in the code structure to implement the parser is in `backend/appcfg/pdns`.

### Save Information about Detected Instances in the Central Database
Once the Stork agent has enough knowledge about the running DNS instance, it must communicate it back to the Stork server, which subsequently stores this information in the database. It may require some new data structures to represent PowerDNS-specific state information.

Stork server is using gRPC protocol to communicate with the agent. Stork API contains several calls (see `backend/api/agent.proto`) to communicate with the monitored apps via Stork agent. These calls are `ForwardToKeaOverHTTP`, `ForwardToNamedStats`, `ForwardRndcCommand`. These calls are used by the Stork server to collect information about the monitored servers (e.g., version number). There is a need for similar call for PowerDNS, so the Stork server can fetch the information about it and store in the database. The REST API client implemented in another task described here should be helpful to implement it.

### Fetch Zones from the DNS Servers into Zone Inventory
The Stork agent contains a mechanism called "zone inventory", which holds the information about the zones configured in the DNS server. In order to fetch the information from the DNS server and put it into the zone inventory, we need to use a REST API of the monitored DNS server, or AXFR zone transfer, assuming it is available to the Stork agent. The former may be easier to use because it doesn't require implementing a DNS protocol in Stork. The latter, however, can be more portable and work for all DNS implementations. A goal here is to evaluate both solutions.

The REST API client for PowerDNS can be implemented similar to the existing client for BIND9 statistics (see `backend/agent/bind9statsclient.go`).

### Using AXFR to fetch zone details
If time permits, we should try to use AXFR to fetch the zone data into the zone inventory. This data can be subsequently transferred on demand to the Stork server and saved in its database. The data can be then displayed in the UI. The https://github.com/miekg/dns is a feature-rich library that can be used to perform the zone transfer.

Zone transfer requires permissive configuration of the DNS server, so the agent is allowed to perform this action. There are also other aspects. For example: BIND9 allows for specifying views, which is not a DNS concept. Performing zone transfer from the particular view requires associating the view with a TSIG key that can be used for view selection. We want to see if there are similar or other issues while using zone transfer for PowerDNS and what are the limitations.

### Extensions to the Zone Viewer in the UI
In order to present the information specific to a new DNS implementation some extensions to the Stork UI will be required. If nothing else, we need a new page showing registered PowerDNS app details.

### Document the Solution and the Roadblocks
The solution implemented during the hackathon will be documented, and the specific issues will be highlighted, with some possible recommendations on how they can be addressed in the future. 

## Expected Outcome
We'd like to come up with best practices how to integrate new DNS solutions into Stork. While the conclusions will be primarily helpful in the Stork development, they should include general observations about the portability of different DNS implementations. We also want to explore how relying on the DNS protocols (AXFR, IXFR, DNS NOTIFY) can improve the portability of the DNS servers for integration with the monitoring solutions.


