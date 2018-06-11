package awsparser

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/pavlov-tony/xproject/pkg/cloud/awsparser/csvparseutils"
	errs "github.com/pavlov-tony/xproject/pkg/cloud/awsparser/errors"
)

// RawCsv represents the matrix of strings parsed from CSV
type RawCsv struct {
	records [][]string
}

// FromRows creates RawCsv from [][]string
func RawCsvFromRows(rows [][]string) *RawCsv {
	return &RawCsv{
		records: rows,
	}
}

// FromReader creates RawCsv from io.Reader
func RawCsvFromReader(in io.Reader) (*RawCsv, error) {
	r := csv.NewReader(in)
	records, err := r.ReadAll()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("can't create RawCsv from io.Reader: %v", err))
	}

	return &RawCsv{
		records: records,
	}, nil
}

// Rows returns the containing rows of RawCsv
func (raw *RawCsv) Rows() [][]string {
	return raw.records
}

// Row returns the row of RawCsv by index
func (raw *RawCsv) Row(index int) ([]string, error) {
	if index < len(raw.records) {
		return raw.records[index], nil
	} else {
		return nil, errs.NewIndexError(index)
	}
}

// ColumnIndexByName returns the number of the provided header name
func (raw *RawCsv) ColumnIndexByName(name string) (int, error) {
	if len(raw.records) == 0 {
		return 0, errors.New("RawCsv is empty.")
	}

	header := raw.records[0]

	for i, colName := range header {
		if colName == name {
			return i, nil
		}
	}

	return 0, errors.New(fmt.Sprintf("header with name \"%v\" was not found", name))
}

// FilterByNames returns RawCsv with provided header names only
func (raw *RawCsv) FilterByNames(columns []string) (*RawCsv, error) {
	indeces := make([]int, len(columns))
	for i, columnName := range columns {
		idx, err := raw.ColumnIndexByName(columnName)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("can't get column index by name \"%v\": %v", columnName, err))
		}
		indeces[i] = idx
	}

	return raw.FilterByIndeces(indeces)
}

// FilterByIndeces returns RawCsv with provided header indeces only
func (raw *RawCsv) FilterByIndeces(indeces []int) (*RawCsv, error) {
	rows := raw.Rows()
	if rows == nil {
		return nil, errors.New("RawCsv is empty.")
	}

	final := make([][]string, len(rows))
	for i, r := range rows {
		final[i] = make([]string, len(indeces))
		for j, c := range indeces {
			final[i][j] = r[c]
		}
	}

	return RawCsvFromRows(final), nil
}

// GetSummary returns the summary of the RawCsv
// It's basically grouping by the GroupBy field, and sum the SumBy columns of the RawCsv
func (rawcsv *RawCsv) GetSummary(cfg *SummaryConfig) (*CsvSummary, error) {
	if cfg.GroupBy == "" || cfg.SumBy == nil {
		return nil, errors.New("can't produce CsvSummary: GroupBy and SumBy mustn't be empty")
	}

	filtered, err := rawcsv.FilterByNames(append([]string{cfg.GroupBy}, cfg.SumBy...))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("can't filter by names: %v", err))
	}

	summ := make(map[string]([]float64))

	for _, row := range filtered.Rows()[1:] {
		if summ[row[0]] == nil {
			floats, err := csvparseutils.StringSliceToFloat64(row[1:])
			if err != nil {
				return nil, errors.New(fmt.Sprintf("can't parse SumBy to slice of floats: %v", err))
			}
			summ[row[0]] = floats

			continue
		} else {
			prevRow := summ[row[0]]
			for j, v := range prevRow {
				value, err := strconv.ParseFloat(row[j+1], 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("can't parse '%v' to float64: %v", v, err))
				}
				prevRow[j] += value
			}
			summ[row[0]] = prevRow
		}
	}

	return NewCsvSummary(filtered.Rows()[0], &summ), nil
}
