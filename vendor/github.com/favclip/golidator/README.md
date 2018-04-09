# golidator

The programable validator.

## Description

Existing validator is not fully flexible.
We want to customize error object generation.
We want to modify value instead of raise error when validation failed.
We want it.

## Samples

see [usage](https://github.com/favclip/golidator/blob/master/usage_test.go)

### Basic usage

```
v := golidator.NewValidator()
err := v.Validate(obj)
```

### Use Custom Validator

```
v := golidator.NewValidator()
v.SetValidationFunc("req", func(param string, val reflect.Value) (golidator.ValidationResult, error) {
    if str := val.String(); str == "" {
        return golidator.ValidationNG, nil
    }

    return golidator.ValidationOK, nil
})
```
