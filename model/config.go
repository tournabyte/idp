/*
 * package model describes the data types utilized by the idp service
 */

package model

type CommandOpts struct {
	Port    int
	Dbhosts []string
	Dbname  string
	Dbuser  string
	Dbpass  string
}
