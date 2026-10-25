// Package veropts configures a versioned type.
//
// Exactly one version must be the current one; every other version needs a
// migration function with the same name.
package veropts

import (
	"errors"
	"reflect"
)

// Options lists the versions of a type.
type Options struct {
	Versions []Version
}

// Validate reports a missing or duplicated current version.
func Validate(o *Options) error {
	foundCurrVersion := false
	for _, v := range o.Versions {
		if v.Migration == "" {
			if foundCurrVersion {
				return errors.New("multiple current versions found")
			}
			foundCurrVersion = true
		}
	}
	return nil
}

// Version is one representation of a type together with the function that converts it to the current one.
type Version struct {
	Type      reflect.Type
	Migration string
}

// SetOption changes one field of Options.
type SetOption func(o *Options)

// WithVersion adds an older version and the name of its migration function.
func WithVersion(t reflect.Type, migration string) SetOption {
	return func(o *Options) {
		o.Versions = append(o.Versions, Version{
			Type:      t,
			Migration: migration,
		})
	}
}

// WithCurrentVersion adds the version the serializer writes.
func WithCurrentVersion(t reflect.Type) SetOption {
	return func(o *Options) {
		o.Versions = append(o.Versions, Version{
			Type:      t,
			Migration: "",
		})
	}
}

// Apply runs opts in order and validates the result.
func Apply(o *Options, opts ...SetOption) error {
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
	return Validate(o)
}
