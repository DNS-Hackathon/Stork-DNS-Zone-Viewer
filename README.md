# Stork DNS Zone Viewer

## Background
Stork (https://stork.isc.org) is an open-source project implementing a monitoring and management system for Kea DHCP servers and basic integration with BIND9 DNS servers. Lately, ISC has been working on the zone viewer for BIND9 servers. We have added an efficient mechanism to detect and fetch the configured zones from the monitored BIND9 servers into the central database. Stork server makes them available over the REST API so they can be listed, filtered, and viewed in the GUI.

Many administrators integrate DNS servers from different vendors in a single installation. It protects against potential issues specific to a selected DNS implementation, provides redundancy, and increases performance and flexibility.

## The Hackathon Project

During the hackathon, we'd like to extend the zone viewer to work with another DNS implementation (i.e., PowerDNS). We'd like to create a proof of concept that the Stork project can be used in environments with diverse DNS implementations and identify any roadblocks and possible portability issues.

## Tasks

The following is the preliminary list of tasks for integrating PowerDNS with Stork.

### Detect a Running PowerDNS Instance
Stork agent is a program running on each monitored machine. It needs to be extended to detect a PowerDNS instance and report it back to the Stork server, which maintains a central database of the monitored services. We need to take into account that there might be different ways to launch the DNS servers, and different access privileges may be required. A usual way to detect a running application is to list processes and match executable names against known naming patterns. Once the DNS server instance is detected, the Stork agent must locate appropriate configuration files for this server.

### Fetch Configuration from PowerDNS
A usual way to fetch configuration from a server is either by looking into its configuration file (if the file location is known) or by using its configuration API. The latter is usually impossible without first looking into the configuration. It may require some basic configuration parser to locate the meaningful configuration bits.

### Save Information about Detected Instances in the Central Database
Once the Stork agent has enough knowledge about the running DNS instance, it must communicate it back to the Stork server, which subsequently stores this information in the database. It may require some new data structures to represent PowerDNS-specific state information.

### Fetch Zones from the DNS Servers into Zone Inventory
The Stork agent contains a mechanism called "zone inventory", which holds the information about the zones configured in the DNS server. In order to fetch the information from the DNS server and put it into the zone inventory, we need to use a REST API of the monitored DNS server, or AXFR zone transfer, assuming it is available to the Stork agent. The former may be easier to use because it doesn't require implementing a DNS protocol in Stork. The latter, however, can be more portable and work for all DNS implementations. A goal here is to evaluate both solutions.

### Extensions to the Zone Viewer in the UI
In order to present the information specific to a new DNS implementation some extensions to the Stork UI will be required.

### Document the Solution and the Roadblocks
The solution implemented during the hackathon will be documented, and the specific issues will be highlighted, with some possible recommendations on how they can be addressed in the future. 

## Expected Outcome
We'd like to come up with best practices how to integrate new DNS solutions into Stork. While the conclusions will be primarily helpful in the Stork development, they should include general observations about the portability of different DNS implementations. We also want to explore how relying on the DNS protocols (AXFR, IXFR, DNS NOTIFY) can improve the portability of the DNS servers for integration with the monitoring solutions.


