package fayl

import (
	"context"
	"encoding/json"
	"fayl/mapper"
	"fayl/parser"
	"io"

	"github.com/pkg/errors"
)

type Runnerer interface {
	// WithParams initializes a new query with params.
	// Params can be a map or a struct, doesn't matter if you pass its pointer or its value.
	WithParams(interface{}) Runnerer
	// WithParam initializes a new query with param.
	// Param is a key-value pair.
	// The key is the parameter name, and the value is the parameter value.
	// If the parameter already exists, it will be overwritten.
	WithParam(key string, value interface{}) Runnerer
	// // WithCache(key string, ttl time.Duration, shouldCache ...ShouldCache) Runnerer
	// WithPaging(Paging) Runnerer
	// WithSorting(order []string) Runnerer
	ScanMap(dest map[string]interface{}) Runnerer
	ScanMaps(dest *[]map[string]interface{}) Runnerer
	ScanStruct(dest interface{}) Runnerer
	ScanStructs(dest interface{}) Runnerer
	ScanWriter(dest io.Writer) Runnerer
	Execute(ctx context.Context) (*Result, error)
}

// // Runner is a struct that contains runner configs to be executed.
type Runner struct {
	runnerCode string
	params     map[string]interface{}
	client     *Client
	// log           log.Logger
	inTransaction bool
	// requestID     string
	// // cacher        *Cacher
	scanner *Scanner
	// tabling *tabling.Tabling
	// result  *result.Result
	result *Result
	err    error
}

type runnerParams struct {
	runnerCode string
	client     *Client
	// log           log.Logger
	inTransaction bool
}

// newRunner returns a new Runner.
func newRunner(runnerParams runnerParams) *Runner {

	return &Runner{
		runnerCode: runnerParams.runnerCode,
		client:     runnerParams.client,
		params:     make(map[string]interface{}),
		// log:           runnerParams.log,
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
func (r *Runner) WithParam(key string, value interface{}) Runnerer {

	r.params[key] = value
	return r

}

// WithParams initializes a new query with params.
// Params can be a map or a struct, doesn't matter if you pass its pointer or its value.
func (r *Runner) WithParams(params interface{}) Runnerer {

	// check if params is a map
	if p, ok := params.(map[string]interface{}); ok {
		r.params = p
		return r
	}

	// check if params is a pointer to a map
	if p, ok := params.(*map[string]interface{}); ok {
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

// ScanMap initializes a runner with scanner map.
// dest is the destination of the scanner.
// It must be a map.
func (r *Runner) ScanMap(dest map[string]interface{}) Runnerer {

	r.scanner = newScanner(scannerMap, dest)
	return r

}

// ScanMaps initializes a runner with scanner maps.
// dest is the destination of the scanner.
// It must be a pointer to a slice of maps.
func (r *Runner) ScanMaps(dest *[]map[string]interface{}) Runnerer {

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

// Execute executes the query and returns the result.
func (r *Runner) Execute(ctx context.Context) (*Result, error) {

	parsr := parser.New()
	// parse query
	query, parameters, err := parsr.Parse(ctx, r.client.runners[r.runnerCode], r.params)
	if err != nil {
		return nil, err
	}

	// execute query
	rows, err := r.client.db.QueryxContext(ctx, query, parameters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	responser := &responser{
		rows: rows,
		// res:         res,
		mapScanFunc: MapScan,
		jsonMarshalFunc: func(v interface{}) ([]byte, error) {
			return json.Marshal(v)
		},
		// tabling: tabling,
		// meta:    metadata,
	} // create and return response

	// if rows != nil {
	// columns, _ := rows.Columns()
	// metadata.Columns = columns
	// }

	return &Result{
		Scanner: responser,
	}, nil
}
