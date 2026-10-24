// Package analytics defines the record shape shared by iam-authz-server and
// iam-pump.
package analytics

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Record encodes the details of an authorization request.
type Record struct {
	TimeStamp  int64     `json:"timestamp" msgpack:"timestamp"`
	Username   string    `json:"username" msgpack:"username"`
	Effect     string    `json:"effect" msgpack:"effect"`
	Conclusion string    `json:"conclusion" msgpack:"conclusion"`
	Request    string    `json:"request" msgpack:"request"`
	Policies   string    `json:"policies" msgpack:"policies"`
	Deciders   string    `json:"deciders" msgpack:"deciders"`
	ExpireAt   time.Time `json:"expireAt" bson:"expireAt" msgpack:"expireAt"`
}

// GetFieldNames returns all field names in declaration order. It is used by
// the CSV pump to write a header row.
func (r *Record) GetFieldNames() []string {
	val := reflect.ValueOf(r).Elem()
	fields := make([]string, 0, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		fields = append(fields, val.Type().Field(i).Name)
	}

	return fields
}

// GetLineValues returns all field values in declaration order.
func (r *Record) GetLineValues() []string {
	val := reflect.ValueOf(r).Elem()
	fields := make([]string, 0, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		var thisVal string

		switch typeField.Type.String() {
		case "int", "int64":
			thisVal = strconv.FormatInt(valueField.Int(), 10)
		case "[]string":
			tmpVal, _ := reflect.TypeAssert[[]string](valueField)
			thisVal = strings.Join(tmpVal, ";")

		case "time.Time":
			tmpVal, _ := reflect.TypeAssert[time.Time](valueField)
			thisVal = tmpVal.String()

		default:
			thisVal = valueField.String()
		}

		fields = append(fields, thisVal)
	}

	return fields
}

// Filters defines the per-pump record filters.
type Filters struct {
	Usernames        []string `json:"usernames" mapstructure:"usernames"`
	SkippedUsernames []string `json:"skip_usernames" mapstructure:"skip_usernames"`
}

// ShouldFilter reports whether a record should be skipped.
func (f Filters) ShouldFilter(record Record) bool {
	switch {
	case len(f.SkippedUsernames) > 0 && stringInSlice(record.Username, f.SkippedUsernames):
		return true
	case len(f.Usernames) > 0 && !stringInSlice(record.Username, f.Usernames):
		return true
	}

	return false
}

// HasFilter reports whether any filter is configured.
func (f Filters) HasFilter() bool {
	return len(f.Usernames) > 0 || len(f.SkippedUsernames) > 0
}

func stringInSlice(a string, list []string) bool {
	return slices.Contains(list, a)
}
