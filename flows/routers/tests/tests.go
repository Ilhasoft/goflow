package tests

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/nyaruka/goflow/excellent"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/utils"

	"github.com/nyaruka/phonenumbers"
	"github.com/shopspring/decimal"
)

// TODO:
// InterruptTest
// TimeoutTest
// AirtimeStatusTest

//------------------------------------------------------------------------------------------
// Mapping
//------------------------------------------------------------------------------------------

func init() {
	// register our router tests as Excellent functions
	for name, testFunc := range XTESTS {
		excellent.RegisterFunction(name, testFunc)
	}
}

// XTESTS is our mapping of the excellent test names to their actual functions
var XTESTS = map[string]excellent.XFunction{
	"is_error":           IsError,
	"has_value":          HasValue,
	"has_group":          HasGroup,
	"has_wait_timed_out": HasWaitTimedOut,

	"is_string_eq":    IsStringEQ,
	"has_phrase":      HasPhrase,
	"has_only_phrase": HasOnlyPhrase,
	"has_any_word":    HasAnyWord,
	"has_all_words":   HasAllWords,
	"has_beginning":   HasBeginning,
	"has_text":        HasText,
	"has_pattern":     HasPattern,

	"has_number":         HasNumber,
	"has_number_between": HasNumberBetween,
	"has_number_lt":      HasNumberLT,
	"has_number_lte":     HasNumberLTE,
	"has_number_eq":      HasNumberEQ,
	"has_number_gte":     HasNumberGTE,
	"has_number_gt":      HasNumberGT,

	"has_date":    HasDate,
	"has_date_lt": HasDateLT,
	"has_date_eq": HasDateEQ,
	"has_date_gt": HasDateGT,

	"has_phone": HasPhone,
	"has_email": HasEmail,

	"has_state":    HasState,
	"has_district": HasDistrict,
	"has_ward":     HasWard,
}

//------------------------------------------------------------------------------------------
// Tests
//------------------------------------------------------------------------------------------

// IsStringEQ returns whether two strings are equal (case sensitive). In the case that they
// are, it will return the string as the match.
//
//  @(is_string_eq("foo", "foo")) -> true
//  @(is_string_eq("foo", "FOO")) -> false
//  @(is_string_eq("foo", "bar")) -> false
//  @(is_string_eq("foo", " foo ")) -> false
//  @(is_string_eq(run.status, "completed")) -> true
//  @(is_string_eq(child.status, "expired")) -> false
//  @(is_string_eq(webhook.status, "success")) -> true
//  @(is_string_eq(webhook.status, "connection_error")) -> false
//
// @test is_string_eq(run)
func IsStringEQ(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 2 {
		return fmt.Errorf("IS_STRING_EQ takes exactly two arguments, got %d", len(args))
	}

	// both parameters needs to be strings
	string1, err1 := types.ToString(env, args[0])
	string2, err2 := types.ToString(env, args[1])
	if err1 != nil || err2 != nil {
		return fmt.Errorf("IS_STRING_EQ must be called with strings as both arguments, but got '%s' and '%s'", reflect.TypeOf(args[0]), reflect.TypeOf(args[1]))
	}

	if string1 == string2 {
		return XTestResult{true, string1}
	}

	return XFalseResult
}

// IsError returns whether `value` is an error
//
// Note that `contact.fields` and `run.results` are considered dynamic, so it is not an error
// to try to retrieve a value from fields or results which don't exist, rather these return an empty
// value.
//
//   @(is_error(date("foo"))) -> true
//   @(is_error(run.not.existing)) -> true
//   @(is_error(contact.fields.unset)) -> true
//   @(is_error("hello")) -> false
//
// @test is_error(value)
func IsError(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 1 {
		return fmt.Errorf("IS_ERROR takes exactly one argument, got %d", len(args))
	}

	// nil is not an error
	if args[0] == nil {
		return XFalseResult
	}

	err, isErr := args[0].(error)
	if isErr {
		return XTestResult{true, err}
	}

	return XFalseResult
}

// HasValue returns whether `value` is non-nil and not an error
//
// Note that `contact.fields` and `run.results` are considered dynamic, so it is not an error
// to try to retrieve a value from fields or results which don't exist, rather these return an empty
// value.
//
//   @(has_value(date("foo"))) -> false
//   @(has_value(not.existing)) -> false
//   @(has_value(contact.fields.unset)) -> false
//   @(has_value("hello")) -> true
//
// @test has_value(value)
func HasValue(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 1 {
		return fmt.Errorf("HAS_VALUE takes exactly one argument, got %d", len(args))
	}

	// nil is not a value
	if utils.IsNil(args[0]) {
		return XFalseResult
	}

	// error is not a value
	_, isErr := args[0].(error)
	if isErr {
		return XFalseResult
	}

	return XTestResult{true, args[0]}
}

// HasWaitTimedOut returns whether the last wait timed out.
//
//  @(has_wait_timed_out(run)) -> false
//
// @test has_wait_timed_out(run)
func HasWaitTimedOut(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 1 {
		return fmt.Errorf("HAS_WAIT_TIMED_OUT takes exactly one argument, got %d", len(args))
	}

	// first parameter needs to be a flow run
	run, isRun := args[0].(flows.FlowRun)
	if !isRun {
		return fmt.Errorf("HAS_WAIT_TIMED_OUT must be called with a run as first argument")
	}

	if run.Session().Wait() != nil && run.Session().Wait().HasTimedOut() {
		return XTestResult{true, nil}
	}

	return XFalseResult
}

// HasGroup returns whether the `contact` is part of group with the passed in UUID
//
//  @(has_group(contact, "97fe7029-3a15-4005-b0c7-277b884fc1d5")) -> true
//
// @test has_group(contact)
func HasGroup(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 2 {
		return fmt.Errorf("HAS_GROUP takes exactly two arguments, got %d", len(args))
	}

	// is the first argument a contact?
	contact, isContact := args[0].(*flows.Contact)
	if !isContact {
		return fmt.Errorf("HAS_GROUP must have a contact as its first argument")
	}

	groupUUID, err := types.ToString(env, args[1])
	if err != nil {
		return err
	}

	// iterate through the groups looking for one with the same UUID as passed in
	group := contact.Groups().FindByUUID(flows.GroupUUID(groupUUID))
	if group != nil {
		return XTestResult{true, group}
	}

	return XFalseResult
}

// HasPhrase tests whether `phrase` is contained in `string`
//
// The words in the test phrase must appear in the same order with no other words
// in between.
//
//   @(has_phrase("the quick brown fox", "brown fox")) -> true
//   @(has_phrase("the Quick Brown fox", "quick fox")) -> false
//   @(has_phrase("the Quick Brown fox", "")) -> true
//   @(has_phrase("the.quick.brown.fox", "the quick").match) -> "the quick"
//
// @test has_phrase(string, phrase)
func HasPhrase(env utils.Environment, args ...interface{}) interface{} {
	return testStringTokens(env, "HAS_PHRASE", hasPhraseTest, args)
}

// HasAllWords tests whether all the `words` are contained in `string`
//
// The words can be in any order and may appear more than once.
//
//   @(has_all_words("the quick brown FOX", "the fox")) -> true
//   @(has_all_words("the quick brown FOX", "the fox").match) -> "the FOX"
//   @(has_all_words("the quick brown fox", "red fox")) -> false
//
// @test has_all_words(string, words)
func HasAllWords(env utils.Environment, args ...interface{}) interface{} {
	return testStringTokens(env, "HAS_ALL_WORDS", hasAllWordsTest, args)
}

// HasAnyWord tests whether any of the `words` are contained in the `string`
//
// Only one of the words needs to match and it may appear more than once.
//
//  @(has_any_word("The Quick Brown Fox", "fox quick")) -> true
//  @(has_any_word("The Quick Brown Fox", "red fox")) -> true
//  @(has_any_word("The Quick Brown Fox", "red fox").match) -> "Fox"
//
// @test has_any_word(string, words)
func HasAnyWord(env utils.Environment, args ...interface{}) interface{} {
	return testStringTokens(env, "HAS_ANY_WORD", hasAnyWordTest, args)
}

// HasOnlyPhrase tests whether the `string` contains only `phrase`
//
// The phrase must be the only text in the string to match
//
//  @(has_only_phrase("The Quick Brown Fox", "quick brown")) -> false
//  @(has_only_phrase("Quick Brown", "quick brown")) -> true
//  @(has_only_phrase("the Quick Brown fox", "")) -> false
//  @(has_only_phrase("", "")) -> true
//  @(has_only_phrase("Quick Brown", "quick brown").match) -> "Quick Brown"
//  @(has_only_phrase("The Quick Brown Fox", "red fox")) -> false
//
// @test has_only_phrase(string, phrase)
func HasOnlyPhrase(env utils.Environment, args ...interface{}) interface{} {
	return testStringTokens(env, "HAS_ONLY_PHRASE", hasOnlyPhraseTest, args)
}

// HasText tests whether there the string has any characters in it
//
//   @(has_text("quick brown")) -> true
//   @(has_text("quick brown").match) -> "quick brown"
//   @(has_text("")) -> false
//   @(has_text(" \n")) -> false
//   @(has_text(123)) -> true
//
// @test has_text(string)
func HasText(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 1 {
		return fmt.Errorf("HAS_TEXT takes exactly one arguments, got %d", len(args))
	}

	text, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	// trim any whitespace
	text = strings.TrimSpace(text)

	// if there is anything left then we have text
	if len(text) > 0 {
		return XTestResult{true, text}
	}

	return XFalseResult
}

// HasBeginning tests whether `string` starts with `beginning`
//
// Both strings are trimmed of surrounding whitespace, but otherwise matching is strict
// without any tokenization.
//
//   @(has_beginning("The Quick Brown", "the quick")) -> true
//   @(has_beginning("The Quick Brown", "the quick").match) -> "The Quick"
//   @(has_beginning("The Quick Brown", "the   quick")) -> false
//   @(has_beginning("The Quick Brown", "quick brown")) -> false
//
// @test has_beginning(string, beginning)
func HasBeginning(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 2 {
		return fmt.Errorf("HAS_BEGINNING takes exactly two arguments, got %d", len(args))
	}

	hayStack, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	pinCushion, err := types.ToString(env, args[1])
	if err != nil {
		return err
	}

	// trim both
	pinCushion = strings.TrimSpace(pinCushion)
	hayStack = strings.TrimSpace(hayStack)

	// either are empty, no match
	if hayStack == "" || pinCushion == "" {
		return XFalseResult
	}

	// haystack has to be at least length of needle
	if len(hayStack) < len(pinCushion) {
		return XFalseResult
	}

	segment := hayStack[:len(pinCushion)]
	if strings.ToLower(segment) == strings.ToLower(pinCushion) {
		return XTestResult{true, segment}
	}

	return XFalseResult
}

// Returned by the has_pattern test as its match value
type patternMatch struct {
	groups types.Array
}

func newPatternMatch(matches []string) *patternMatch {
	groups := types.NewArray()
	for _, match := range matches {
		groups.Append(match)
	}
	return &patternMatch{groups: groups}
}

// Resolve resolves the given key when this match is referenced in an expression
func (m *patternMatch) Resolve(key string) interface{} {
	switch key {
	case "groups":
		return m.groups
	}

	return fmt.Errorf("no such key '%s' on pattern match", key)
}

// Atomize is called when this object needs to be reduced to a primitive
func (m *patternMatch) Atomize() interface{} {
	return m.groups.Index(0)
}

var _ types.Atomizable = (*patternMatch)(nil)
var _ types.Resolvable = (*patternMatch)(nil)

// HasPattern tests whether `string` matches the regex `pattern`
//
// Both strings are trimmed of surrounding whitespace and matching is case-insensitive.
//
//   @(has_pattern("Sell cheese please", "buy (\w+)")) -> false
//   @(has_pattern("Buy cheese please", "buy (\w+)")) -> true
//   @(has_pattern("Buy cheese please", "buy (\w+)").match) -> "Buy cheese"
//   @(has_pattern("Buy cheese please", "buy (\w+)").match.groups[0]) -> "Buy cheese"
//   @(has_pattern("Buy cheese please", "buy (\w+)").match.groups[1]) -> "cheese"
//
// @test has_pattern(string, pattern)
func HasPattern(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 2 {
		return fmt.Errorf("HAS_PATTERN takes exactly two arguments, got %d", len(args))
	}

	hayStack, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	pattern, err := types.ToString(env, args[1])
	if err != nil {
		return err
	}

	regex, err := regexp.Compile("(?i)" + strings.TrimSpace(pattern))
	if err != nil {
		return fmt.Errorf("HAS_PATTERN must be called with a valid regular expression")
	}

	matches := regex.FindStringSubmatch(strings.TrimSpace(hayStack))
	if matches != nil {
		return XTestResult{true, newPatternMatch(matches)}
	}

	return XFalseResult
}

// HasNumber tests whether `string` contains a number
//
//   @(has_number("the number is 42")) -> true
//   @(has_number("the number is 42").match) -> 42
//   @(has_number("the number is forty two")) -> false
//
// @test has_number(string)
func HasNumber(env utils.Environment, args ...interface{}) interface{} {
	// only one argument for has number
	if len(args) != 1 {
		return fmt.Errorf("HAS_NUMBER takes exactly one arguments, got %d", len(args))
	}

	testArgs := make([]interface{}, 2)
	testArgs[0] = args[0]

	// set our second argument to a dummy, it isn't used but is need to satisfy our interface
	testArgs[1] = "0"

	return testDecimal(env, "HAS_NUMBER", isNumberTest, testArgs)
}

// HasNumberBetween tests whether `string` contains a number between `min` and `max` inclusive
//
//   @(has_number_between("the number is 42", 40, 44)) -> true
//   @(has_number_between("the number is 42", 40, 44).match) -> 42
//   @(has_number_between("the number is 42", 50, 60)) -> false
//   @(has_number_between("the number is not there", 50, 60)) -> false
//   @(has_number_between("the number is not there", "foo", 60)) -> ERROR
//
// @test has_number_between(string, min, max)
func HasNumberBetween(env utils.Environment, args ...interface{}) interface{} {
	// need three arguments, value being tested and min, max
	if len(args) != 3 {
		return fmt.Errorf("HAS_NUMBER_BETWEEN takes exactly three arguments, got %d", len(args))
	}

	values, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	min, err := types.ToDecimal(env, args[1])
	if err != nil {
		return err
	}
	max, err := types.ToDecimal(env, args[2])
	if err != nil {
		return err
	}

	// for each of our values, try to evaluate to a decimal
	for _, value := range strings.Fields(values) {
		decimalValue, err := types.ToDecimal(env, value)
		if err == nil {
			if decimalValue.Cmp(min) >= 0 && decimalValue.Cmp(max) <= 0 {
				return XTestResult{true, decimalValue}
			}
		}
	}
	return XFalseResult
}

// HasNumberLT tests whether `string` contains a number less than `max`
//
//   @(has_number_lt("the number is 42", 44)) -> true
//   @(has_number_lt("the number is 42", 44).match) -> 42
//   @(has_number_lt("the number is 42", 40)) -> false
//   @(has_number_lt("the number is not there", 40)) -> false
//   @(has_number_lt("the number is not there", "foo")) -> ERROR
//
// @test has_number_lt(string, max)
func HasNumberLT(env utils.Environment, args ...interface{}) interface{} {
	return testDecimal(env, "HAS_NUMBER_LT", isNumberLT, args)
}

// HasNumberLTE tests whether `value` contains a number less than or equal to `max`
//
//   @(has_number_lte("the number is 42", 42)) -> true
//   @(has_number_lte("the number is 42", 44).match) -> 42
//   @(has_number_lte("the number is 42", 40)) -> false
//   @(has_number_lte("the number is not there", 40)) -> false
//   @(has_number_lte("the number is not there", "foo")) -> ERROR
//
// @test has_number_lte(string, max)
func HasNumberLTE(env utils.Environment, args ...interface{}) interface{} {
	return testDecimal(env, "HAS_NUMBER_LTE", isNumberLTE, args)
}

// HasNumberEQ tests whether `strung` contains a number equal to the `value`
//
//   @(has_number_eq("the number is 42", 42)) -> true
//   @(has_number_eq("the number is 42", 42).match) -> 42
//   @(has_number_eq("the number is 42", 40)) -> false
//   @(has_number_eq("the number is not there", 40)) -> false
//   @(has_number_eq("the number is not there", "foo")) -> ERROR
//
// @test has_number_eq(string, value)
func HasNumberEQ(env utils.Environment, args ...interface{}) interface{} {
	return testDecimal(env, "HAS_NUMBER_EQ", isNumberEQ, args)
}

// HasNumberGTE tests whether `string` contains a number greater than or equal to `min`
//
//   @(has_number_gte("the number is 42", 42)) -> true
//   @(has_number_gte("the number is 42", 42).match) -> 42
//   @(has_number_gte("the number is 42", 45)) -> false
//   @(has_number_gte("the number is not there", 40)) -> false
//   @(has_number_gte("the number is not there", "foo")) -> ERROR
//
// @test has_number_gte(string, min)
func HasNumberGTE(env utils.Environment, args ...interface{}) interface{} {
	return testDecimal(env, "HAS_NUMBER_GTE", isNumberGTE, args)
}

// HasNumberGT tests whether `string` contains a number greater than `min`
//
//   @(has_number_gt("the number is 42", 40)) -> true
//   @(has_number_gt("the number is 42", 40).match) -> 42
//   @(has_number_gt("the number is 42", 42)) -> false
//   @(has_number_gt("the number is not there", 40)) -> false
//   @(has_number_gt("the number is not there", "foo")) -> ERROR
//
// @test has_number_gt(string, min)
func HasNumberGT(env utils.Environment, args ...interface{}) interface{} {
	return testDecimal(env, "HAS_NUMBER_GT", isNumberGT, args)
}

// HasDate tests whether `string` contains a date formatted according to our environment
//
//   @(has_date("the date is 2017-01-15")) -> true
//   @(has_date("the date is 2017-01-15").match) -> 2017-01-15T00:00:00.000000Z
//   @(has_date("there is no date here, just a year 2017")) -> false
//
// @test has_date(string)
func HasDate(env utils.Environment, args ...interface{}) interface{} {
	// only one argument for has date
	if len(args) != 1 {
		return fmt.Errorf("HAS_DATE takes exactly one arguments, got %d", len(args))
	}

	testArgs := make([]interface{}, 2)
	testArgs[0] = args[0]

	// set our second argument to a dummy, it isn't used but is need to satisfy our interface
	testArgs[1] = time.Now()

	return testDate(env, "HAS_DATE", isDateTest, testArgs)
}

// HasDateLT tests whether `value` contains a date before the date `max`
//
//   @(has_date_lt("the date is 2017-01-15", "2017-06-01")) -> true
//   @(has_date_lt("the date is 2017-01-15", "2017-06-01").match) -> 2017-01-15T00:00:00.000000Z
//   @(has_date_lt("there is no date here, just a year 2017", "2017-06-01")) -> false
//   @(has_date_lt("there is no date here, just a year 2017", "not date")) -> ERROR
//
// @test has_date_lt(string, max)
func HasDateLT(env utils.Environment, args ...interface{}) interface{} {
	return testDate(env, "HAS_DATE_LT", isDateLTTest, args)
}

// HasDateEQ tests whether `string` a date equal to `date`
//
//   @(has_date_eq("the date is 2017-01-15", "2017-01-15")) -> true
//   @(has_date_eq("the date is 2017-01-15", "2017-01-15").match) -> 2017-01-15T00:00:00.000000Z
//   @(has_date_eq("the date is 2017-01-15 15:00", "2017-01-15")) -> false
//   @(has_date_eq("there is no date here, just a year 2017", "2017-06-01")) -> false
//   @(has_date_eq("there is no date here, just a year 2017", "not date")) -> ERROR
//
// @test has_date_eq(string, date)
func HasDateEQ(env utils.Environment, args ...interface{}) interface{} {
	return testDate(env, "HAS_DATE_EQ", isDateEQTest, args)
}

// HasDateGT tests whether `string` a date after the date `min`
//
//   @(has_date_gt("the date is 2017-01-15", "2017-01-01")) -> true
//   @(has_date_gt("the date is 2017-01-15", "2017-01-01").match) -> 2017-01-15T00:00:00.000000Z
//   @(has_date_gt("the date is 2017-01-15", "2017-03-15")) -> false
//   @(has_date_gt("there is no date here, just a year 2017", "2017-06-01")) -> false
//   @(has_date_gt("there is no date here, just a year 2017", "not date")) -> ERROR
//
// @test has_date_gt(string, min)
func HasDateGT(env utils.Environment, args ...interface{}) interface{} {
	return testDate(env, "HAS_DATE_GT", isDateGTTest, args)
}

var emailAddressRE = regexp.MustCompile(`([\pL\pN][-_.\pL\pN]*)@([\pL\pN][-_\pL\pN]*)(\.[\pL\pN][-_\pL\pN]*)+`)

// HasEmail tests whether an email is contained in `string`
//
//   @(has_email("my email is foo1@bar.com, please respond")) -> true
//   @(has_email("my email is foo1@bar.com, please respond").match) -> "foo1@bar.com"
//   @(has_email("my email is <foo@bar2.com>")) -> true
//   @(has_email("i'm not sharing my email")) -> false
//
// @test has_email(string)
func HasEmail(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 1 {
		return fmt.Errorf("HAS_EMAIL takes exactly one argument, got %d", len(args))
	}

	// convert our arg to a string
	text, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	// split by whitespace
	email := emailAddressRE.FindString(text)
	if email != "" {
		return XTestResult{true, email}
	}

	return XFalseResult
}

// HasPhone tests whether a phone number (in the passed in `country_code`) is contained in the `string`
//
//   @(has_phone("my number is 2067799294", "US")) -> true
//   @(has_phone("my number is 206 779 9294", "US").match) -> "+12067799294"
//   @(has_phone("my number is none of your business", "US")) -> false
//
// @test has_phone(string, country_code)
func HasPhone(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 2 {
		return fmt.Errorf("HAS_PHONE takes exactly two arguments, the string to search and the country code, got %d", len(args))
	}

	// grab the text we will search
	text, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	// and the country we are searching
	country, err := types.ToString(env, args[1])
	if err != nil {
		return err
	}

	// try to find a phone number
	phone, err := phonenumbers.Parse(text, country)
	if err != nil {
		return XFalseResult
	}

	// format as E164 number
	formatted := phonenumbers.Format(phone, phonenumbers.E164)
	return XTestResult{true, formatted}
}

// HasState tests whether a state name is contained in the `string`
//
//   @(has_state("Kigali")) -> true
//   @(has_state("Boston")) -> false
//   @(has_state("¡Kigali!")) -> true
//   @(has_state("I live in Kigali")) -> true
//
// @test has_state(string)
func HasState(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 1 {
		return fmt.Errorf("HAS_STATE takes exactly one arguments, got %d", len(args))
	}

	runEnv, _ := env.(flows.RunEnvironment)

	// grab the text we will search
	text, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	states, err := runEnv.FindLocationsFuzzy(text, flows.LocationLevel(1), nil)
	if err != nil {
		return err
	}
	if len(states) > 0 {
		return XTestResult{true, states[0]}
	}
	return XFalseResult
}

// HasDistrict tests whether a district name is contained in the `string`. If `state` is also provided
// then the returned district must be within that state.
//
//   @(has_district("Gasabo", "Kigali")) -> true
//   @(has_district("I live in Gasabo", "Kigali")) -> true
//   @(has_district("Gasabo", "Boston")) -> false
//   @(has_district("Gasabo")) -> true
//
// @test has_district(string, state)
func HasDistrict(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 1 && len(args) != 2 {
		return fmt.Errorf("HAS_DISTRICT takes one or two arguments, got %d", len(args))
	}

	runEnv, _ := env.(flows.RunEnvironment)

	var text, stateText string
	var err error

	// grab the text we will search and the parent state name
	if text, err = types.ToString(env, args[0]); err != nil {
		return err
	}
	if len(args) == 2 {
		if stateText, err = types.ToString(env, args[1]); err != nil {
			return err
		}
	}

	states, err := runEnv.FindLocationsFuzzy(stateText, flows.LocationLevel(1), nil)
	if err != nil {
		return err
	}
	if len(states) > 0 {
		districts, err := runEnv.FindLocationsFuzzy(text, flows.LocationLevel(2), states[0])
		if err != nil {
			return err
		}
		if len(districts) > 0 {
			return XTestResult{true, districts[0]}
		}
	}

	// try without a parent state - it's ok as long as we get a single match
	if stateText == "" {
		districts, err := runEnv.FindLocationsFuzzy(text, flows.LocationLevel(2), nil)
		if err != nil {
			return err
		}
		if len(districts) == 1 {
			return XTestResult{true, districts[0]}
		}
	}

	return XFalseResult
}

// HasWard tests whether a ward name is contained in the `string`
//
//   @(has_ward("Gisozi", "Gasabo", "Kigali")) -> true
//   @(has_ward("I live in Gisozi", "Gasabo", "Kigali")) -> true
//   @(has_ward("Gisozi", "Gasabo", "Brooklyn")) -> false
//   @(has_ward("Gisozi", "Brooklyn", "Kigali")) -> false
//   @(has_ward("Brooklyn", "Gasabo", "Kigali")) -> false
//   @(has_ward("Gasabo")) -> false
//   @(has_ward("Gisozi")) -> true
//
// @test has_ward(string, district, state)
func HasWard(env utils.Environment, args ...interface{}) interface{} {
	if len(args) != 1 && len(args) != 3 {
		return fmt.Errorf("HAS_WARD takes one or three arguments, got %d", len(args))
	}

	runEnv, _ := env.(flows.RunEnvironment)

	var text, districtText, stateText string
	var err error

	// grab the text we will search, as well as the parent district and state names
	if text, err = types.ToString(env, args[0]); err != nil {
		return err
	}
	if len(args) == 3 {
		if districtText, err = types.ToString(env, args[1]); err != nil {
			return err
		}
		if stateText, err = types.ToString(env, args[2]); err != nil {
			return err
		}
	}

	states, err := runEnv.FindLocationsFuzzy(stateText, flows.LocationLevel(1), nil)
	if err != nil {
		return err
	}
	if len(states) > 0 {
		districts, err := runEnv.FindLocationsFuzzy(districtText, flows.LocationLevel(2), states[0])
		if err != nil {
			return err
		}
		if len(districts) > 0 {
			wards, err := runEnv.FindLocationsFuzzy(text, flows.LocationLevel(3), districts[0])
			if err != nil {
				return err
			}
			if len(wards) > 0 {
				return XTestResult{true, wards[0]}
			}
		}
	}

	// try without a parent district - it's ok as long as we get a single match
	if districtText == "" {
		wards, err := runEnv.FindLocationsFuzzy(text, flows.LocationLevel(3), nil)
		if err != nil {
			return err
		}
		if len(wards) == 1 {
			return XTestResult{true, wards[0]}
		}
	}

	return XFalseResult
}

//------------------------------------------------------------------------------------------
// String Test Functions
//------------------------------------------------------------------------------------------

type stringTokenTest func(origHayTokens []string, hayTokens []string, pinTokens []string) interface{}

func testStringTokens(env utils.Environment, name string, test stringTokenTest, args []interface{}) interface{} {
	if len(args) != 2 {
		return fmt.Errorf("%s takes exactly two arguments, got %d", name, len(args))
	}

	hayStack, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	pinCushion, err := types.ToString(env, args[1])
	if err != nil {
		return err
	}

	hayStack = strings.TrimSpace(hayStack)
	pinCushion = strings.TrimSpace(pinCushion)

	origHays := utils.TokenizeString(hayStack)
	hays := utils.TokenizeString(strings.ToLower(hayStack))
	pins := utils.TokenizeString(strings.ToLower(pinCushion))

	return test(origHays, hays, pins)
}

func hasPhraseTest(origHays []string, hays []string, pins []string) interface{} {
	if len(pins) == 0 {
		return XTestResult{true, ""}
	}

	pinIdx := 0
	matches := make([]string, len(pins))
	for i, hay := range hays {
		if hay == pins[pinIdx] {
			matches[pinIdx] = origHays[i]
			pinIdx++
			if pinIdx == len(pins) {
				break
			}
		} else {
			pinIdx = 0
		}
	}

	if pinIdx == len(pins) {
		return XTestResult{true, strings.Join(matches, " ")}
	}

	return XFalseResult
}

func hasAllWordsTest(origHays []string, hays []string, pins []string) interface{} {
	matches := make([]string, 0, len(pins))
	pinMatches := make([]int, len(pins))

	for i, hay := range hays {
		matched := false
		for j, pin := range pins {
			if hay == pin {
				matched = true
				pinMatches[j]++
			}
		}

		if matched {
			matches = append(matches, origHays[i])
		}
	}

	allMatch := true
	for _, matchCount := range pinMatches {
		if matchCount == 0 {
			allMatch = false
			break
		}

	}

	if allMatch {
		return XTestResult{true, strings.Join(matches, " ")}
	}

	return XFalseResult
}

func hasAnyWordTest(origHays []string, hays []string, pins []string) interface{} {
	matches := make([]string, 0, len(pins))
	for i, hay := range hays {
		matched := false
		for _, pin := range pins {
			if hay == pin {
				matched = true
				break
			}
		}
		if matched {
			matches = append(matches, origHays[i])
		}

	}

	if len(matches) > 0 {
		return XTestResult{true, strings.Join(matches, " ")}
	}

	return XFalseResult
}

func hasOnlyPhraseTest(origHays []string, hays []string, pins []string) interface{} {
	// must be same length
	if len(hays) != len(pins) {
		return XFalseResult
	}

	// and every token must match
	matches := make([]string, 0, len(pins))
	for i := range hays {
		if hays[i] != pins[i] {
			return XFalseResult
		}
		matches = append(matches, origHays[i])
	}

	return XTestResult{true, strings.Join(matches, " ")}
}

//------------------------------------------------------------------------------------------
// Decimal Test Functions
//------------------------------------------------------------------------------------------

type decimalTest func(value decimal.Decimal, test decimal.Decimal) bool

func testDecimal(env utils.Environment, name string, test decimalTest, args []interface{}) interface{} {
	if len(args) != 2 {
		return fmt.Errorf("%s takes exactly two arguments, got %d", name, len(args))
	}

	values, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	decimalTest, err := types.ToDecimal(env, args[1])
	if err != nil {
		return err
	}

	// for each of our values, try to evaluate to a decimal
	for _, value := range strings.Fields(values) {
		decimalValue, err := types.ToDecimal(env, value)
		if err == nil {
			if test(decimalValue, decimalTest) {
				return XTestResult{true, decimalValue}
			}
		}
	}

	return XFalseResult
}

func isNumberTest(value decimal.Decimal, test decimal.Decimal) bool {
	return true
}

func isNumberLT(value decimal.Decimal, test decimal.Decimal) bool {
	return value.Cmp(test) < 0
}

func isNumberLTE(value decimal.Decimal, test decimal.Decimal) bool {
	return value.Cmp(test) <= 0
}

func isNumberEQ(value decimal.Decimal, test decimal.Decimal) bool {
	return value.Cmp(test) == 0
}

func isNumberGTE(value decimal.Decimal, test decimal.Decimal) bool {
	return value.Cmp(test) >= 0
}

func isNumberGT(value decimal.Decimal, test decimal.Decimal) bool {
	return value.Cmp(test) > 0
}

//------------------------------------------------------------------------------------------
// Date Test Functions
//------------------------------------------------------------------------------------------

type dateTest func(value time.Time, test time.Time) bool

func testDate(env utils.Environment, name string, test dateTest, args []interface{}) interface{} {
	if len(args) != 2 {
		return fmt.Errorf("%s takes exactly two arguments, got %d", name, len(args))
	}

	// if we can't convert this to a string, then that's an error
	_, err := types.ToString(env, args[0])
	if err != nil {
		return err
	}

	// error is if we don't find a date on our test value, that's ok but no match
	value, err := types.ToDate(env, args[0])
	if err != nil {
		return XFalseResult
	}

	dateTest, err := types.ToDate(env, args[1])
	if err != nil {
		return err
	}

	if test(value, dateTest) {
		return XTestResult{true, value}
	}

	return XFalseResult
}

func isDateTest(value time.Time, test time.Time) bool {
	return true
}

func isDateLTTest(value time.Time, test time.Time) bool {
	return value.Before(test)
}

func isDateEQTest(value time.Time, test time.Time) bool {
	return value.Equal(test)
}

func isDateGTTest(value time.Time, test time.Time) bool {
	return value.After(test)
}
