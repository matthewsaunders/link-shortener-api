package data

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/matthewsaunders/link-shortener-api/internal/validator"
	"github.com/rs/zerolog"
)

const layoutISO = "2006-01-02"

type Visit struct {
	ID         uuid.UUID `json:"id"`
	LinkID     uuid.UUID `json:"link_id"`
	CreatedAt  time.Time `json:"created_at"`
	Referrer   string    `json:"referrer"`
	RemoteAddr string    `json:remote_address`
}

type AggregatedVists struct {
	Date   string `json:"date"`
	Visits int    `json:"visits"`
}

type VisitData struct {
	TotalVisits     int                `json:"total_visits"`
	SevenDayVisits  int                `json:"seven_day_visits"`
	VisitsPerDay    float64            `json:"visits_per_day"`
	AggregatedVists []*AggregatedVists `json:"visits"`
}

type VisitModel struct {
	DB       *sql.DB
	InfoLog  *zerolog.Logger
	ErrorLog *zerolog.Logger
}

func (m VisitModel) Insert(visit *Visit) error {
	query := `
		INSERT INTO visits (link_id, referrer, remote_address)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{visit.LinkID, visit.Referrer, visit.RemoteAddr}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&visit.ID, &visit.CreatedAt)
}

func (m VisitModel) Seed(visit *Visit) error {
	query := `
		INSERT INTO visits (link_id, referrer, remote_address, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{visit.LinkID, visit.Referrer, visit.RemoteAddr, visit.CreatedAt}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&visit.ID)
}

func (m VisitModel) printFormattedDate(date time.Time) string {
	return date.Format(layoutISO)
}

func (m VisitModel) parseFormattedDate(date string) time.Time {
	parsedTime, _ := time.Parse(layoutISO, date)
	return parsedTime
}

func (m VisitModel) getTotalCount(link *Link, data *VisitData) error {
	query := fmt.Sprintf(`
		SELECT count(*)
		FROM visits
		WHERE visits.link_id = $1
	`)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, link.ID).Scan(&data.TotalVisits)
	if err != nil {
		return err
	}

	return nil
}

func (m VisitModel) getAggregatedCount(link *Link, data *VisitData) error {
	dateBuckets := make(map[string]int)
	numDays := 7
	date := time.Now()

	// Initialize the date buckets
	for i := 0; i < numDays; i++ {
		bucketName := m.printFormattedDate(date)
		dateBuckets[bucketName] = 0

		// Move date back one day
		date = date.AddDate(0, 0, -1)
	}

	query := fmt.Sprintf(`
		SELECT CAST(created_at as DATE), count(*)
		FROM visits
		WHERE visits.link_id = $1
		AND created_at > now() - interval '1 week'
		GROUP BY CAST(created_at as DATE)
		ORDER BY CAST(created_at as DATE)
	`)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{link.ID}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			m.ErrorLog.Error().Err(err).Msg("")
		}
	}()

	for rows.Next() {
		var date time.Time
		var visits int

		err := rows.Scan(
			&date,
			&visits,
		)
		if err != nil {
			return err
		}

		bucketName := m.printFormattedDate(date)
		dateBuckets[bucketName] = visits
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// Update the dateBuckets with the queried data
	for date, visits := range dateBuckets {
		data.AggregatedVists = append(data.AggregatedVists, &AggregatedVists{
			Date:   date,
			Visits: visits,
		})
	}

	sort.Slice(data.AggregatedVists, func(i, j int) bool {
		iDate := m.parseFormattedDate(data.AggregatedVists[i].Date)
		jDate := m.parseFormattedDate(data.AggregatedVists[j].Date)
		return iDate.Before(jDate)
	})

	return nil
}

func (m *VisitModel) calculateVisitData(data *VisitData) error {
	count := 0

	for _, aggregatedVists := range data.AggregatedVists {
		count = count + aggregatedVists.Visits
	}

	avgFloat := float64(count) / float64(len(data.AggregatedVists))
	data.VisitsPerDay = math.Round(avgFloat*100) / 100
	data.SevenDayVisits = count

	return nil
}

func (m VisitModel) GetData(link *Link) (*VisitData, error) {
	data := &VisitData{}

	err := m.getAggregatedCount(link, data)
	if err != nil {
		return nil, err
	}

	err = m.getTotalCount(link, data)
	if err != nil {
		return nil, err
	}

	err = m.calculateVisitData(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ValidateVisit(v *validator.Validator, visit *Visit) {
	// TODO
}
