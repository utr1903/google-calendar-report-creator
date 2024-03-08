package dto

type Event struct {
	Persona string `json:"persona"`
	Action  string `json:"action"`
	Topic   string `json:"topic"`
	Details string `json:"details"`
}
