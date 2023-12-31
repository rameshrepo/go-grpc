package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	//List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)
	List(in *racing.ListRacesRequest) ([]*racing.Race, error)

	// Get a Race by identifier
	GetRace(filter *racing.GetRaceRequest) (*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *racesRepo) List(in *racing.ListRacesRequest) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, in.Filter)
	query = r.applySorting(query, in.Sortoptions)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) GetRace(filter *racing.GetRaceRequest) (*racing.Race, error) {
	var query string
	query = getRaceQueries()[racesList]
	if filter.RaceId != nil {
		query = query + " WHERE id = ?"
		rows, err := r.db.Query(query, filter.RaceId.GetValue())
		if err != nil {
			return nil, err
		}
		races, err := r.scanRaces(rows)
		if err != nil {
			return nil, err
		}
		if len(races) == 1 {
			return races[0], nil
		}
		return &racing.Race{}, nil
	}
	return &racing.Race{}, nil
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}
	if filter.Visible != nil {
		if filter.Visible.GetValue() == true {
			clauses = append(clauses, "visible=1")
		} else {
			clauses = append(clauses, "visible=0")
		}
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

func (r *racesRepo) applySorting(query string, sortoptions []*racing.SortOptions) string {
	var sort []string
	if sortoptions != nil {
		for _, s := range sortoptions {
			if s.Sortorder == racing.SortOrder_SORT_DESC {
				sort = append(sort, s.Field+" "+"DESC")
			} else if s.Sortorder == racing.SortOrder_SORT_ASC {
				sort = append(sort, s.Field+" "+"ASC")
			}
		}
		query += " ORDER BY " + strings.Join(sort, ",")
	}
	return query
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}
		if advertisedStart.IsZero() != true && advertisedStart.After(time.Now()) {
			race.Status = racing.Status_OPEN
		} else {
			race.Status = racing.Status_CLOSED
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
