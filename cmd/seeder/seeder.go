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

	err := s.models.Visits.Insert(visit)
	if err != nil {
		return nil, err
	}

	s.logger.Info().Str("LinkID", linkId.String()).Str("Visit ID", visit.ID.String()).Msg("Created Visit")

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
	s.logger.Info().Str("LinkID", link.ID.String()).Msg("Created Link")

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

func (s *seeder) seedDB() error {
	err := s.seedLink1()
	if err != nil {
		return err
	}

	return nil
}
