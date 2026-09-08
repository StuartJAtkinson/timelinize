/*
	Timelinize
	Copyright (c) 2013 Matthew Holt

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published
	by the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"embed"

	tlcmd "github.com/StuartJAtkinson/timelinize/cmd"
	// plug in data sources
	_ "github.com/StuartJAtkinson/timelinize/datasources/applecontacts"
	_ "github.com/StuartJAtkinson/timelinize/datasources/applephotos"
	_ "github.com/StuartJAtkinson/timelinize/datasources/calendar"
	_ "github.com/StuartJAtkinson/timelinize/datasources/contactlist"
	_ "github.com/StuartJAtkinson/timelinize/datasources/element"
	_ "github.com/StuartJAtkinson/timelinize/datasources/email"
	_ "github.com/StuartJAtkinson/timelinize/datasources/facebook"
	_ "github.com/StuartJAtkinson/timelinize/datasources/firefox"
	_ "github.com/StuartJAtkinson/timelinize/datasources/flighty"
	_ "github.com/StuartJAtkinson/timelinize/datasources/generic"
	_ "github.com/StuartJAtkinson/timelinize/datasources/geojson"
	_ "github.com/StuartJAtkinson/timelinize/datasources/github"
	_ "github.com/StuartJAtkinson/timelinize/datasources/goodreads"
	_ "github.com/StuartJAtkinson/timelinize/datasources/googlelocation"
	_ "github.com/StuartJAtkinson/timelinize/datasources/googlephotos"
	_ "github.com/StuartJAtkinson/timelinize/datasources/googlevoice"
	_ "github.com/StuartJAtkinson/timelinize/datasources/gpx"
	_ "github.com/StuartJAtkinson/timelinize/datasources/icloud"
	_ "github.com/StuartJAtkinson/timelinize/datasources/instagram"
	_ "github.com/StuartJAtkinson/timelinize/datasources/iphone"
	_ "github.com/StuartJAtkinson/timelinize/datasources/kmlgx"
	_ "github.com/StuartJAtkinson/timelinize/datasources/line"
	_ "github.com/StuartJAtkinson/timelinize/datasources/media"
	_ "github.com/StuartJAtkinson/timelinize/datasources/nmea"
	_ "github.com/StuartJAtkinson/timelinize/datasources/smsbackuprestore"
	_ "github.com/StuartJAtkinson/timelinize/datasources/strava"
	_ "github.com/StuartJAtkinson/timelinize/datasources/telegram"
	_ "github.com/StuartJAtkinson/timelinize/datasources/twitter"
	_ "github.com/StuartJAtkinson/timelinize/datasources/vcard"
	_ "github.com/StuartJAtkinson/timelinize/datasources/whatsapp"
)

// Package main is the entry point of the application.
func main() {
	tlcmd.Main(embeddedWebsite)
}

// The frontend assets. This will always be embedded into the binary
// even if the config or command line flag says to use a folder on
// disk (usually for dev).
//
// This is only defined here because goembed can't embed something
// from a parent directory, and I just don't care to have the folder
// nested within another folder right now.
//
//go:embed all:frontend
var embeddedWebsite embed.FS
