package fayl

import "database/sql"

type ResultExec struct {
	sql.Result
}

type ResultQuery struct {
	Pagination *PaginationResponse
}
