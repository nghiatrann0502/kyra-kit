package validator

import (
	"context"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/vi"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	viTranslations "github.com/go-playground/validator/v10/translations/vi"
	"github.com/nghiatrann0502/clinic/pkg/errors"
)

type Validator interface {
	GetClient() *validator.Validate
	GetTranslator() *ut.UniversalTranslator
	Struct(ctx context.Context, dto any) error
}

type validatorImpl struct {
	validate   *validator.Validate
	translator *ut.UniversalTranslator
}

// Struct implements Validator.
func (v *validatorImpl) Struct(ctx context.Context, dto any) error {
	langCode := "en"
	// lang, ok := ctx.Value(contextutils.I18NFromContext(ctx)).(string)
	// if ok {
	// 	langCode = lang
	// }
	uni := v.translator
	trans, found := uni.GetTranslator(langCode)
	if !found {
		trans, _ = uni.GetTranslator("en")
	}

	err := v.validate.Struct(dto)
	if err != nil {
		validateErr := errors.New(errors.ErrCodeInvalidInput, "validate errors")

		errs := err.(validator.ValidationErrors)
		mappingErrs := make(map[string]string)
		for _, fieldErr := range errs {
			mappingErrs[fieldErr.Namespace()] = fieldErr.Field()
		}

		translatedErrs := errs.Translate(trans)

		for field, msg := range translatedErrs {
			validateErr.WithField(mappingErrs[field], msg)
		}

		return validateErr
	}

	return nil
}

func (v *validatorImpl) GetClient() *validator.Validate {
	return v.validate
}

func (v *validatorImpl) GetTranslator() *ut.UniversalTranslator {
	return v.translator
}

func New() Validator {
	validate := validator.New()
	enLocale := en.New()
	viLocale := vi.New()

	uni := ut.New(viLocale, enLocale, viLocale)

	enTrans, _ := uni.GetTranslator("en")
	enTranslations.RegisterDefaultTranslations(validate, enTrans)

	viTrans, _ := uni.GetTranslator("vi")
	viTranslations.RegisterDefaultTranslations(validate, viTrans)

	return &validatorImpl{
		validate:   validate,
		translator: uni,
	}
}
