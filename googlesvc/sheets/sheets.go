package sheets

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService struct {
	client *http.Client
}

func New(
	client *http.Client,
) *GoogleSheetsService {
	return &GoogleSheetsService{
		client: client,
	}
}

func (s *GoogleSheetsService) CreateSheet(
	ctx context.Context,
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

	return sheet, nil
}
