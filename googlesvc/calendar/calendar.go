package calendar

import (
	"calendar/googlesvc/dto"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GoogleCalendarService struct {
	client *http.Client
}

func New(
	client *http.Client,
) *GoogleCalendarService {
	return &GoogleCalendarService{
		client: client,
	}
}

// Fetch necessary Google Calendar events
func (s *GoogleCalendarService) FetchEvents(
	ctx context.Context,
) (
	[]dto.Event,
	error,
) {

	svc, err := s.createCalendarService(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch calendar allEvents
	allEvents, err := s.fetchEvents(svc)
	if err != nil {
		return nil, err
	}

	// Parse events and filter the necessary ones
	filteredEvents := s.parseAndFilterEvents(allEvents)

	// Write the filtered events into JSON
	err = writeToJsonFile(filteredEvents)
	if err != nil {
		return nil, err
	}

	return filteredEvents, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func (s *GoogleCalendarService) createCalendarService(
	ctx context.Context,
) (
	*calendar.Service,
	error,
) {
	// Create Google Calendar service
	svc, err := calendar.NewService(ctx, option.WithHTTPClient(s.client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
		return nil, err
	}
	return svc, nil
}

// Fetch events from Google Calendar
func (s *GoogleCalendarService) fetchEvents(
	srv *calendar.Service,
) (
	[]*calendar.Event,
	error,
) {
	// Define the time range for the last week
	now := time.Now().UTC()
	weekAgo := now.AddDate(0, 0, -7)
	timeMin := weekAgo.Format(time.RFC3339)
	timeMax := now.Format(time.RFC3339)

	// Fetch first pagination of initialEvents
	initialEvents, err := srv.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		TimeMax(timeMax).
		MaxResults(20).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}

	var allEvents []*calendar.Event
	allEvents = append(allEvents, initialEvents.Items...)

	fmt.Println("Upcoming events:")
	if len(initialEvents.Items) == 0 {
		fmt.Println("No upcoming events found.")
	}

	// Continue paginating until all events are fetched
	for initialEvents.NextPageToken != "" {
		pageToken := initialEvents.NextPageToken
		initialEvents, err = srv.Events.List("primary").
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(timeMin).
			TimeMax(timeMax).
			MaxResults(20).
			OrderBy("startTime").
			PageToken(pageToken).
			Do()
		if err != nil {
			fmt.Println("failed to retrieve events: " + err.Error())
			return nil, err
		}
		allEvents = append(allEvents, initialEvents.Items...)
	}

	return allEvents, nil
}

// Parse & filter necessary calendar events
func (s *GoogleCalendarService) parseAndFilterEvents(
	allEvents []*calendar.Event,
) []dto.Event {

	var filteredEvents []dto.Event

	for _, events := range allEvents {

		// Event time
		startDate := events.Start.DateTime
		if startDate == "" {
			startDate = events.Start.Date
		}
		endDate := events.End.DateTime
		if endDate == "" {
			endDate = events.End.Date
		}

		if strings.HasPrefix(events.Summary, "!!!") {

			// Fundamental information
			title := events.Summary[3:]
			items := strings.Split(title, ":")
			persona := items[0]
			action := items[1]
			topic := items[2]
			details := events.Description

			// Attendees
			attendees := []dto.Attendee{}
			for _, attendee := range events.Attendees {
				attendees = append(attendees, dto.Attendee{
					Name:     attendee.DisplayName,
					Email:    attendee.Email,
					Response: attendee.ResponseStatus,
				})
			}

			filteredEvent := dto.Event{
				StartDate: startDate,
				EndDate:   endDate,
				Persona:   persona,
				Action:    action,
				Topic:     topic,
				Details:   details,
				Attendees: attendees,
			}

			filteredEvents = append(filteredEvents, filteredEvent)
		}
	}

	// TODO: Sort events by startDate
	return filteredEvents
}

// Write results into a CSV file
func writeToJsonFile(
	data []dto.Event,
) error {
	// Open or create a JSON file for writing
	file, err := os.Create("report.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	// Encode the data and write it to the file
	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return err
	}

	fmt.Println("Data written successfully.")
	return nil
}
