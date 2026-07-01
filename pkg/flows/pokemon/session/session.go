package session

import pkgSession "github.com/muhlba91/hochschule-burgenland-bswe-director/pkg/session"

// Session represents a Pokémon game session.
type Session struct {
	pkgSession.Base

	// Players contains the two players participating in the session.
	Players map[string]*Player `json:"players" redis:"players"`
	// StartedAt is the timestamp when the game session started.
	StartedAt int64 `json:"startedAt" redis:"started_at"`
	// NextTurn indicates which player's turn is next in the game.
	NextTurn *string `json:"nextTurn,omitempty" redis:"next_turn,omitempty"`
	// Winner indicates which player has won the game.
	Winner *string `json:"winner,omitempty" redis:"winner,omitempty"`
}

// GetPlayerURLs returns the URLs of both players in the session.
func (s *Session) GetPlayerURLs() []string {
	urls := make([]string, 0, len(s.Players))
	for _, player := range s.Players {
		urls = append(urls, player.URL)
	}
	return urls
}

// GetPlayerByURL returns the player corresponding to the given URL.
// url: The URL of the player to retrieve.
func (s *Session) GetPlayerByURL(url string) *Player {
	for _, player := range s.Players {
		if player.URL == url {
			return player
		}
	}
	return nil
}

// GetPlayerByID returns the player corresponding to the given ID.
// id: The ID of the player to retrieve.
func (s *Session) GetPlayerByID(id string) *Player {
	return s.Players[id]
}

// GetOpponentForID returns the opponent player corresponding to the given ID.
// id: The ID of the player whose opponent is to be retrieved.
func (s *Session) GetOpponentForID(id string) *Player {
	for _, player := range s.Players {
		if player.ID != id {
			return player
		}
	}

	return nil
}
