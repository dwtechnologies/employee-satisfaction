package main

import (
	"database/sql"
	"fmt"
)

// save will safe the query to the redshift table.
// Returns error.
func (srv *server) save() error {
	s, err := sql.Open("postgres", srv.redshiftConfig())
	if err != nil {
		return err
	}

	if _, err := s.Query(srv.query); err != nil {
		return err
	}

	return nil
}

// redshiftConfig will return a config string for the connection.
func (*server) redshiftConfig() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?port=%s&connect_timeout=15", user, pass, host, db, port)
}
