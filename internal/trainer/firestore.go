package main

import (
	"context"
	"math"
	"net/http"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/render"
	"github.com/oapi-codegen/runtime/types"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type db struct {
	firestoreClient *firestore.Client
}

type Pagination *struct {
	Page     *int32 `json:"page,omitempty"`
	PageSize *int32 `json:"page_size,omitempty"`
}

type DataWithPagination[T any] struct {
	Data               []T                `json:"data"`
	PaginationMetadata PaginationMetadata `json:"paginationMetadata"`
}

func (d db) TrainerHoursCollection() *firestore.CollectionRef {
	return d.firestoreClient.Collection("trainer-hours")
}

func (d db) DocumentRef(dateTimeToUpdate time.Time) *firestore.DocumentRef {
	return d.TrainerHoursCollection().Doc(dateTimeToUpdate.Format("2006-01-02"))
}

func (d db) DateModel(ctx context.Context, dateToFind time.Time) (Date, error) {
	date := Date{}

	doc, err := d.DocumentRef(dateToFind).Get(ctx)

	if err != nil && status.Code(err) != codes.NotFound {
		return Date{}, err
	}
	if err != nil && status.Code(err) == codes.NotFound {
		date = Date{
			Date: types.Date{Time: dateToFind.UTC().Truncate(time.Hour * 24)},
		}
	} else {
		err := doc.DataTo(&date)
		if err != nil {
			return Date{}, err
		}
	}

	date = setDefaultAvailability(date)

	return date, nil
}

func (d db) GetDates(ctx context.Context, params *GetTrainerAvailableHoursParams) (DataWithPagination[Date], error) {
	datesWithPagination, err := d.QueryDates(params, ctx)
	if err != nil {
		return DataWithPagination[Date]{}, err
	}
	// TODO: figure out what is this functionality
	// datesWithPagination.Data = addMissingDates(params, datesWithPagination.Data)

	for _, date := range datesWithPagination.Data {
		sort.Slice(date.Hours, func(i, j int) bool { return date.Hours[i].Hour.Before(date.Hours[j].Hour) })
	}
	sort.Slice(datesWithPagination.Data, func(i, j int) bool {
		date1 := datesWithPagination.Data[i].Date
		date2 := datesWithPagination.Data[j].Date
		return date1.Before(date2.Time)
	})

	return datesWithPagination, nil
}

func (d db) QueryDates(params *GetTrainerAvailableHoursParams, ctx context.Context) (DataWithPagination[Date], error) {
	query := d.
		TrainerHoursCollection().
		Where("Date.Time", ">=", params.DateFrom).
		Where("Date.Time", "<=", params.DateTo)
	limit, offset := d.getOffsetAndLimit(params.Pagination)
	total, err := d.getTotal(query, ctx)
	if err != nil {
		return DataWithPagination[Date]{}, err
	}
	iter := query.
		Limit(limit).
		Offset(offset).
		Documents(ctx)

	var dates []Date
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return DataWithPagination[Date]{}, err
		}

		date := Date{}
		if err := doc.DataTo(&date); err != nil {
			return DataWithPagination[Date]{}, err
		}
		date = setDefaultAvailability(date)
		dates = append(dates, date)
	}
	pagination, err := d.createPaginationMetadata(total, limit, offset)
	if err != nil {
		return DataWithPagination[Date]{}, err
	}
	return DataWithPagination[Date]{Data: dates, PaginationMetadata: pagination}, nil
}

func (d db) SaveModel(ctx context.Context, date Date) error {
	hasFreeHours := false
	for _, hour := range date.Hours {
		if hour.Available && !hour.HasTrainingScheduled {
			hasFreeHours = true
			break
		}
	}
	date.HasFreeHours = hasFreeHours

	_, err := d.DocumentRef(date.Date.Time).Set(ctx, date)
	return err
}

const (
	week = time.Hour * 24 * 7

	weeksAllowedToSet = 6
)

func (d db) UpdateAvailability(r *http.Request, availabilityToSet bool) error {
	hourUpdate := &HourUpdate{}
	if err := render.Decode(r, hourUpdate); err != nil {
		return errors.Wrap(err, "unable to decode request")
	}

	hoursByDates := map[time.Time][]time.Time{}
	for _, hour := range hourUpdate.Hours {
		if !hour.Round(time.Hour).Equal(hour) {
			return errors.Errorf("hour should be a full hour, but is %s", hour)
		}

		if hour.After(time.Now().Add(week * weeksAllowedToSet)) {
			return errors.Errorf("schedule can be only set for next %d weeks", weeksAllowedToSet)
		}
		if hour.Before(time.Now().Truncate(time.Hour * 24)) {
			return errors.New("cannot set schedule for past")
		}
		if hour.UTC().Hour() > maxHour || hour.UTC().Hour() < minHour {
			return errors.Errorf("hour must be between %d and %d UTC", minHour, maxHour)
		}

		date := hour.Truncate(time.Hour * 24)
		if _, ok := hoursByDates[date]; !ok {
			hoursByDates[date] = []time.Time{}
		}

		hoursByDates[date] = append(hoursByDates[date], hour)
	}

	for dateToFind, hours := range hoursByDates {
		date, err := d.DateModel(r.Context(), dateToFind)
		if err != nil {
			return errors.Wrapf(err, "unable to get data model for %s", dateToFind)
		}

		for _, hour := range hours {
			date, err = setAvailability(hour, date, availabilityToSet)
			if err != nil {
				return errors.Wrapf(err, "unable to set availability of %s %s", hour, date.Date)
			}
		}

		if err := d.SaveModel(r.Context(), date); err != nil {
			return errors.Wrapf(err, "unable to save %s", dateToFind)
		}
	}

	return nil
}

func setAvailability(hourUpdate time.Time, date Date, availabilityToSet bool) (Date, error) {
	found := false

	for i := range date.Hours {
		if date.Hours[i].Hour.Equal(hourUpdate) {
			if date.Hours[i].HasTrainingScheduled && (availabilityToSet == false) {
				return Date{}, errors.New("it is training already scheduled, cannot make not available")
			}

			date.Hours[i].Available = availabilityToSet
			found = true
		}
	}
	if !found {
		newHour := Hour{
			Available: availabilityToSet,
			Hour:      hourUpdate,
		}
		date.Hours = append(date.Hours, newHour)
	}

	return date, nil
}

func (d db) getOffsetAndLimit(pagination Pagination) (limit int, offset int) {
	if pagination == nil {
		return 10, 0
	}
	if pagination.Page == nil || *pagination.Page < 0 {
		return 10, 0
	}
	if pagination.PageSize == nil || *pagination.PageSize < 1 {
		return 10, 0
	}
	limit = int(*pagination.PageSize)
	offset = int((*pagination.Page - 1)) * limit
	return limit, offset
}

func (d db) createPaginationMetadata(total int, limit int, offset int) (PaginationMetadata, error) {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	currentPage := (offset / limit) + 1
	return PaginationMetadata{CurrentPage: currentPage, PageSize: limit, TotalPages: totalPages, TotalRecords: total}, nil
}

// TODO: figure out better way to get total count in firestore
func (d db) getTotal(query firestore.Query, ctx context.Context) (int, error) {
	iter := query.Documents(ctx)
	total := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}
		total++
	}
	return total, nil
}
