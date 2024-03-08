package main

import (
	gauth "calendar/googlesvc/auth"
	gcalendar "calendar/googlesvc/calendar"
	gdrive "calendar/googlesvc/drive"
	gsheets "calendar/googlesvc/sheets"
	"context"
)

func main() {

	// Get context
	ctx := context.Background()

	// Authenticate to Google
	authSvc := gauth.New()
	client, err := authSvc.CreateClient()
	if err != nil {
		return
	}

	// Fetch Google Calendar events
	calendarSvc := gcalendar.New(client)
	events, err := calendarSvc.FetchEvents(ctx)
	if err != nil {
		return
	}

	// Create Google Sheet with fetched events
	sheetsSvc := gsheets.New(client)
	spreadsheet, err := sheetsSvc.CreateSheet(ctx, events)
	if err != nil {
		return
	}

	// Move Google Sheet to specific folder
	driveSvc := gdrive.New(client)
	err = driveSvc.MoveFile(ctx, spreadsheet.SpreadsheetId)
	if err != nil {
		return
	}
}
