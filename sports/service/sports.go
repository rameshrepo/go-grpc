package service

import (
	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"golang.org/x/net/context"
)

type Sports interface {
	// ListSportEvents will return a collection of sport events.
	ListSportEvents(ctx context.Context, in *sports.ListSportsEventRequest) (*sports.ListSportsEventResponse, error)
}

// racingService implements the Racing interface.
type sportsService struct {
	sportsRepo db.SportsRepo
}

func NewSportsService(sportsRepo db.SportsRepo) Sports {
	return &sportsService{sportsRepo}
}

func (s *sportsService) ListSportEvents(ctx context.Context, in *sports.ListSportsEventRequest) (*sports.ListSportsEventResponse, error) {
	sportsevents, err := s.sportsRepo.List(in)
	if err != nil {
		return nil, err
	}

	return &sports.ListSportsEventResponse{Sportevents: sportsevents}, nil
}
