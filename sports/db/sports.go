package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// SportsRepo provides repository access to races.
type SportsRepo interface {
	// Init will initialise the repository.
	Init() error

	// List will return a list of Sports Events.
	List(filter *sports.ListSportsEventRequest) ([]*sports.SportsEvent, error)
}

type sportsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRepo(db *sql.DB) SportsRepo {
	return &sportsRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *sportsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *sportsRepo) List(in *sports.ListSportsEventRequest) ([]*sports.SportsEvent, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getSportEventQuery()
	query, args = r.applyFilter(query, in.Filter)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *sportsRepo) applyFilter(query string, filter *sports.EventsFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.Ids) > 0 {
		clauses = append(clauses, "id IN ("+strings.Repeat("?,", len(filter.Ids)-1)+"?)")

		for _, sportsID := range filter.Ids {
			args = append(args, sportsID)
		}
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

func (m *sportsRepo) scanRaces(
	rows *sql.Rows,
) ([]*sports.SportsEvent, error) {
	var races []*sports.SportsEvent

	for rows.Next() {
		var race sports.SportsEvent
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.Name, &race.Location, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}
		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		races = append(races, &race)
	}

	return races, nil
}
