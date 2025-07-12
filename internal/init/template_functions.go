package init

import (
	"errors"
	"html/template"
	"reflect"
	"strconv"
	"time"
)

// GetCustomTemplateFunctions returns a map of custom template functions
// This allows for easy extension of template functionality
func GetCustomTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"minus":   minus,
		"plus":    plus,
		"mul":     multiply,
		"div":     divide,
		"mod":     modulo,
		"max":     max,
		"min":     min,
		"inc":     increment,
		"dec":     decrement,
		"eq":      equal,
		"ne":      notEqual,
		"lt":      lessThan,
		"le":      lessThanOrEqual,
		"gt":      greaterThan,
		"ge":      greaterThanOrEqual,
		"len":     length,
		"str":     toString,
		"timeAgo": timeAgo,
		"dict":    createDict,
		"default": defaultValue,
	}
}

// minus subtracts the second number from the first
func minus(a, b interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() - bv.Int(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) - bv.Float(), nil
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() - float64(bv.Int()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() - bv.Float(), nil
		}
	}

	return nil, errors.New("minus: invalid operands")
}

// plus adds two numbers
func plus(a, b interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float(), nil
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() + float64(bv.Int()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() + bv.Float(), nil
		}
	}

	return nil, errors.New("plus: invalid operands")
}

// multiply multiplies two numbers
func multiply(a, b interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() * bv.Int(), nil
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) * bv.Float(), nil
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() * float64(bv.Int()), nil
		case reflect.Float32, reflect.Float64:
			return av.Float() * bv.Float(), nil
		}
	}

	return nil, errors.New("mul: invalid operands")
}

// divide divides the first number by the second
func divide(a, b interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if bv.Int() == 0 {
				return nil, errors.New("div: division by zero")
			}
			return av.Int() / bv.Int(), nil
		case reflect.Float32, reflect.Float64:
			if bv.Float() == 0 {
				return nil, errors.New("div: division by zero")
			}
			return float64(av.Int()) / bv.Float(), nil
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if bv.Int() == 0 {
				return nil, errors.New("div: division by zero")
			}
			return av.Float() / float64(bv.Int()), nil
		case reflect.Float32, reflect.Float64:
			if bv.Float() == 0 {
				return nil, errors.New("div: division by zero")
			}
			return av.Float() / bv.Float(), nil
		}
	}

	return nil, errors.New("div: invalid operands")
}

// modulo returns the remainder of dividing the first number by the second
func modulo(a, b interface{}) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	if av.Kind() == reflect.Int64 && bv.Kind() == reflect.Int64 {
		if bv.Int() == 0 {
			return nil, errors.New("mod: division by zero")
		}
		return av.Int() % bv.Int(), nil
	}

	return nil, errors.New("mod: invalid operands")
}

// max returns the maximum of two numbers
func max(a, b interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if av.Int() > bv.Int() {
				return av.Int()
			}
			return bv.Int()
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Float32, reflect.Float64:
			if av.Float() > bv.Float() {
				return av.Float()
			}
			return bv.Float()
		}
	}

	return a
}

// min returns the minimum of two numbers
func min(a, b interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if av.Int() < bv.Int() {
				return av.Int()
			}
			return bv.Int()
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Float32, reflect.Float64:
			if av.Float() < bv.Float() {
				return av.Float()
			}
			return bv.Float()
		}
	}

	return a
}

// increment adds 1 to a number
func increment(a interface{}) interface{} {
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return av.Int() + 1
	case reflect.Float32, reflect.Float64:
		return av.Float() + 1
	}

	return a
}

// decrement subtracts 1 from a number
func decrement(a interface{}) interface{} {
	av := reflect.ValueOf(a)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return av.Int() - 1
	case reflect.Float32, reflect.Float64:
		return av.Float() - 1
	}

	return a
}

// equal checks if two values are equal
func equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// notEqual checks if two values are not equal
func notEqual(a, b interface{}) bool {
	return !reflect.DeepEqual(a, b)
}

// lessThan checks if a is less than b
func lessThan(a, b interface{}) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() < bv.Int()
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) < bv.Float()
		}
	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() < float64(bv.Int())
		case reflect.Float32, reflect.Float64:
			return av.Float() < bv.Float()
		}
	}

	return false
}

// lessThanOrEqual checks if a is less than or equal to b
func lessThanOrEqual(a, b interface{}) bool {
	return lessThan(a, b) || equal(a, b)
}

// greaterThan checks if a is greater than b
func greaterThan(a, b interface{}) bool {
	return !lessThanOrEqual(a, b)
}

// greaterThanOrEqual checks if a is greater than or equal to b
func greaterThanOrEqual(a, b interface{}) bool {
	return !lessThan(a, b)
}

// length returns the length of a slice, array, map, or string
func length(v interface{}) int {
	if v == nil {
		return 0
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return rv.Len()
	}

	return 0
}

// toString converts a value to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(val).Int(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(val).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return ""
	}
}

// timeAgo returns a human-readable time difference from now
func timeAgo(t interface{}) string {
	var timeVal time.Time

	switch v := t.(type) {
	case time.Time:
		timeVal = v
	case *time.Time:
		if v == nil {
			return "unknown"
		}
		timeVal = *v
	default:
		return "unknown"
	}

	now := time.Now()
	duration := now.Sub(timeVal)

	// Handle future times
	if duration < 0 {
		duration = -duration
		// Could add "in X time" format if needed
		return "just now"
	}

	seconds := int(duration.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case years > 0:
		if years == 1 {
			return "1 year ago"
		}
		return strconv.Itoa(years) + " years ago"
	case months > 0:
		if months == 1 {
			return "1 month ago"
		}
		return strconv.Itoa(months) + " months ago"
	case weeks > 0:
		if weeks == 1 {
			return "1 week ago"
		}
		return strconv.Itoa(weeks) + " weeks ago"
	case days > 0:
		if days == 1 {
			return "1 day ago"
		}
		return strconv.Itoa(days) + " days ago"
	case hours > 0:
		if hours == 1 {
			return "1 hour ago"
		}
		return strconv.Itoa(hours) + " hours ago"
	case minutes > 0:
		if minutes == 1 {
			return "1 minute ago"
		}
		return strconv.Itoa(minutes) + " minutes ago"
	default:
		return "just now"
	}
}

// createDict creates a dictionary from alternating key-value pairs
func createDict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("dict: odd number of arguments")
	}

	dict := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict: keys must be strings")
		}
		dict[key] = values[i+1]
	}

	return dict, nil
}

// defaultValue returns the first value if it's not nil/empty, otherwise returns the default
func defaultValue(value, defaultVal interface{}) interface{} {
	if value == nil {
		return defaultVal
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		if rv.String() == "" {
			return defaultVal
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if rv.Len() == 0 {
			return defaultVal
		}
	case reflect.Bool:
		if !rv.Bool() {
			return defaultVal
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Int() == 0 {
			return defaultVal
		}
	case reflect.Float32, reflect.Float64:
		if rv.Float() == 0 {
			return defaultVal
		}
	}

	return value
}
