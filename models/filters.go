package models

import (
	"errors"
	"math"
	"strings"
)

// Filters ...
type Filters struct {
	Page     int
	PageSize int
	OrderBy  string
	Query    string
}

// MetaData ...
type MetaData struct {
	CurrentPage  int
	PageSize     int // PageSize is the number of records in a page
	FirstPage    int
	NextPage     int
	PrevPage     int
	LastPage     int
	TotalRecords int // TotalRecords is the total number of records across all pages
}

// Validate ...
func (f *Filters) Validate() error {
	if f.Page <= 0 || f.Page >= 10_000_000 {
		return errors.New("Invalid page range")
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		return errors.New("Invalid page size")
	}
	return nil
}

func (f *Filters) addOrdering(query string) string {
	switch {
	case f.OrderBy == "popular":
		return strings.Replace(query, "#orderby#", "ORDER BY votes DESC, p.created_at DESC", 1)
	default:
		return strings.Replace(query, "#orderby#", "ORDER BY p.created_at DESC", 1)
	}
}

func (f *Filters) addWhere(query string) string {
	switch {
	case len(f.Query) > 0:
		return strings.Replace(query, "#where#", "WHERE LOWER(p.title) LIKE $1", 1)
	default:
		return strings.Replace(query, "#where#", "", 1)
	}
}

func (f *Filters) addLimitOffset(query string) string {
	switch {
	case len(f.Query) > 0:
		return strings.Replace(query, "#limit#", "LIMIT $2 OFFSET $3", 1) // because then $1 will be the query argument
	default:
		return strings.Replace(query, "#limit#", "LIMIT $1 OFFSET $2", 1)
	}
}

func (f *Filters) applyTemplate(query string) string {
	return f.addLimitOffset(f.addWhere(f.addOrdering(query)))
}

func (f *Filters) limit() int {
	return f.PageSize
}

func (f *Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calculateMetaData(totalRecords, page, pageSize int) MetaData {
	if totalRecords == 0 {
		return MetaData{}
	}
	meta := MetaData{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
	meta.NextPage = meta.CurrentPage + 1

	if meta.CurrentPage <= meta.FirstPage {
		meta.PrevPage = 0
	} else {
		meta.PrevPage = meta.CurrentPage - 1
	}

	return meta
}
