package agent

import (
	"path"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	nsdconfig "isc.org/stork/appcfg/nsd"
	storkutil "isc.org/stork/util"
)

var _ App = (*Bind9App)(nil)

// Represents the NSD process metadata.
type NsdDaemon struct {
	Pid     int32
	Name    string
	Version string
	Active  bool
}

// Represents the state of BIND 9.
type NsdState struct {
	Version string
	Active  bool
	Daemon  NsdDaemon
}

// It holds common and NSD specific runtime information.
type NsdApp struct {
	BaseApp
	nsdClient     *nsdClient
	zoneInventory *zoneInventory
}

// Get base information about NSD app.
func (ba *NsdApp) GetBaseApp() *BaseApp {
	return &ba.BaseApp
}

// Detect allowed logs provided by NSD.
// TODO: currently it is not implemented and not used,
// it returns always empty list and no error.
func (ba *NsdApp) DetectAllowedLogs() ([]string, error) {
	return nil, nil
}

// Waits for the zone inventory to complete background tasks.
func (ba *NsdApp) AwaitBackgroundTasks() {
	if ba.zoneInventory != nil {
		ba.zoneInventory.awaitBackgroundTasks()
	}
}

// List of BIND 9 executables used during app detection.
const (
	nsdControlExec = "nsd-control"
	nsdExec        = "nsd"
)

func detectNSDApp(match []string, cwd string, ce storkutil.CommandExecutor) App {
	if len(match) < 3 {
		log.Warnf("Problem with parsing nsd cmdline: %s", match[0])
		return nil
	}

	nsdParams := match[2]
	rootPrefix := ""

	// Look for the chroot directory.
	chrootPathPattern := regexp.MustCompile(`--chroot\s+(\S+)`)
	m := chrootPathPattern.FindStringSubmatch(nsdParams)
	if m != nil {
		rootPrefix = strings.TrimRight(m[1], "/")

		// The cwd path is already prefixed with the chroot directory
		// because the /proc/(pid)/cwd is absolute.
		cwd = strings.TrimPrefix(cwd, rootPrefix)
	}

	configDir := "."
	configName := "nsd.conf"
	paramsSlice := strings.Split(nsdParams, " ")
	for _, param := range paramsSlice {
		if strings.HasPrefix(param, "-c") {
			split := strings.Split(param, " ")
			if len(split) > 1 {
				configDir = split[1]
			}
			continue
		}
	}

	configPath := path.Join(configDir, configName)

	// If path to config is not absolute then join it with CWD of nsd.
	if !path.IsAbs(configPath) {
		configPath = path.Join(cwd, configPath)
	}

	log.WithFields(log.Fields{
		"configPath": configPath,
	}).Info("Using the following nsd config path")

	parser := nsdconfig.NewParser(configPath)
	parsedConfig, err := parser.Parse()
	if err != nil {
		return nil
	}

	controlInterface := ""
	if address, ok := parsedConfig["control-interface"]; ok {
		controlInterface = *address.StringValue
	}

	accessPoints := []AccessPoint{
		{
			Type:    AccessPointControl,
			Address: controlInterface,
		},
	}

	nsdClient := NewNSDClient(ce)

	inventory := newZoneInventory(newZoneInventoryStorageMemory(), nsdClient, "localhost", int64(0))

	nsdApp := &NsdApp{
		BaseApp: BaseApp{
			Type:         AppTypeNSD,
			AccessPoints: accessPoints,
		},
		nsdClient:     nsdClient,
		zoneInventory: inventory,
	}

	return nsdApp
}

// Send a command to named using rndc client.
func (na *NsdApp) sendCommand(command []string) (output []byte, err error) {
	return na.nsdClient.SendCommand(command)
}
