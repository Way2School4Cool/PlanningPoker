package models

type Game struct {
	Status         Status
	Players        []Player
	QuestionNumber int
}

type Status int

const (
	WAITING_FOR_PLAYERS Status = iota
	WAITING_FOR_GUESSES
)

// String returns the string representation of the Status
func (s Status) ToString() string {
	switch s {
	case WAITING_FOR_PLAYERS:
		return "WAITING_FOR_PLAYERS"
	case WAITING_FOR_GUESSES:
		return "WAITING_FOR_GUESSES"
	default:
		return "UNKNOWN"
	}
}
