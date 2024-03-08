package sheets

import (
	"calendar/googlesvc/dto"
	"calendar/sentiment"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService struct {
	client            *http.Client
	sentimentAnalyzer sentiment.SentimentAnalyzer
}

func New(
	client *http.Client,
) *GoogleSheetsService {
	return &GoogleSheetsService{
		client:            client,
		sentimentAnalyzer: sentiment.NewGoVader(),
	}
}

func (s *GoogleSheetsService) CreateSheet(
	ctx context.Context,
	events []dto.Event,
) (
	*sheets.Spreadsheet,
	error,
) {
	// Create Google Sheets service
	svc, err := s.createSheetsService(ctx)
	if err != nil {
		return nil, err
	}

	// Call the Sheets API to create the sheet
	spreadsheet, err := s.createSheet(svc)
	if err != nil {
		return nil, err
	}

	// Call the Sheets API to create the sheet
	err = s.addEventDataToSheet(svc, spreadsheet, events)
	if err != nil {
		return nil, err
	}

	return spreadsheet, nil
}

func (s *GoogleSheetsService) createSheetsService(
	ctx context.Context,
) (
	*sheets.Service,
	error,
) {
	// Create Sheets service
	svc, err := sheets.NewService(ctx, option.WithHTTPClient(s.client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		return nil, err
	}
	return svc, nil
}

func (s *GoogleSheetsService) createSheet(
	svc *sheets.Service,
) (
	*sheets.Spreadsheet,
	error,
) {
	// Create a new Google Sheet
	sheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: fmt.Sprintf("New Spreadsheet - %s", time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	// Call the Sheets API to create the sheet
	spreadsheet, err := svc.Spreadsheets.Create(sheet).Do()
	if err != nil {
		log.Fatalf("Unable to create spreadsheet: %v", err)
	}

	fmt.Println(spreadsheet.SpreadsheetId)
	fmt.Println(spreadsheet.SpreadsheetUrl)

	return spreadsheet, nil
}

func (s *GoogleSheetsService) addEventDataToSheet(
	svc *sheets.Service,
	spreadsheet *sheets.Spreadsheet,
	events []dto.Event,
) error {

	// Add data to the sheet
	var values [][]interface{}
	values = append(values, []interface{}{
		"StartDate",
		"EndDate",
		"Persona",
		"Action",
		"Topic",
		"Details",
		"Attendees",
		"Sentiment",
	})

	for _, event := range events {

		// Format attendees
		attendees := ""
		for _, attendee := range event.Attendees {
			attendees = attendees +
				fmt.Sprintf("- %s (%s) [%s]\n", attendee.Email, attendee.Name, attendee.Response)
		}

		// Run sentiment analysis
		sentiment := s.sentimentAnalyzer.Run(event.Details)

		// Create row
		values = append(values, []interface{}{
			event.StartDate,
			event.EndDate,
			event.Persona,
			event.Action,
			event.Topic,
			event.Details,
			attendees,
			sentiment,
		})
	}

	// Specify range for writing data
	rangeData := fmt.Sprintf("A1:H%d", len(events)+1)

	// Write data to the sheet
	_, err := svc.Spreadsheets.Values.Update(spreadsheet.SpreadsheetId, rangeData,
		&sheets.ValueRange{
			Values: values,
		}).ValueInputOption("RAW").Do()
	if err != nil {
		fmt.Println("Unable to write data to sheet: " + err.Error())
		return err
	}

	return nil
}
