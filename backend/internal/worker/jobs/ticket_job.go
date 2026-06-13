package jobs

import "log"

func HandleTicketGenerated(logger *log.Logger, payload []byte) error {
	logger.Printf("ticket job received payload=%s", string(payload))
	return nil
}
