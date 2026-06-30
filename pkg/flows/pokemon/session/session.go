package session

import pkgSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"

// Session represents a Pokémon game session.
type Session struct {
	pkgSession.Base

	// PlayerA is the first player in the game.
	PlayerA Player `json:"playerA" redis:"player_a"`
	// PlayerB is the second player in the game.
	PlayerB Player `json:"playerB" redis:"player_b"`
	// StartedAt is the timestamp when the game session started.
	StartedAt int64 `json:"startedAt" redis:"started_at"`
	// NextTurn indicates which player's turn is next in the game.
	NextTurn *string `json:"nextTurn,omitempty" redis:"next_turn,omitempty"`
	// Winner indicates which player has won the game.
	Winner *string `json:"winner,omitempty" redis:"winner,omitempty"`
}

// GetPlayerURLs returns the URLs of both players in the session.
func (s *Session) GetPlayerURLs() []string {
	return []string{s.PlayerA.URL, s.PlayerB.URL}
}
