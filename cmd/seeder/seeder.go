package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/matthewsaunders/link-shortener-api/internal/data"
	"github.com/rs/zerolog"
)

type seeder struct {
	filename string
	logger   *zerolog.Logger
	models   data.Models
}

func (s *seeder) createVisit(linkId uuid.UUID, createdAt time.Time) (*data.Visit, error) {
	visit := &data.Visit{
		LinkID:    linkId,
		CreatedAt: createdAt,
	}

	err := s.models.Visits.Seed(visit)
	if err != nil {
		return nil, err
	}

	s.logger.Info().Str("LinkID", linkId.String()).Str("Visit ID", visit.ID.String()).Str("created_at", visit.CreatedAt.String()).Msg("")

	return visit, nil
}

func (s *seeder) seedLink1() error {
	numHours := 24
	now := time.Now()
	duration := time.Hour * -time.Duration(numHours)
	linkTime := now.Add(duration)

	link := &data.Link{
		Name:        "HeroIcons",
		Destination: "https://heroicons.com/",
		Token:       s.models.Links.GenerateNewToken(),
		CreatedAt:   linkTime,
	}

	err := s.models.Links.Insert(link)
	if err != nil {
		return err
	}
	s.logger.Info().Str("LinkID", link.ID.String()).Msg("")

	// Create a bunch of visits for the Link
	var visitTime time.Time

	for i := 0; i < numHours; i++ {
		duration := time.Hour * -time.Duration(i)
		visitTime = linkTime.Add(duration)

		// Each hour should have 1 more visit than the previous hour
		for j := 0; j < i+1; j++ {
			_, err = s.createVisit(link.ID, visitTime)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *seeder) seedLink2() error {
	numDays := 10
	now := time.Now()
	duration := time.Hour * 24 * -time.Duration(numDays)
	linkTime := now.Add(duration)

	link := &data.Link{
		Name:        "tailwindcss",
		Destination: "https://tailwindcss.com/",
		Token:       s.models.Links.GenerateNewToken(),
		CreatedAt:   linkTime,
	}

	err := s.models.Links.Insert(link)
	if err != nil {
		return err
	}
	s.logger.Info().Str("LinkID", link.ID.String()).Msg("Created Link")

	// Create a bunch of visits for the Link
	var visitTime time.Time

	for i := 0; i < numDays; i++ {
		visitTime = linkTime.AddDate(0, 0, -1*(numDays-i))

		// Each hour should have 1 more visit than the previous day
		for j := 0; j < i+1; j++ {
			_, err = s.createVisit(link.ID, visitTime)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *seeder) seedDB() error {
	var err error

	err = s.seedLink1()
	if err != nil {
		return err
	}

	err = s.seedLink2()
	if err != nil {
		return err
	}

	return nil
}
