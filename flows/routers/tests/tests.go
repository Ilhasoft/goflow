package tests

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/excellent/functions"
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/utils"

	"github.com/nyaruka/phonenumbers"
	"github.com/shopspring/decimal"
)

//------------------------------------------------------------------------------------------
// Mapping
//------------------------------------------------------------------------------------------

func init() {
	// register our router tests as Excellent functions
	for name, testFunc := range XTESTS {
		functions.RegisterXFunction(name, testFunc)
	}
}

// RegisterXTest registers a new router test (and Excellent function)
func RegisterXTest(name string, function functions.XFunction) {
	XTESTS[name] = function
	functions.RegisterXFunction(name, function)
}

// XTESTS is our mapping of the excellent test names to their actual functions
var XTESTS = map[string]functions.XFunction{
	"is_error":  functions.OneArgFunction(IsError),
	"has_value": functions.OneArgFunction(HasValue),

	"has_group":          functions.TwoArgFunction(HasGroup),
	"has_wait_timed_out": functions.OneArgFunction(HasWaitTimedOut),

	"is_text_eq":      functions.TwoTextFunction(IsTextEQ),
	"has_phrase":      functions.TwoTextFunction(HasPhrase),
	"has_only_phrase": functions.TwoTextFunction(HasOnlyPhrase),
	"has_any_word":    functions.TwoTextFunction(HasAnyWord),
	"has_all_words":   functions.TwoTextFunction(HasAllWords),
	"has_beginning":   functions.TwoTextFunction(HasBeginning),
	"has_text":        functions.OneTextFunction(HasText),
	"has_pattern":     functions.TwoTextFunction(HasPattern),

	"has_number":         functions.OneTextFunction(HasNumber),
	"has_number_between": functions.ThreeArgFunction(HasNumberBetween),
	"has_number_lt":      functions.TextAndNumberFunction(HasNumberLT),
	"has_number_lte":     functions.TextAndNumberFunction(HasNumberLTE),
	"has_number_eq":      functions.TextAndNumberFunction(HasNumberEQ),
	"has_number_gte":     functions.TextAndNumberFunction(HasNumberGTE),
	"has_number_gt":      functions.TextAndNumberFunction(HasNumberGT),

	"has_date":    functions.OneTextFunction(HasDate),
	"has_date_lt": functions.TextAndDateFunction(HasDateLT),
	"has_date_eq": functions.TextAndDateFunction(HasDateEQ),
	"has_date_gt": functions.TextAndDateFunction(HasDateGT),

	"has_time":  functions.OneTextFunction(HasTime),
	"has_phone": functions.InitialTextFunction(0, 1, HasPhone),
	"has_email": functions.OneTextFunction(HasEmail),

	"has_state":    functions.OneTextFunction(HasState),
	"has_district": HasDistrict,
	"has_ward":     HasWard,
}

//------------------------------------------------------------------------------------------
// Tests
//------------------------------------------------------------------------------------------

// IsTextEQ returns whether two text values are equal (case sensitive). In the case that they
// are, it will return the text as the match.
//
//   @(is_text_eq("foo", "foo")) -> true
//   @(is_text_eq("foo", "FOO")) -> false
//   @(is_text_eq("foo", "bar")) -> false
//   @(is_text_eq("foo", " foo ")) -> false
//   @(is_text_eq(run.status, "completed")) -> true
//   @(is_text_eq(results.webhook.category, "Success")) -> true
//   @(is_text_eq(results.webhook.category, "Failure")) -> false
//
// @test is_text_eq(text1, text2)
func IsTextEQ(env utils.Environment, text1 types.XText, text2 types.XText) types.XValue {
	if text1.Equals(text2) {
		return NewTrueResult(text1)
	}

	return XFalseResult
}

// IsError returns whether `value` is an error
//
// Note that `contact.fields` and `run.results` are considered dynamic, so it is not an error
// to try to retrieve a value from fields or results which don't exist, rather these return an empty
// value.
//
//   @(is_error(datetime("foo"))) -> true
//   @(is_error(run.not.existing)) -> true
//   @(is_error(contact.fields.unset)) -> true
//   @(is_error("hello")) -> false
//
// @test is_error(value)
func IsError(env utils.Environment, value types.XValue) types.XValue {
	if types.IsXError(value) {
		return NewTrueResult(value)
	}

	return XFalseResult
}

// HasValue returns whether `value` is non-nil and not an error
//
// Note that `contact.fields` and `run.results` are considered dynamic, so it is not an error
// to try to retrieve a value from fields or results which don't exist, rather these return an empty
// value.
//
//   @(has_value(datetime("foo"))) -> false
//   @(has_value(not.existing)) -> false
//   @(has_value(contact.fields.unset)) -> false
//   @(has_value("")) -> false
//   @(has_value("hello")) -> true
//
// @test has_value(value)
func HasValue(env utils.Environment, value types.XValue) types.XValue {
	if types.IsEmpty(value) || types.IsXError(value) {
		return XFalseResult
	}

	return NewTrueResult(value)
}

// HasWaitTimedOut returns whether the last wait timed out.
//
//   @(has_wait_timed_out(run)) -> false
//
// @test has_wait_timed_out(run)
func HasWaitTimedOut(env utils.Environment, value types.XValue) types.XValue {
	// first parameter needs to be a flow run
	run, isRun := value.(flows.FlowRun)
	if !isRun {
		return types.NewXErrorf("must be called with a run as first argument")
	}

	// look to see if the last input event was a message or a timeout
	runEvents := run.Events()
	for e := len(runEvents) - 1; e >= 0; e-- {
		event := runEvents[e]

		_, isTimeout := event.(*events.WaitTimedOutEvent)
		if isTimeout {
			return NewTrueResult(types.NewXDateTime(event.CreatedOn()))
		}

		_, isInput := event.(*events.MsgReceivedEvent)
		if isInput {
			break
		}
	}

	return XFalseResult
}

// HasGroup returns whether the `contact` is part of group with the passed in UUID
//
//   @(has_group(contact, "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d")) -> true
//   @(has_group(contact, "b7cf0d83-f1c9-411c-96fd-c511a4cfa86d").match) -> Testers
//   @(has_group(contact, "97fe7029-3a15-4005-b0c7-277b884fc1d5")) -> false
//
// @test has_group(contact, group_uuid)
func HasGroup(env utils.Environment, arg1 types.XValue, arg2 types.XValue) types.XValue {
	// is the first argument a contact?
	contact, isContact := arg1.(*flows.Contact)
	if !isContact {
		return types.NewXErrorf("must have a contact as its first argument")
	}

	groupUUID, xerr := types.ToXText(env, arg2)
	if xerr != nil {
		return xerr
	}

	// iterate through the groups looking for one with the same UUID as passed in
	group := contact.Groups().FindByUUID(assets.GroupUUID(groupUUID.Native()))
	if group != nil {
		return NewTrueResult(group)
	}

	return XFalseResult
}

// HasPhrase tests whether `phrase` is contained in `text`
//
// The words in the test phrase must appear in the same order with no other words
// in between.
//
//   @(has_phrase("the quick brown fox", "brown fox")) -> true
//   @(has_phrase("the Quick Brown fox", "quick fox")) -> false
//   @(has_phrase("the Quick Brown fox", "")) -> true
//   @(has_phrase("the.quick.brown.fox", "the quick").match) -> the quick
//
// @test has_phrase(text, phrase)
func HasPhrase(env utils.Environment, text types.XText, test types.XText) types.XValue {
	return testStringTokens(env, text, test, hasPhraseTest)
}

// HasAllWords tests whether all the `words` are contained in `text`
//
// The words can be in any order and may appear more than once.
//
//   @(has_all_words("the quick brown FOX", "the fox")) -> true
//   @(has_all_words("the quick brown FOX", "the fox").match) -> the FOX
//   @(has_all_words("the quick brown fox", "red fox")) -> false
//
// @test has_all_words(text, words)
func HasAllWords(env utils.Environment, text types.XText, test types.XText) types.XValue {
	return testStringTokens(env, text, test, hasAllWordsTest)
}

// HasAnyWord tests whether any of the `words` are contained in the `text`
//
// Only one of the words needs to match and it may appear more than once.
//
//   @(has_any_word("The Quick Brown Fox", "fox quick")) -> true
//   @(has_any_word("The Quick Brown Fox", "red fox")) -> true
//   @(has_any_word("The Quick Brown Fox", "red fox").match) -> Fox
//
// @test has_any_word(text, words)
func HasAnyWord(env utils.Environment, text types.XText, test types.XText) types.XValue {
	return testStringTokens(env, text, test, hasAnyWordTest)
}

// HasOnlyPhrase tests whether the `text` contains only `phrase`
//
// The phrase must be the only text in the text to match
//
//   @(has_only_phrase("The Quick Brown Fox", "quick brown")) -> false
//   @(has_only_phrase("Quick Brown", "quick brown")) -> true
//   @(has_only_phrase("the Quick Brown fox", "")) -> false
//   @(has_only_phrase("", "")) -> true
//   @(has_only_phrase("Quick Brown", "quick brown").match) -> Quick Brown
//   @(has_only_phrase("The Quick Brown Fox", "red fox")) -> false
//
// @test has_only_phrase(text, phrase)
func HasOnlyPhrase(env utils.Environment, text types.XText, test types.XText) types.XValue {
	return testStringTokens(env, text, test, hasOnlyPhraseTest)
}

// HasText tests whether there the text has any characters in it
//
//   @(has_text("quick brown")) -> true
//   @(has_text("quick brown").match) -> quick brown
//   @(has_text("")) -> false
//   @(has_text(" \n")) -> false
//   @(has_text(123)) -> true
//   @(has_text(contact.fields.not_set)) -> false
//
// @test has_text(text)
func HasText(env utils.Environment, text types.XText) types.XValue {
	// trim any whitespace
	text = types.NewXText(strings.TrimSpace(text.Native()))

	// if there is anything left then we have text
	if text.Length() > 0 {
		return NewTrueResult(text)
	}

	return XFalseResult
}

// HasBeginning tests whether `text` starts with `beginning`
//
// Both text values are trimmed of surrounding whitespace, but otherwise matching is strict
// without any tokenization.
//
//   @(has_beginning("The Quick Brown", "the quick")) -> true
//   @(has_beginning("The Quick Brown", "the quick").match) -> The Quick
//   @(has_beginning("The Quick Brown", "the   quick")) -> false
//   @(has_beginning("The Quick Brown", "quick brown")) -> false
//
// @test has_beginning(text, beginning)
func HasBeginning(env utils.Environment, text types.XText, beginning types.XText) types.XValue {
	// trim both
	hayStack := strings.TrimSpace(text.Native())
	pinCushion := strings.TrimSpace(beginning.Native())

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
		return NewTrueResult(types.NewXText(segment))
	}

	return XFalseResult
}

// HasPattern tests whether `text` matches the regex `pattern`
//
// Both text values are trimmed of surrounding whitespace and matching is case-insensitive.
//
//   @(has_pattern("Sell cheese please", "buy (\w+)")) -> false
//   @(has_pattern("Buy cheese please", "buy (\w+)")) -> true
//   @(has_pattern("Buy cheese please", "buy (\w+)").match) -> Buy cheese
//
// @test has_pattern(text, pattern)
func HasPattern(env utils.Environment, text types.XText, pattern types.XText) types.XValue {
	regex, err := regexp.Compile("(?mi)" + strings.TrimSpace(pattern.Native()))
	if err != nil {
		return types.NewXErrorf("must be called with a valid regular expression")
	}

	matches := regex.FindStringSubmatch(text.Native())
	if matches != nil {
		extra := make(map[string]string, len(matches))
		for g, group := range matches {
			extra[strconv.Itoa(g)] = group
		}
		return NewTrueResultWithExtra(types.NewXText(matches[0]), extra)
	}

	return XFalseResult
}

// HasNumber tests whether `text` contains a number
//
//   @(has_number("the number is 42")) -> true
//   @(has_number("the number is 42").match) -> 42
//   @(has_number("the number is forty two")) -> false
//
// @test has_number(text)
func HasNumber(env utils.Environment, text types.XText) types.XValue {
	return testNumber(env, text, types.XNumberZero, types.XNumberZero, isNumberTest)
}

// HasNumberBetween tests whether `text` contains a number between `min` and `max` inclusive
//
//   @(has_number_between("the number is 42", 40, 44)) -> true
//   @(has_number_between("the number is 42", 40, 44).match) -> 42
//   @(has_number_between("the number is 42", 50, 60)) -> false
//   @(has_number_between("the number is not there", 50, 60)) -> false
//   @(has_number_between("the number is not there", "foo", 60)) -> ERROR
//
// @test has_number_between(text, min, max)
func HasNumberBetween(env utils.Environment, arg1 types.XValue, arg2 types.XValue, arg3 types.XValue) types.XValue {
	text, xerr := types.ToXText(env, arg1)
	if xerr != nil {
		return xerr
	}
	min, xerr := types.ToXNumber(env, arg2)
	if xerr != nil {
		return xerr
	}
	max, xerr := types.ToXNumber(env, arg3)
	if xerr != nil {
		return xerr
	}

	return testNumber(env, text, min, max, isNumberBetween)
}

// HasNumberLT tests whether `text` contains a number less than `max`
//
//   @(has_number_lt("the number is 42", 44)) -> true
//   @(has_number_lt("the number is 42", 44).match) -> 42
//   @(has_number_lt("the number is 42", 40)) -> false
//   @(has_number_lt("the number is not there", 40)) -> false
//   @(has_number_lt("the number is not there", "foo")) -> ERROR
//
// @test has_number_lt(text, max)
func HasNumberLT(env utils.Environment, text types.XText, num types.XNumber) types.XValue {
	return testNumber(env, text, num, types.XNumberZero, isNumberLT)
}

// HasNumberLTE tests whether `text` contains a number less than or equal to `max`
//
//   @(has_number_lte("the number is 42", 42)) -> true
//   @(has_number_lte("the number is 42", 44).match) -> 42
//   @(has_number_lte("the number is 42", 40)) -> false
//   @(has_number_lte("the number is not there", 40)) -> false
//   @(has_number_lte("the number is not there", "foo")) -> ERROR
//
// @test has_number_lte(text, max)
func HasNumberLTE(env utils.Environment, text types.XText, num types.XNumber) types.XValue {
	return testNumber(env, text, num, types.XNumberZero, isNumberLTE)
}

// HasNumberEQ tests whether `text` contains a number equal to the `value`
//
//   @(has_number_eq("the number is 42", 42)) -> true
//   @(has_number_eq("the number is 42", 42).match) -> 42
//   @(has_number_eq("the number is 42", 40)) -> false
//   @(has_number_eq("the number is not there", 40)) -> false
//   @(has_number_eq("the number is not there", "foo")) -> ERROR
//
// @test has_number_eq(text, value)
func HasNumberEQ(env utils.Environment, text types.XText, num types.XNumber) types.XValue {
	return testNumber(env, text, num, types.XNumberZero, isNumberEQ)
}

// HasNumberGTE tests whether `text` contains a number greater than or equal to `min`
//
//   @(has_number_gte("the number is 42", 42)) -> true
//   @(has_number_gte("the number is 42", 42).match) -> 42
//   @(has_number_gte("the number is 42", 45)) -> false
//   @(has_number_gte("the number is not there", 40)) -> false
//   @(has_number_gte("the number is not there", "foo")) -> ERROR
//
// @test has_number_gte(text, min)
func HasNumberGTE(env utils.Environment, text types.XText, num types.XNumber) types.XValue {
	return testNumber(env, text, num, types.XNumberZero, isNumberGTE)
}

// HasNumberGT tests whether `text` contains a number greater than `min`
//
//   @(has_number_gt("the number is 42", 40)) -> true
//   @(has_number_gt("the number is 42", 40).match) -> 42
//   @(has_number_gt("the number is 42", 42)) -> false
//   @(has_number_gt("the number is not there", 40)) -> false
//   @(has_number_gt("the number is not there", "foo")) -> ERROR
//
// @test has_number_gt(text, min)
func HasNumberGT(env utils.Environment, text types.XText, num types.XNumber) types.XValue {
	return testNumber(env, text, num, types.XNumberZero, isNumberGT)
}

// HasDate tests whether `text` contains a date formatted according to our environment
//
//   @(has_date("the date is 2017-01-15")) -> true
//   @(has_date("the date is 2017-01-15").match) -> 2017-01-15T13:24:30.123456-05:00
//   @(has_date("there is no date here, just a year 2017")) -> false
//
// @test has_date(text)
func HasDate(env utils.Environment, text types.XText) types.XValue {
	return testDate(env, text, types.XDateTimeZero, isDateTest)
}

// HasDateLT tests whether `text` contains a date before the date `max`
//
//   @(has_date_lt("the date is 2017-01-15", "2017-06-01")) -> true
//   @(has_date_lt("the date is 2017-01-15", "2017-06-01").match) -> 2017-01-15T13:24:30.123456-05:00
//   @(has_date_lt("there is no date here, just a year 2017", "2017-06-01")) -> false
//   @(has_date_lt("there is no date here, just a year 2017", "not date")) -> ERROR
//
// @test has_date_lt(text, max)
func HasDateLT(env utils.Environment, text types.XText, date types.XDateTime) types.XValue {
	return testDate(env, text, date, isDateLTTest)
}

// HasDateEQ tests whether `text` a date equal to `date`
//
//   @(has_date_eq("the date is 2017-01-15", "2017-01-15")) -> true
//   @(has_date_eq("the date is 2017-01-15", "2017-01-15").match) -> 2017-01-15T13:24:30.123456-05:00
//   @(has_date_eq("the date is 2017-01-15 15:00", "2017-01-15")) -> true
//   @(has_date_eq("there is no date here, just a year 2017", "2017-06-01")) -> false
//   @(has_date_eq("there is no date here, just a year 2017", "not date")) -> ERROR
//
// @test has_date_eq(text, date)
func HasDateEQ(env utils.Environment, text types.XText, date types.XDateTime) types.XValue {
	return testDate(env, text, date, isDateEQTest)
}

// HasDateGT tests whether `text` a date after the date `min`
//
//   @(has_date_gt("the date is 2017-01-15", "2017-01-01")) -> true
//   @(has_date_gt("the date is 2017-01-15", "2017-01-01").match) -> 2017-01-15T13:24:30.123456-05:00
//   @(has_date_gt("the date is 2017-01-15", "2017-03-15")) -> false
//   @(has_date_gt("there is no date here, just a year 2017", "2017-06-01")) -> false
//   @(has_date_gt("there is no date here, just a year 2017", "not date")) -> ERROR
//
// @test has_date_gt(text, min)
func HasDateGT(env utils.Environment, text types.XText, date types.XDateTime) types.XValue {
	return testDate(env, text, date, isDateGTTest)
}

// HasTime tests whether `text` contains a time.
//
//   @(has_time("the time is 10:30")) -> true
//   @(has_time("the time is 10 PM")) -> true
//   @(has_time("the time is 10:30 PM").match) -> 22:30:00.000000
//   @(has_time("the time is 10:30:45").match) -> 10:30:45.000000
//   @(has_time("there is no time here, just the number 25")) -> false
//
// @test has_time(text)
func HasTime(env utils.Environment, text types.XText) types.XValue {
	t, xerr := types.ToXTime(env, text)
	if xerr == nil {
		return NewTrueResult(t)
	}

	return XFalseResult
}

var emailAddressRE = regexp.MustCompile(`([\pL\pN][-_.\pL\pN]*)@([\pL\pN][-_\pL\pN]*)(\.[\pL\pN][-_\pL\pN]*)+`)

// HasEmail tests whether an email is contained in `text`
//
//   @(has_email("my email is foo1@bar.com, please respond")) -> true
//   @(has_email("my email is foo1@bar.com, please respond").match) -> foo1@bar.com
//   @(has_email("my email is <foo@bar2.com>")) -> true
//   @(has_email("i'm not sharing my email")) -> false
//
// @test has_email(text)
func HasEmail(env utils.Environment, text types.XText) types.XValue {
	// split by whitespace
	email := emailAddressRE.FindString(text.Native())
	if email != "" {
		return NewTrueResult(types.NewXText(email))
	}

	return XFalseResult
}

// HasPhone tests whether `text` contains a phone number. The optional `country_code` argument specifies
// the country to use for parsing.
//
//   @(has_phone("my number is +12067799294")) -> true
//   @(has_phone("my number is 2067799294", "US")) -> true
//   @(has_phone("my number is 206 779 9294", "US").match) -> +12067799294
//   @(has_phone("my number is none of your business", "US")) -> false
//
// @test has_phone(text, country_code)
func HasPhone(env utils.Environment, text types.XText, args ...types.XValue) types.XValue {
	var country types.XText
	var xerr types.XError
	if len(args) == 1 {
		country, xerr = types.ToXText(env, args[0])
		if xerr != nil {
			return xerr
		}
	} else {
		country = types.NewXText(string(env.DefaultCountry()))
	}

	// try to find a phone number
	phone, err := phonenumbers.Parse(text.Native(), country.Native())
	if err != nil {
		return XFalseResult
	}

	if !phonenumbers.IsPossibleNumber(phone) {
		return XFalseResult
	}

	// format as E164 number
	formatted := phonenumbers.Format(phone, phonenumbers.E164)
	return NewTrueResult(types.NewXText(formatted))
}

// HasState tests whether a state name is contained in the `text`
//
//   @(has_state("Kigali")) -> true
//   @(has_state("Boston")) -> false
//   @(has_state("¡Kigali!")) -> true
//   @(has_state("¡Kigali!").match) -> Rwanda > Kigali City
//   @(has_state("I live in Kigali")) -> true
//
// @test has_state(text)
func HasState(env utils.Environment, text types.XText) types.XValue {
	runEnv, _ := env.(flows.RunEnvironment)

	states, err := runEnv.FindLocationsFuzzy(text.Native(), flows.LocationLevelState, nil)
	if err != nil {
		return types.NewXError(err)
	}
	if len(states) > 0 {
		return NewTrueResult(types.NewXText(states[0].Path()))
	}
	return XFalseResult
}

// HasDistrict tests whether a district name is contained in the `text`. If `state` is also provided
// then the returned district must be within that state.
//
//   @(has_district("Gasabo", "Kigali")) -> true
//   @(has_district("I live in Gasabo", "Kigali")) -> true
//   @(has_district("I live in Gasabo", "Kigali").match) -> Rwanda > Kigali City > Gasabo
//   @(has_district("Gasabo", "Boston")) -> false
//   @(has_district("Gasabo")) -> true
//
// @test has_district(text, state)
func HasDistrict(env utils.Environment, args ...types.XValue) types.XValue {
	if len(args) != 1 && len(args) != 2 {
		return types.NewXErrorf("takes one or two arguments, got %d", len(args))
	}

	runEnv, _ := env.(flows.RunEnvironment)

	var text, stateText types.XText
	var xerr types.XError

	// grab the text we will search and the parent state name
	if text, xerr = types.ToXText(env, args[0]); xerr != nil {
		return xerr
	}
	if len(args) == 2 {
		if stateText, xerr = types.ToXText(env, args[1]); xerr != nil {
			return xerr
		}
	}

	states, err := runEnv.FindLocationsFuzzy(stateText.Native(), flows.LocationLevelState, nil)
	if err != nil {
		return types.NewXError(err)
	}
	if len(states) > 0 {
		districts, err := runEnv.FindLocationsFuzzy(text.Native(), flows.LocationLevelDistrict, states[0])
		if err != nil {
			return types.NewXError(err)
		}
		if len(districts) > 0 {
			return NewTrueResult(types.NewXText(districts[0].Path()))
		}
	}

	// try without a parent state - it's ok as long as we get a single match
	if stateText.Empty() {
		districts, err := runEnv.FindLocationsFuzzy(text.Native(), flows.LocationLevelDistrict, nil)
		if err != nil {
			return types.NewXError(err)
		}
		if len(districts) == 1 {
			return NewTrueResult(types.NewXText(districts[0].Path()))
		}
	}

	return XFalseResult
}

// HasWard tests whether a ward name is contained in the `text`
//
//   @(has_ward("Gisozi", "Gasabo", "Kigali")) -> true
//   @(has_ward("I live in Gisozi", "Gasabo", "Kigali")) -> true
//   @(has_ward("I live in Gisozi", "Gasabo", "Kigali").match) -> Rwanda > Kigali City > Gasabo > Gisozi
//   @(has_ward("Gisozi", "Gasabo", "Brooklyn")) -> false
//   @(has_ward("Gisozi", "Brooklyn", "Kigali")) -> false
//   @(has_ward("Brooklyn", "Gasabo", "Kigali")) -> false
//   @(has_ward("Gasabo")) -> false
//   @(has_ward("Gisozi")) -> true
//
// @test has_ward(text, district, state)
func HasWard(env utils.Environment, args ...types.XValue) types.XValue {
	if len(args) != 1 && len(args) != 3 {
		return types.NewXErrorf("takes one or three arguments, got %d", len(args))
	}

	runEnv, _ := env.(flows.RunEnvironment)

	var text, districtText, stateText types.XText
	var xerr types.XError

	// grab the text we will search, as well as the parent district and state names
	if text, xerr = types.ToXText(env, args[0]); xerr != nil {
		return xerr
	}
	if len(args) == 3 {
		if districtText, xerr = types.ToXText(env, args[1]); xerr != nil {
			return xerr
		}
		if stateText, xerr = types.ToXText(env, args[2]); xerr != nil {
			return xerr
		}
	}

	states, err := runEnv.FindLocationsFuzzy(stateText.Native(), flows.LocationLevelState, nil)
	if err != nil {
		return types.NewXError(err)
	}
	if len(states) > 0 {
		districts, err := runEnv.FindLocationsFuzzy(districtText.Native(), flows.LocationLevelDistrict, states[0])
		if err != nil {
			return types.NewXError(err)
		}
		if len(districts) > 0 {
			wards, err := runEnv.FindLocationsFuzzy(text.Native(), flows.LocationLevelWard, districts[0])
			if err != nil {
				return types.NewXError(err)
			}
			if len(wards) > 0 {
				return NewTrueResult(types.NewXText(wards[0].Path()))
			}
		}
	}

	// try without a parent district - it's ok as long as we get a single match
	if districtText.Empty() {
		wards, err := runEnv.FindLocationsFuzzy(text.Native(), flows.LocationLevelWard, nil)
		if err != nil {
			return types.NewXError(err)
		}
		if len(wards) == 1 {
			return NewTrueResult(types.NewXText(wards[0].Path()))
		}
	}

	return XFalseResult
}

//------------------------------------------------------------------------------------------
// Text Test Functions
//------------------------------------------------------------------------------------------

type stringTokenTest func(origHayTokens []string, hayTokens []string, pinTokens []string) XTestResult

func testStringTokens(env utils.Environment, str types.XText, testStr types.XText, testFunc stringTokenTest) types.XValue {
	hayStack := strings.TrimSpace(str.Native())
	needle := strings.TrimSpace(testStr.Native())

	origHays := utils.TokenizeString(hayStack)
	hays := utils.TokenizeString(strings.ToLower(hayStack))
	needles := utils.TokenizeString(strings.ToLower(needle))

	return testFunc(origHays, hays, needles)
}

func hasPhraseTest(origHays []string, hays []string, pins []string) XTestResult {
	if len(pins) == 0 {
		return NewTrueResult(types.XTextEmpty)
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
		return NewTrueResult(types.NewXText(strings.Join(matches, " ")))
	}

	return XFalseResult
}

func hasAllWordsTest(origHays []string, hays []string, pins []string) XTestResult {
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
		return NewTrueResult(types.NewXText(strings.Join(matches, " ")))
	}

	return XFalseResult
}

func hasAnyWordTest(origHays []string, hays []string, pins []string) XTestResult {
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
		return NewTrueResult(types.NewXText(strings.Join(matches, " ")))
	}

	return XFalseResult
}

func hasOnlyPhraseTest(origHays []string, hays []string, pins []string) XTestResult {
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

	return NewTrueResult(types.NewXText(strings.Join(matches, " ")))
}

//------------------------------------------------------------------------------------------
// Numerical Test Functions
//------------------------------------------------------------------------------------------

// ParseDecimalFuzzy parses a decimal from a string
func ParseDecimalFuzzy(val string, format *utils.NumberFormat) (decimal.Decimal, error) {
	// remove digit grouping symbol
	cleaned := strings.Replace(val, format.DigitGroupingSymbol, "", -1)

	// replace non-period decimal symbols
	cleaned = strings.Replace(cleaned, format.DecimalSymbol, ".", -1)

	return decimal.NewFromString(cleaned)
}

type decimalTest func(value decimal.Decimal, test1 decimal.Decimal, test2 decimal.Decimal) bool

func testNumber(env utils.Environment, str types.XText, testNum1 types.XNumber, testNum2 types.XNumber, testFunc decimalTest) types.XValue {
	// create a number finding regex based on current environment
	pattern := regexp.MustCompile(fmt.Sprintf(`[-+]?[\pNlO\%s]+(\%s[\pNlO]+)?`, env.NumberFormat().DigitGroupingSymbol, env.NumberFormat().DecimalSymbol))

	// look for number like things in the input and use the first one that we can actually parse
	for _, value := range pattern.FindAllString(str.Native(), -1) {
		num, err := ParseDecimalFuzzy(value, env.NumberFormat())
		if err == nil {
			if testFunc(num, testNum1.Native(), testNum2.Native()) {
				return NewTrueResult(types.NewXNumber(num))
			}
		}
	}

	return XFalseResult
}

func isNumberTest(value decimal.Decimal, _ decimal.Decimal, _ decimal.Decimal) bool {
	return true
}

func isNumberLT(value decimal.Decimal, test decimal.Decimal, _ decimal.Decimal) bool {
	return value.Cmp(test) < 0
}

func isNumberLTE(value decimal.Decimal, test decimal.Decimal, _ decimal.Decimal) bool {
	return value.Cmp(test) <= 0
}

func isNumberEQ(value decimal.Decimal, test decimal.Decimal, _ decimal.Decimal) bool {
	return value.Cmp(test) == 0
}

func isNumberGTE(value decimal.Decimal, test decimal.Decimal, _ decimal.Decimal) bool {
	return value.Cmp(test) >= 0
}

func isNumberGT(value decimal.Decimal, test decimal.Decimal, _ decimal.Decimal) bool {
	return value.Cmp(test) > 0
}

func isNumberBetween(value decimal.Decimal, test1 decimal.Decimal, test2 decimal.Decimal) bool {
	return value.Cmp(test1) >= 0 && value.Cmp(test2) <= 0
}

//------------------------------------------------------------------------------------------
// Date Test Functions
//------------------------------------------------------------------------------------------

type dateTest func(value time.Time, test time.Time) bool

func testDate(env utils.Environment, str types.XText, testDate types.XDateTime, testFunc dateTest) types.XValue {
	// first parse wiwth time filling which will the rest result
	value, xerr := types.ToXDateTimeWithTimeFill(env, str)

	// but comparsion should be against only the date portions
	valueAsDate := value.Native().Truncate(24 * time.Hour)
	testAsDate := testDate.Native().Truncate(24 * time.Hour)

	if xerr != nil {
		return XFalseResult
	}

	if testFunc(valueAsDate, testAsDate) {
		return NewTrueResult(value)
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
