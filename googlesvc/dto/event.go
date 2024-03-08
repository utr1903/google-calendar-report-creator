package dto

type Attendee struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Response string `json:"response"`
}

type Event struct {
	StartDate string     `json:"startDate"`
	EndDate   string     `json:"endDate"`
	Persona   string     `json:"persona"`
	Action    string     `json:"action"`
	Topic     string     `json:"topic"`
	Details   string     `json:"details"`
	Attendees []Attendee `json:"attendees"`
}
