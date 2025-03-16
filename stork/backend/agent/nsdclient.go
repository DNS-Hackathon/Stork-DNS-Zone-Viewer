package agent

import (
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"isc.org/stork/appdata/bind9stats"
	storkutil "isc.org/stork/util"
)

type NSDControlZones struct {
	// The configuration contains a list of Statements separated by semicolons.
	Zones []*NSDControlZone `parser:"( @@ )*"`
}

type NSDControlZone struct {
	ZoneName       string           `parser:"'zone' Colon @Ident"`
	ZoneStatements []*ZoneStatement `parser:"( @@ )*"`
}

func (z *NSDControlZone) GetState() string {
	for _, statement := range z.ZoneStatements {
		if statement.State != "" {
			return statement.State
		}
	}
	return ""
}

func (z *NSDControlZone) GetServedSerial() int64 {
	for _, statement := range z.ZoneStatements {
		if statement.ServedSerial != "" {
			parts := strings.Split(statement.ServedSerial, " ")
			if len(parts) > 0 {
				convertedSerial, err := strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					return 0
				}
				return convertedSerial
			}
		}
	}
	return 0
}

type ZoneStatement struct {
	State        string `parser:"( 'state' Colon @Ident )"`
	ServedSerial string `parser:"| ( 'served-serial' Colon ( @String ) )"`
	CommitSerial string `parser:"| ( 'commit-serial' Colon ( @String ) )"`
	Wait         string `parser:"| ( 'wait' Colon ( @String ) )"`
}

func Parse(input string) (*NSDControlZones, error) {
	// Define the custom lexer.
	lexer := lexer.MustSimple([]lexer.SimpleRule{
		{Name: "String", Pattern: `"(\\"|[^"])*"`},
		{Name: "Ident", Pattern: `[0-9a-zA-Z-\.]+`},
		{Name: "Colon", Pattern: `[:]`},
		//		{Name: "Punct", Pattern: `[:]`},
		{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
		{Name: "EOL", Pattern: `[\n\r]+`},
	})

	parser := participle.MustBuild[NSDControlZones](
		participle.Lexer(lexer),
		participle.Unquote("String"),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2),
	)
	config, err := parser.Parse("", strings.NewReader(input))
	if err != nil {
		return nil, err
	}
	return config, nil
}

// nsd-control client
type nsdClient struct {
	executor    storkutil.CommandExecutor
	BaseCommand []string
}

// Instantiates REST client for PowerDNS.
func NewNSDClient(ce storkutil.CommandExecutor) *nsdClient {
	nsdClient := &nsdClient{
		executor: ce,
	}
	return nsdClient
}

// Send command to named using rndc executable.
func (rc *nsdClient) SendCommand(command []string) (output []byte, err error) {
	var nsdCommand []string
	nsdCommand = append(nsdCommand, rc.BaseCommand...)
	nsdCommand = append(nsdCommand, command...)
	log.Debugf("nsd-control: %+v", nsdCommand)

	if len(nsdCommand) == 0 {
		return nil, errors.New("no nsd-control command specified")
	}

	return rc.executor.Output(nsdCommand[0], nsdCommand[1:]...)
}

// Makes a request to retrieve BIND9 views over the stats channel.
/* func (request *pdnsClientRequest) getViews() (httpResponse, *bind9stats.Views, error) {
	// The /zones path returns the top level stats structure. Besides the
	// map of views it returns other top-level information. We need to embed
	// the Views field in the structure to fit the returned data. Next
	// we will extract the views map from it.
	var result struct {
		Zones *pdnsdata.Zones
	}
	response, err := request.getJSON("/servers/localhost/zones", &result.Zones)
	if err != nil {
		return nil, nil, err
	}

	var bind9Zones []*bind9stats.Zone
	for zone := range result.Zones.GetIterator() {
		bind9Zone := &bind9stats.Zone{
			ZoneName: zone.Name(),
			Class:    "IN",
			Serial:   zone.Serial,
			Type:     zone.Kind,
			Loaded:   time.Now(),
		}
		bind9Zones = append(bind9Zones, bind9Zone)
	}
	view := bind9stats.NewView("localhost", bind9Zones)
	views := bind9stats.NewViews([]*bind9stats.View{view})

	// Extract the views and drop other top-level information.
	return response, views, err
} */

func (client *nsdClient) getViews(host string, port int64) (httpResponse, *bind9stats.Views, error) {
	output, err := client.SendCommand([]string{"/opt/homebrew/sbin/nsd-control", "zonestatus"})
	if err != nil {
		return nil, nil, err
	}

	parsed, err := Parse(string(output))
	if err != nil {
		log.Errorf("error parsing nsd-control zonestatus output: %s", err)
		return nil, nil, err
	}
	zones := []*bind9stats.Zone{}
	for _, zone := range parsed.Zones {
		zones = append(zones, &bind9stats.Zone{
			ZoneName: zone.ZoneName,
			Class:    "IN",
			Type:     zone.GetState(),
			Serial:   zone.GetServedSerial(),
			Loaded:   time.Now(),
		})
	}

	views := bind9stats.NewViews([]*bind9stats.View{bind9stats.NewView("_default", zones)})
	return nil, views, err
}
