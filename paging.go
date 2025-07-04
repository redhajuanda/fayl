package fayl

import (
	"github.com/redhajuanda/fayl/parser"
	"github.com/redhajuanda/kuysor"
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
