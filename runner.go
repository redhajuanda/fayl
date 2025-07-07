package fayl

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"

	"github.com/jmoiron/sqlx"
	"github.com/redhajuanda/fayl/mapper"
	"github.com/redhajuanda/fayl/parser"
	"github.com/redhajuanda/fayl/vars"
	"github.com/redhajuanda/kuysor"

	"github.com/pkg/errors"
)

type Runnerer interface {
	// WithParams initializes a new query with params.
	// Params can be a map or a struct, doesn't matter if you pass its pointer or its value.
	WithParams(any) Runnerer
	// WithParam initializes a new query with param.
	// Param is a key-value pair.
	// The key is the parameter name, and the value is the parameter value.
	// If the parameter already exists, it will be overwritten.
	WithParam(key string, value any) Runnerer
	// // WithCache(key string, ttl time.Duration, shouldCache ...ShouldCache) Runnerer
	// WithPagination adds pagination to the query.
	// pagination is a Pagination struct that contains the pagination options.
	// Pagination can be nil, in which case it will not be added to the query.
	// If you use cursor pagination, you are required to set the order by clause using WithOrderBy.
	WithPagination(pagination *Pagination) Runnerer
	// WithOrderBy initializes a new query with order by.
	// orderBy is a list of columns to order by.
	// It can be a single column or multiple columns.
	// The order can be ascending or descending using prefix: "+" for ascending and "-" for descending.
	// If no prefix is used, it will default to ascending order.
	// Example: WithOrderBy("name", "-created_at") will order by name ascending and created_at descending.
	WithOrderBy(orderBy ...string) Runnerer
	// ScanMap initializes a runner with scanner map.
	// dest is the destination of the scanner.
	// It must be a map.
	ScanMap(dest map[string]any) Runnerer
	// ScanMaps initializes a runner with scanner maps.
	// dest is the destination of the scanner.
	// It must be a pointer to a slice of maps.
	// If you want to scan a single map, use ScanMap instead.
	ScanMaps(dest *[]map[string]any) Runnerer
	// ScanStruct initializes a runner with scanner struct.
	// dest is the destination of the scanner.
	// It must be a pointer to a struct.
	ScanStruct(dest any) Runnerer
	// ScanStructs initializes a runner with scanner structs.
	// dest is the destination of the scanner.
	// It must be a pointer to a slice of structs.
	ScanStructs(dest any) Runnerer
	// ScanWriter initializes a runner with scanner writer.
	// dest is the destination of the scanner.
	ScanWriter(dest io.Writer) Runnerer
	// Exec executes the query and returns the result.
	// It returns a ResultExec struct that contains the result of the execution.
	Exec(ctx context.Context) (*ResultExec, error)
	// Query executes the query and scans the result to the destination.
	// It returns a ResultQuery struct that contains the scanned result.
	// The destination must be set using ScanMap, ScanMaps, ScanStruct, ScanStructs, or ScanWriter.
	Query(ctx context.Context) (*ResultQuery, error)
}

// // Runner is a struct that contains runner configs to be executed.
type Runner struct {
	runnerCode    string
	params        map[string]any
	client        *Client
	log           Logger
	inTransaction bool
	// requestID     string
	// // cacher        *Cacher
	scanner *Scanner
	// tabling *tabling.Tabling
	// result  *result.Result
	tabling *Tabling
	kuysor  *kuysor.Kuysor
	// result     *Result
	err error
}

type runnerParams struct {
	runnerCode    string
	client        *Client
	log           Logger
	inTransaction bool
}

// newRunner returns a new Runner.
func newRunner(runnerParams runnerParams) *Runner {

	return &Runner{
		runnerCode:    runnerParams.runnerCode,
		client:        runnerParams.client,
		params:        make(map[string]any),
		log:           runnerParams.log,
		inTransaction: runnerParams.inTransaction,
		// cacher:        &Cacher{},
		// result: &result.Result{
		// 	Metadata: &result.Metadata{},
		// },
	}

}

// WithParam initializes a new query with param.
// Param is a key-value pair.
// The key is the parameter name, and the value is the parameter value.
// If the parameter already exists, it will be overwritten.
func (r *Runner) WithParam(key string, value any) Runnerer {

	r.params[key] = value
	return r

}

// WithParams initializes a new query with params.
// Params can be a map or a struct, doesn't matter if you pass its pointer or its value.
func (r *Runner) WithParams(params any) Runnerer {

	// check if params is a map
	if p, ok := params.(map[string]any); ok {
		r.params = p
		return r
	}

	// check if params is a pointer to a map
	if p, ok := params.(*map[string]any); ok {
		r.params = *p
		return r
	}

	// check if params is a struct
	if isStruct(params) {

		err := mapper.Decode(params, &r.params)
		if err != nil {
			r.err = errors.Wrap(err, "failed to decode params")
		}

		return r

	}

	r.err = errors.New("params must be a map or a struct")
	return r

}

// WithPagination adds pagination to the query.
func (r *Runner) WithPagination(pagination *Pagination) Runnerer {

	if r.tabling == nil {
		r.tabling = &Tabling{
			Pagination: pagination,
		}
	} else {
		r.tabling.Pagination = pagination
	}
	return r

}

// WithOrderBy initializes a new query with order by.
// orderBy is a list of columns to order by.
// It can be a single column or multiple columns.
// The order can be ascending or descending using prefix: "+" for ascending and "-" for descending.
// If no prefix is used, it will default to ascending order.
// Example: WithOrderBy("name", "-created_at") will order by name ascending and created_at descending.
func (r *Runner) WithOrderBy(orderBy ...string) Runnerer {

	if r.tabling == nil {
		r.tabling = &Tabling{
			OrderBy: orderBy,
		}
	} else {
		r.tabling.OrderBy = orderBy
	}
	return r

}

// ScanMap initializes a runner with scanner map.
// dest is the destination of the scanner.
// It must be a map.
func (r *Runner) ScanMap(dest map[string]any) Runnerer {

	r.scanner = newScanner(scannerMap, dest)
	return r

}

// ScanMaps initializes a runner with scanner maps.
// dest is the destination of the scanner.
// It must be a pointer to a slice of maps.
func (r *Runner) ScanMaps(dest *[]map[string]any) Runnerer {

	r.scanner = newScanner(scannerMaps, dest)
	return r

}

// ScanStruct initializes a runner with scanner struct.
// dest is the destination of the scanner.
// It must be a pointer to a struct.
func (r *Runner) ScanStruct(dest interface{}) Runnerer {

	r.scanner = newScanner(scannerStruct, dest)
	return r

}

// ScanStructs initializes a runner with scanner structs.
// dest is the destination of the scanner.
// It must be a pointer to a slice of structs.
func (r *Runner) ScanStructs(dest interface{}) Runnerer {

	r.scanner = newScanner(scannerStructs, dest)
	return r

}

// ScanWriter initializes a runner with scanner writer.
// dest is the destination of the scanner.
// It must be a writer.
func (r *Runner) ScanWriter(dest io.Writer) Runnerer {

	r.scanner = newScanner(scannerWriter, dest)
	return r

}

// Exec executes the query and returns the result.
func (r *Runner) Exec(ctx context.Context) (*ResultExec, error) {

	var (
		ps     = parser.New()
		result sql.Result
	)

	r.log.WithContext(ctx).WithParams(map[string]any{
		"runner_code": r.runnerCode,
		"params":      r.params,
		"placeholder": r.client.placeholder,
	}).Debug("Parsing query")

	// parse query
	query, parameters, err := ps.Parse(ctx, r.client.runners[r.runnerCode], r.params, r.client.placeholder)
	if err != nil {
		return nil, err
	}

	if r.inTransaction {

		r.log.WithContext(ctx).WithParams(map[string]any{
			"runner_code": r.runnerCode,
			"query":       query,
			"params":      parameters,
		}).Info("Executing query in transaction")

		// if in transaction, use the transaction context
		tx, err := r.client.db.getTx(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get transaction")
		}

		// execute query
		result, err = tx.ExecContext(ctx, query, parameters...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to execute query in transaction")
		}

	} else {

		r.log.WithContext(ctx).WithParams(map[string]any{
			"runner_code": r.runnerCode,
			"query":       query,
			"params":      parameters,
		}).Info("Executing query")

		// execute query
		result, err = r.client.db.ExecContext(ctx, query, parameters...)
		if err != nil {
			return nil, err
		}
	}

	return &ResultExec{
		result,
	}, nil

}

// Query executes the query and scans the result to the destination.
func (r *Runner) Query(ctx context.Context) (*ResultQuery, error) {

	var (
		ps   = parser.New()
		rows *sqlx.Rows
	)

	r.log.WithContext(ctx).WithParams(map[string]any{
		"runner_code": r.runnerCode,
		"params":      r.params,
		"placeholder": r.client.placeholder,
	}).Debug("Parsing query")

	// parse query
	query, parameters, err := ps.Parse(ctx, r.client.runners[r.runnerCode], r.params, r.client.placeholder)
	if err != nil {
		return nil, err
	}

	// build pagination cursor if pagination is set
	rs, err := r.buildTabling(ctx, query, parameters...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build pagination cursor")
	}
	if rs != nil {
		query = rs.Query
		parameters = rs.Args
	}

	if r.inTransaction {

		r.log.WithContext(ctx).WithParams(map[string]any{
			"runner_code": r.runnerCode,
			"query":       query,
			"params":      parameters,
		}).Info("Querying query in transaction")

		// if in transaction, use the transaction context
		tx, err := r.client.db.getTx(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get transaction")
		}

		// execute query
		rows, err = tx.QueryxContext(ctx, query, parameters...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to execute query in transaction")
		}

	} else {

		r.log.WithContext(ctx).WithParams(map[string]any{
			"runner_code": r.runnerCode,
			"query":       query,
			"params":      parameters,
		}).Info("Querying query")

		// execute query
		rows, err = r.client.db.QueryxContext(ctx, query, parameters...)
		if err != nil {
			return nil, err
		}
	}

	responser := &responser{
		rows:        rows,
		mapScanFunc: MapScan,
		jsonMarshalFunc: func(v interface{}) ([]byte, error) {
			return json.Marshal(v)
		},
		kuysor:  rs,
		tabling: r.tabling,
	}

	// scan result
	err = r.scan(ctx, responser)
	if err != nil {
		return nil, err
	}

	return &ResultQuery{
		Pagination: responser.pgres,
	}, nil
}

func (r *Runner) buildTabling(ctx context.Context, query string, parameters ...any) (*kuysor.Result, error) {

	kys := kuysor.NewInstance(kuysor.Options{
		StructTag:       vars.TagKey,
		PlaceHolderType: placeholderMapping[r.client.placeholder],
		NullSortMethod:  kuysor.BoolSort,
		DefaultLimit:    10,
	})

	if r.tabling != nil {
		// if pagination is set, build kuysor query
		if r.tabling.Pagination != nil && r.tabling.Pagination.Type == "cursor" {

			ks := kys.NewQuery(query, kuysor.Cursor).
				WithOrderBy(r.tabling.OrderBy...).
				WithArgs(parameters...).
				WithCursor(r.tabling.Pagination.Cursor).
				WithLimit(r.tabling.Pagination.Limit)

			res, err := ks.Build()
			if err != nil {
				return nil, errors.Wrap(err, "failed to build kuysor query")
			}

			r.kuysor = ks

			return res, nil
		} else if r.tabling.Pagination != nil && r.tabling.Pagination.Type == "offset" {
			ks := kys.NewQuery(query, kuysor.Offset).
				WithOrderBy(r.tabling.OrderBy...).
				WithArgs(parameters...).
				WithOffset(r.tabling.Pagination.Offset).
				WithLimit(r.tabling.Pagination.Limit)

			res, err := ks.Build()
			if err != nil {
				return nil, errors.Wrap(err, "failed to build kuysor query")
			}

			r.kuysor = ks

			return res, nil
		} else if r.tabling.Pagination == nil && r.tabling.OrderBy != nil && len(r.tabling.OrderBy) > 0 {
			ks := kuysor.NewQuery(query, "").
				WithOrderBy(r.tabling.OrderBy...).
				WithArgs(parameters...)

			res, err := ks.Build()
			if err != nil {
				return nil, errors.Wrap(err, "failed to build kuysor query")
			}

			r.kuysor = ks

			return res, nil
		}
	}

	return nil, nil

}

// scan scans the result to the destination.
func (r *Runner) scan(ctx context.Context, sc Scannerer) error {

	if r.scanner == nil {
		r.scanner = newScanner(noScanner, nil)
	}

	// logging debug
	// r.log.With(ctx).WithParams(log.Params{"runner_code": r.runnerCode}).Debug("scanning result")

	// scan result
	switch r.scanner.scannerType {
	case scannerMap:

		err := sc.ScanMap(r.scanner.dest.(map[string]any))
		if err != nil {
			return err
		}

	case scannerMaps:

		err := sc.ScanMaps(r.scanner.dest.(*[]map[string]any))
		if err != nil {
			return err
		}

	case scannerStruct:

		err := sc.ScanStruct(r.scanner.dest)
		if err != nil {
			return err
		}

	case scannerStructs:

		err := sc.ScanStructs(r.scanner.dest)
		if err != nil {
			return err
		}

	case scannerWriter:

		err := sc.ScanWriter(r.scanner.dest.(io.Writer))
		if err != nil {
			return err
		}

	default:

		// r.log.With(ctx).WithParams(log.Params{"runner_code": r.runnerCode}).Debug("no scanner found, closing scanner")
		err := sc.Close()
		if err != nil {
			return err
		}

	}

	return nil
}
