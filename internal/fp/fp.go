package fp

import (
	"fmt"
	"reflect"
)

func Map[T1, T2 any](f func(T1) T2, values []T1) []T2 {
	mapped := make([]T2, 0, len(values))

	for _, value := range values {
		mapped = append(mapped, f(value))
	}

	return mapped
}

func Reduce[T any](f func(T, T) T, initial T, values []T) T {
	acc := initial

	for _, value := range values {
		acc = f(acc, value)
	}

	return acc
}

func Where[T any](f func(T) bool, values []T) []T {
	var filtered []T

	for _, value := range values {
		if f(value) {
			filtered = append(filtered, value)
		}
	}

	return filtered
}

func Pluck[T1, T2 any](field string, values []T1) []T2 {
	plucked := make([]T2, 0, len(values))

	for _, value := range values {
		v := reflect.ValueOf(value)

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			panic("Pluck: only slices of structs can be plucked")
		}

		fieldValue := v.FieldByName(field)

		if !fieldValue.IsValid() {
			panic(fmt.Sprintf("Pluck: field '%s' does not exist on struct", field))
		}

		plucked = append(plucked, fieldValue.Interface().(T2))
	}

	return plucked
}
