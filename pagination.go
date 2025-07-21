package fayl

import (
	"context"

	"github.com/redhajuanda/fayl/parser"
	"github.com/redhajuanda/fayl/vars"
	"github.com/redhajuanda/perkakas/pagination"

	"github.com/VauntDev/tqla"
	"github.com/pkg/errors"
	"github.com/redhajuanda/kuysor"
)

var (
	// Question is a PlaceholderFormat instance that replaces placeholders with
	// question-prefixed positional placeholders (e.g. ?, ?, ?).
	Question = tqla.Question
	// Dollar is a PlaceholderFormat instance that replaces placeholders with
	// dollar-prefixed positional placeholders (e.g. $1, $2, $3).
	Dollar = tqla.Dollar
	// Colon is a PlaceholderFormat instance that replaces placeholders with
	// colon-prefixed positional placeholders (e.g. :1, :2, :3).
	Colon = tqla.Colon
	// AtP is a PlaceholderFormat instance that replaces placeholders with
	// "@p"-prefixed positional placeholders (e.g. @p1, @p2, @p3).
	AtP = tqla.AtP
)

type Tabling struct {
	OrderBy         []string
	Pagination      *pagination.Pagination
	OffsetTotalData int64
}

var placeholderMapping = map[parser.Placeholder]kuysor.PlaceHolderType{
	Question: kuysor.Question,
	Dollar:   kuysor.Dollar,
	Colon:    kuysor.Colon,
	AtP:      kuysor.At,
}

func buildTabling(tabling *Tabling, pagination *pagination.Pagination) error {

	if pagination == nil {
		return errors.New("pagination cannot be nil")
	}

	if pagination.PerPage <= 0 {
		pagination.PerPage = vars.DefaultPaginationPerPage
	}

	if pagination.Page <= 0 {
		pagination.Page = vars.DefaultPaginationPage
	}

	if tabling == nil {
		tabling = &Tabling{
			Pagination: pagination,
		}
	} else {
		tabling.Pagination = pagination
	}

	return nil

}

func processTabling(_ context.Context, client *Client, tabling *Tabling, query string, parameters ...any) (*kuysor.Result, error) {

	if tabling == nil {
		return &kuysor.Result{Query: query, Args: parameters}, nil
	}

	kys := kuysor.NewInstance(kuysor.Options{
		StructTag:       vars.TagKey,
		PlaceHolderType: placeholderMapping[client.placeholder],
		NullSortMethod:  kuysor.BoolSort,
		DefaultLimit:    vars.DefaultPaginationPage,
	})

	// if pagination is set, build kuysor query
	if tabling.Pagination != nil && tabling.Pagination.Type == "cursor" {

		ks := kys.NewQuery(query, kuysor.Cursor).
			WithOrderBy(tabling.OrderBy...).
			WithArgs(parameters...).
			WithCursor(tabling.Pagination.Cursor).
			WithLimit(tabling.Pagination.PerPage)

		res, err := ks.Build()
		if err != nil {
			return nil, errors.Wrap(err, "failed to build kuysor query")
		}

		return res, nil
	} else if tabling.Pagination != nil && tabling.Pagination.Type == "offset" {
		ks := kys.NewQuery(query, kuysor.Offset).
			WithOrderBy(tabling.OrderBy...).
			WithArgs(parameters...).
			WithOffset(tabling.Pagination.GetOffset()).
			WithLimit(tabling.Pagination.PerPage)

		res, err := ks.Build()
		if err != nil {
			return nil, errors.Wrap(err, "failed to build kuysor query")
		}

		return res, nil

	} else if len(tabling.OrderBy) > 0 {
		ks := kys.NewQuery(query, "").
			WithOrderBy(tabling.OrderBy...).
			WithArgs(parameters...)

		res, err := ks.Build()
		if err != nil {
			return nil, errors.Wrap(err, "failed to build kuysor query")
		}

		return res, nil
	}

	return nil, nil

}
