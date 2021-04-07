package main

import "testing"

// TestUsage just calls the usage-function for each of our handlers,
// it achieves no useful testing.
func TestUsage(t *testing.T) {

	add := addCmd{}
	add.Info()

	cron := cronCmd{}
	cron.Info()

	config := configCmd{}
	config.Info()

	daemon := daemonCmd{}
	daemon.Info()

	del := delCmd{}
	del.Info()

	export := exportCmd{}
	export.Info()

	imprt := importCmd{}
	imprt.Info()

	list := listCmd{}
	list.Info()

	ldt := listDefaultTemplateCmd{}
	ldt.Info()

	vers := versionCmd{}
	vers.Info()

}
