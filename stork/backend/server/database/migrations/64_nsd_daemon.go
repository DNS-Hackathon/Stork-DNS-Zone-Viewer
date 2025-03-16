package dbmigs

import "github.com/go-pg/migrations/v8"

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		_, err := db.Exec(`
			-- This table holds NSD daemon-specific information.
			CREATE TABLE IF NOT EXISTS nsd_daemon (
			id BIGSERIAL NOT NULL,
			daemon_id BIGINT NOT NULL,
			CONSTRAINT nsd_daemon_pkey PRIMARY KEY (id),
			CONSTRAINT nsd_daemon_id_unique UNIQUE (daemon_id),
			CONSTRAINT nsd_daemon_id_fkey FOREIGN KEY (daemon_id)
				REFERENCES daemon (id) MATCH SIMPLE
					ON UPDATE CASCADE
					ON DELETE CASCADE
			);
		`)
		return err
	}, func(db migrations.DB) error {
		_, err := db.Exec(`
				DROP TABLE IF EXISTS nsd_daemon;
			`)
		return err
	})
}
