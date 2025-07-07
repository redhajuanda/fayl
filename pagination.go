package fayl

import (
	"github.com/VauntDev/tqla"
	"github.com/redhajuanda/fayl/parser"
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
	OrderBy    []string
	Pagination *Pagination
}

type Pagination struct {
	Type   string
	Limit  int
	Offset int
	Cursor string
}

func NewPaginationCursor(cursor string, limit int) *Pagination {

	return &Pagination{
		Type:   "cursor",
		Limit:  limit,
		Offset: 0,
		Cursor: cursor,
	}
}

func NewPaginationOffset(offset, limit int) *Pagination {
	return &Pagination{
		Type:   "offset",
		Limit:  limit,
		Offset: offset,
		Cursor: "",
	}
}

type PaginationResponse struct {
	Type      string
	Next      string
	Prev      string
	TotalData int
}

var placeholderMapping = map[parser.Placeholder]kuysor.PlaceHolderType{
	Question: kuysor.Question,
	Dollar:   kuysor.Dollar,
	Colon:    kuysor.Colon,
	AtP:      kuysor.At,
}
