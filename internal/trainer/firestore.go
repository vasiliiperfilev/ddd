package main

import (
	"context"
	"math"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/vasiliiperfilev/ddd/internal/trainer/openapi_types"
	"google.golang.org/api/iterator"
)

type db struct {
	firestoreClient *firestore.Client
}

type Pagination *struct {
	Page     *int32 `json:"page,omitempty"`
	PageSize *int32 `json:"page_size,omitempty"`
}

type DataWithPagination[T any] struct {
	Data               []T                              `json:"data"`
	PaginationMetadata openapi_types.PaginationMetadata `json:"paginationMetadata"`
}

func (d db) TrainerHoursCollection() *firestore.CollectionRef {
	return d.firestoreClient.Collection("trainer-hours")
}

func (d db) DocumentRef(dateTimeToUpdate time.Time) *firestore.DocumentRef {
	return d.TrainerHoursCollection().Doc(dateTimeToUpdate.Format("2006-01-02"))
}

func (d db) GetDates(ctx context.Context, params *openapi_types.GetTrainerAvailableHoursParams) (DataWithPagination[openapi_types.Date], error) {
	datesWithPagination, err := d.QueryDates(params, ctx)
	if err != nil {
		return DataWithPagination[openapi_types.Date]{}, err
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

func (d db) QueryDates(params *openapi_types.GetTrainerAvailableHoursParams, ctx context.Context) (DataWithPagination[openapi_types.Date], error) {
	query := d.
		TrainerHoursCollection().
		Where("Date.Time", ">=", params.DateFrom).
		Where("Date.Time", "<=", params.DateTo)
	limit, offset := d.getOffsetAndLimit(params.Pagination)
	total, err := d.getTotal(query, ctx)
	if err != nil {
		return DataWithPagination[openapi_types.Date]{}, err
	}
	iter := query.
		Limit(limit).
		Offset(offset).
		Documents(ctx)

	var dates []openapi_types.Date
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return DataWithPagination[openapi_types.Date]{}, err
		}

		date := openapi_types.Date{}
		if err := doc.DataTo(&date); err != nil {
			return DataWithPagination[openapi_types.Date]{}, err
		}
		date = setDefaultAvailability(date)
		dates = append(dates, date)
	}
	pagination, err := d.createPaginationMetadata(total, limit, offset)
	if err != nil {
		return DataWithPagination[openapi_types.Date]{}, err
	}
	return DataWithPagination[openapi_types.Date]{Data: dates, PaginationMetadata: pagination}, nil
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

func (d db) createPaginationMetadata(total int, limit int, offset int) (openapi_types.PaginationMetadata, error) {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	currentPage := (offset / limit) + 1
	return openapi_types.PaginationMetadata{CurrentPage: currentPage, PageSize: limit, TotalPages: totalPages, TotalRecords: total}, nil
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
