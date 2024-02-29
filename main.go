package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {

	// Get context
	ctx := context.Background()

	// Prepare OAuth config for Google
	config, err := prepareConfig()
	if err != nil {
		return
	}

	// Create Google Calendar service
	srv, err := createCalendarService(ctx, config)
	if err != nil {
		return
	}

	// Fetch calendar allEvents
	allEvents, err := fetchEvents(srv)
	if err != nil {
		return
	}

	// Parse events and filter the necessary ones
	filteredEvents := parseAndFilterEvents(allEvents)

	// Write the filtered events into CSV
	err = writeToCsvFile(filteredEvents)
	if err != nil {
		return
	}
}

// Prepare config
func prepareConfig() (
	*oauth2.Config,
	error,
) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		return nil, err
	}

	return config, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func createCalendarService(
	ctx context.Context,
	config *oauth2.Config,
) (
	*calendar.Service,
	error,
) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	client := config.Client(context.Background(), tok)

	// Create Google Calendar service
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
		return nil, err
	}
	return srv, nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Fetch events from Google Calendar
func fetchEvents(
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
func parseAndFilterEvents(
	allEvents []*calendar.Event,
) [][]string {

	var parsedEvents [][]string
	columns := []string{"Persona", "Action", "Topic", "Details"}
	parsedEvents = append(parsedEvents, columns)

	for _, events := range allEvents {
		date := events.Start.DateTime
		if date == "" {
			date = events.Start.Date
		}
		fmt.Printf("[%v] %v -> %v\n", date, events.Summary, events.Description)

		if strings.HasPrefix(events.Summary, "!!!") {
			title := events.Summary[3:]
			items := strings.Split(title, ":")

			persona := items[0]
			action := items[1]
			topic := items[2]
			details, _ := strconv.Unquote(events.Description)

			values := []string{persona, action, topic, details}
			parsedEvents = append(parsedEvents, values)
		}
	}
	return parsedEvents
}

// Write results into a CSV file
func writeToCsvFile(
	data [][]string,
) error {
	// Create a new CSV file
	file, err := os.Create("data.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write data to CSV file
	for _, value := range data {
		err := writer.Write(value)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
