package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type ParseResult[T any] struct {
	value  T
	source *Source
}

type Source struct {
	str   string
	index int
}

func NewSource(str string, index int) *Source {
	return &Source{str, index}
}

func (s *Source) Match(pattern string) *ParseResult[string] {
	if s.index >= len(s.str) {
		return nil
	}

	cleanPattern := strings.TrimPrefix(pattern, "^")

	regex, err := regexp.Compile("^" + cleanPattern)
	if err != nil {
		panic(err)
	}

	remaining := s.str[s.index:]
	loc := regex.FindStringIndex(remaining)
	if loc == nil {
		return nil
	}

	match := remaining[loc[0]:loc[1]]
	return &ParseResult[string]{
		value:  match,
		source: NewSource(s.str, s.index+len(match)),
	}
}

type Parser[T any] struct {
	Parse func(*Source) *ParseResult[T]
}

func Regexp(pattern string) Parser[string] {
	return Parser[string]{func(source *Source) *ParseResult[string] {
		return source.Match(pattern)
	}}
}

func Constant[T any](value T) Parser[T] {
	return Parser[T]{func(source *Source) *ParseResult[T] {
		return &ParseResult[T]{value: value, source: source}
	}}
}

func Error[T any](message string) Parser[T] {
	return Parser[T]{func(source *Source) *ParseResult[T] {
		panic(errors.New(message))
	}}
}

func Or[T any](parsers ...Parser[T]) Parser[T] {
	return Parser[T]{func(source *Source) *ParseResult[T] {
		for _, parser := range parsers {
			if parser.Parse != nil {
				if result := parser.Parse(source); result != nil {
					return result
				}
			}
		}
		return nil
	}}
}

func Many[T any](parser Parser[T]) Parser[[]T] {
	return Parser[[]T]{func(source *Source) *ParseResult[[]T] {
		results := []T{}
		currentSource := source

		for {
			if parser.Parse == nil {
				break
			}
			result := parser.Parse(currentSource)
			if result == nil {
				break
			}
			results = append(results, result.value)
			currentSource = result.source
		}

		return &ParseResult[[]T]{value: results, source: currentSource}
	}}
}

func Bind[T, U any](parser Parser[T], callback func(T) Parser[U]) Parser[U] {
	return Parser[U]{func(source *Source) *ParseResult[U] {
		result := parser.Parse(source)
		if result == nil {
			return nil
		}
		nextParser := callback(result.value)
		if nextParser.Parse == nil {
			return nil
		}
		return nextParser.Parse(result.source)
	}}
}

func And[T, U any](parser1 Parser[T], parser2 Parser[U]) Parser[U] {
	return Bind(parser1, func(_ T) Parser[U] {
		return parser2
	})
}

func Map[T, U any](parser Parser[T], callback func(T) U) Parser[U] {
	return Bind(parser, func(value T) Parser[U] {
		return Constant(callback(value))
	})
}

func Maybe[T any](parser Parser[T]) Parser[*T] {
	return Parser[*T]{func(source *Source) *ParseResult[*T] {
		if parser.Parse == nil {
			return &ParseResult[*T]{value: nil, source: source}
		}
		result := parser.Parse(source)
		if result == nil {
			return &ParseResult[*T]{value: nil, source: source}
		}
		return &ParseResult[*T]{value: &result.value, source: result.source}
	}}
}

func (p Parser[T]) ParseStringToCompletion(str string) T {
	source := NewSource(str, 0)
	if p.Parse == nil {
		panic("Parse error: parser has nil Parse function")
	}
	result := p.Parse(source)
	if result == nil {
		panic("Parse error: could not parse anything at all")
	}
	if result.source.index != len(result.source.str) {
		panic(fmt.Sprintf("Parse error at index %d, remaining: %s",
			result.source.index, result.source.str[result.source.index:]))
	}
	return result.value
}
