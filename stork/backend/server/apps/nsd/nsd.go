package nsd

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	log "github.com/sirupsen/logrus"
	nsdconfig "isc.org/stork/appcfg/nsd"
	"isc.org/stork/server/agentcomm"
	dbmodel "isc.org/stork/server/database/model"
	"isc.org/stork/server/eventcenter"
)

type ServerResponse struct {
	Version    string `json:"version"`
	DaemonType string `json:"daemon_type"`
}

func GetAppState(ctx context.Context, agents agentcomm.ConnectedAgents, dbApp *dbmodel.App, eventCenter eventcenter.EventCenter) {
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	output, err := agents.ForwardToNSD(ctx2, dbApp, "status")
	if err != nil {
		log.Warnf("Problem getting NSD server status: %s", err)
		return
	}

	daemon := dbmodel.NewNSDDaemon(true)
	if len(dbApp.Daemons) > 0 && dbApp.Daemons[0].ID != 0 {
		daemon = dbApp.Daemons[0]
	}

	daemon.Version, err = nsdconfig.ParseNSDVersion(output.Output)
	if err != nil {
		log.Warnf("Problem parsing NSD version: %s", err)
		return
	}

	log.Infof("NSD server status: %s", output.Output)

	// Save status
	dbApp.Active = daemon.Active
	dbApp.Meta.Version = daemon.Version
	dbApp.Daemons = []*dbmodel.Daemon{
		daemon,
	}
}

// Inserts or updates information about PowerDNS app in the database.
func CommitAppIntoDB(db *pg.DB, app *dbmodel.App, eventCenter eventcenter.EventCenter) (err error) {
	if app.ID == 0 {
		_, err = dbmodel.AddApp(db, app)
		eventCenter.AddInfoEvent("added {app}", app.Machine, app)
	} else {
		_, _, err = dbmodel.UpdateApp(db, app)
	}
	// todo: perform any additional actions required after storing the
	// app in the db.
	return err
}
