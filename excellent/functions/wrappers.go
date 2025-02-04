package functions

import (
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/utils"
)

// ArgCountCheck wraps an XFunction and checks the number of args
func ArgCountCheck(min int, max int, f XFunction) XFunction {
	return func(env utils.Environment, args ...types.XValue) types.XValue {
		if min == max {
			// function requires a fixed number of arguments
			if len(args) != min {
				return types.NewXErrorf("need %d argument(s), got %d", min, len(args))
			}
		} else if max < 0 {
			// function requires a minimum number of arguments
			if len(args) < min {
				return types.NewXErrorf("need at least %d argument(s), got %d", min, len(args))
			}
		} else {
			// function requires the given range of arguments
			if len(args) < min || len(args) > max {
				return types.NewXErrorf("need %d to %d argument(s), got %d", min, max, len(args))
			}
		}

		return f(env, args...)
	}
}

// NoArgFunction creates an XFunction from a no-arg function
func NoArgFunction(f func(utils.Environment) types.XValue) XFunction {
	return ArgCountCheck(0, 0, func(env utils.Environment, args ...types.XValue) types.XValue {
		return f(env)
	})
}

// OneArgFunction creates an XFunction from a single-arg function
func OneArgFunction(f func(utils.Environment, types.XValue) types.XValue) XFunction {
	return ArgCountCheck(1, 1, func(env utils.Environment, args ...types.XValue) types.XValue {
		return f(env, args[0])
	})
}

// TwoArgFunction creates an XFunction from a two-arg function
func TwoArgFunction(f func(utils.Environment, types.XValue, types.XValue) types.XValue) XFunction {
	return ArgCountCheck(2, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		return f(env, args[0], args[1])
	})
}

// ThreeArgFunction creates an XFunction from a three-arg function
func ThreeArgFunction(f func(utils.Environment, types.XValue, types.XValue, types.XValue) types.XValue) XFunction {
	return ArgCountCheck(3, 3, func(env utils.Environment, args ...types.XValue) types.XValue {
		return f(env, args[0], args[1], args[2])
	})
}

// OneTextFunction creates an XFunction from a function that takes a single text arg
func OneTextFunction(f func(utils.Environment, types.XText) types.XValue) XFunction {
	return ArgCountCheck(1, 1, func(env utils.Environment, args ...types.XValue) types.XValue {
		str, xerr := types.ToXText(env, args[0])
		if xerr != nil {
			return xerr
		}
		return f(env, str)
	})
}

// TwoTextFunction creates an XFunction from a function that takes two text args
func TwoTextFunction(f func(utils.Environment, types.XText, types.XText) types.XValue) XFunction {
	return ArgCountCheck(2, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		str1, xerr := types.ToXText(env, args[0])
		if xerr != nil {
			return xerr
		}
		str2, xerr := types.ToXText(env, args[1])
		if xerr != nil {
			return xerr
		}
		return f(env, str1, str2)
	})
}

// ThreeTextFunction creates an XFunction from a function that takes three text args
func ThreeTextFunction(f func(utils.Environment, types.XText, types.XText, types.XText) types.XValue) XFunction {
	return ArgCountCheck(3, 3, func(env utils.Environment, args ...types.XValue) types.XValue {
		str1, xerr := types.ToXText(env, args[0])
		if xerr != nil {
			return xerr
		}
		str2, xerr := types.ToXText(env, args[1])
		if xerr != nil {
			return xerr
		}
		str3, xerr := types.ToXText(env, args[2])
		if xerr != nil {
			return xerr
		}
		return f(env, str1, str2, str3)
	})
}

// TextAndNumberFunction creates an XFunction from a function that takes a text and a number arg
func TextAndNumberFunction(f func(utils.Environment, types.XText, types.XNumber) types.XValue) XFunction {
	return ArgCountCheck(2, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		str, xerr := types.ToXText(env, args[0])
		if xerr != nil {
			return xerr
		}
		num, xerr := types.ToXNumber(env, args[1])
		if xerr != nil {
			return xerr
		}

		return f(env, str, num)
	})
}

// TextAndIntegerFunction creates an XFunction from a function that takes a text and an integer arg
func TextAndIntegerFunction(f func(utils.Environment, types.XText, int) types.XValue) XFunction {
	return ArgCountCheck(2, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		str, xerr := types.ToXText(env, args[0])
		if xerr != nil {
			return xerr
		}
		num, xerr := types.ToInteger(env, args[1])
		if xerr != nil {
			return xerr
		}

		return f(env, str, num)
	})
}

// ThreeIntegerFunction creates an XFunction from a function that takes a text and an integer arg
func ThreeIntegerFunction(f func(utils.Environment, int, int, int) types.XValue) XFunction {
	return ArgCountCheck(3, 3, func(env utils.Environment, args ...types.XValue) types.XValue {
		num1, xerr := types.ToInteger(env, args[0])
		if xerr != nil {
			return xerr
		}
		num2, xerr := types.ToInteger(env, args[1])
		if xerr != nil {
			return xerr
		}
		num3, xerr := types.ToInteger(env, args[2])
		if xerr != nil {
			return xerr
		}

		return f(env, num1, num2, num3)
	})
}

// TextAndDateFunction creates an XFunction from a function that takes a text and a date arg
func TextAndDateFunction(f func(utils.Environment, types.XText, types.XDateTime) types.XValue) XFunction {
	return ArgCountCheck(2, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		str, xerr := types.ToXText(env, args[0])
		if xerr != nil {
			return xerr
		}
		date, xerr := types.ToXDateTime(env, args[1])
		if xerr != nil {
			return xerr
		}

		return f(env, str, date)
	})
}

// InitialTextFunction creates an XFunction from a function that takes an initial text arg followed by other args
func InitialTextFunction(minOtherArgs int, maxOtherArgs int, f func(utils.Environment, types.XText, ...types.XValue) types.XValue) XFunction {
	return ArgCountCheck(minOtherArgs+1, maxOtherArgs+1, func(env utils.Environment, args ...types.XValue) types.XValue {
		str, xerr := types.ToXText(env, args[0])
		if xerr != nil {
			return xerr
		}
		return f(env, str, args[1:]...)
	})
}

// OneNumberFunction creates an XFunction from a single number function
func OneNumberFunction(f func(utils.Environment, types.XNumber) types.XValue) XFunction {
	return ArgCountCheck(1, 1, func(env utils.Environment, args ...types.XValue) types.XValue {
		num, xerr := types.ToXNumber(env, args[0])
		if xerr != nil {
			return xerr
		}

		return f(env, num)
	})
}

// OneNumberAndOptionalIntegerFunction creates an XFunction from a function that takes a number and an optional integer
func OneNumberAndOptionalIntegerFunction(f func(utils.Environment, types.XNumber, int) types.XValue, defaultVal int) XFunction {
	return ArgCountCheck(1, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		num, xerr := types.ToXNumber(env, args[0])
		if xerr != nil {
			return xerr
		}

		intVal := defaultVal
		if len(args) == 2 {
			intVal, xerr = types.ToInteger(env, args[1])
			if xerr != nil {
				return xerr
			}
		}

		return f(env, num, intVal)
	})
}

// TwoNumberFunction creates an XFunction from a function that takes two numbers
func TwoNumberFunction(f func(utils.Environment, types.XNumber, types.XNumber) types.XValue) XFunction {
	return ArgCountCheck(2, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		num1, xerr := types.ToXNumber(env, args[0])
		if xerr != nil {
			return xerr
		}
		num2, xerr := types.ToXNumber(env, args[1])
		if xerr != nil {
			return xerr
		}

		return f(env, num1, num2)
	})
}

// OneDateFunction creates an XFunction from a single date function
func OneDateFunction(f func(utils.Environment, types.XDate) types.XValue) XFunction {
	return ArgCountCheck(1, 1, func(env utils.Environment, args ...types.XValue) types.XValue {
		date, xerr := types.ToXDate(env, args[0])
		if xerr != nil {
			return xerr
		}

		return f(env, date)
	})
}

// OneDateTimeFunction creates an XFunction from a single datetime function
func OneDateTimeFunction(f func(utils.Environment, types.XDateTime) types.XValue) XFunction {
	return ArgCountCheck(1, 1, func(env utils.Environment, args ...types.XValue) types.XValue {
		date, xerr := types.ToXDateTime(env, args[0])
		if xerr != nil {
			return xerr
		}

		return f(env, date)
	})
}
