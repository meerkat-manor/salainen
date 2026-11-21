package generate

import (
	crypto_rand "crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

type GenerateParameters struct {
	MinimumLength int
	MaximumLength int
	ContentSet    string
}

const numberchar = "0123456789"
const lowerchar = "abcdefghijklmnopqrstuvwxyz"
const upperchar = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const mixedchar = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const alphanumeric_mixed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const alphasymbols_mixed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-={}[]<>?:;"
const alphasymbols = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-={}[]<>?:;"

func AuthenticationSecret(parameters *GenerateParameters) (string, error) {

	genParms := GenerateParameters{
		MinimumLength: 10,
		MaximumLength: 20,
		ContentSet:    "ALPHASYMBOLSMIXED",
	}
	if parameters != nil {
		genParms = *parameters
	}

	if genParms.MinimumLength < 1 {
		return "", fmt.Errorf("invalid minimum length")
	}
	if genParms.MaximumLength < 1 || genParms.MaximumLength > 50 {
		return "", fmt.Errorf("invalid maximum length")
	}
	if genParms.MaximumLength < genParms.MinimumLength {
		return "", fmt.Errorf("maximum length must be equal to or greater than minimum length")
	}

	if genParms.ContentSet == "" {
		genParms.ContentSet = "ALPHASYMBOLSMIXED"
	}

	return GenerateCredential(genParms.ContentSet, genParms.MaximumLength)
}

func random64(max int) (int64, error) {
	if max <= 0 {
		return 0, fmt.Errorf("can't define input as <=0")
	}
	nbig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return nbig.Int64(), nil
}

func GenerateCredential(credentialType string, codeLength int) (string, error) {

	if codeLength < 1 || codeLength > 50 {
		return "", fmt.Errorf("generate length is invalid.  The value needs to be between 1 ands 50 inclusive")
	}

	switch strings.ToUpper(credentialType) {
	case "DIGIT":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			ix, err := random64(len(numberchar))
			if err != nil {
				return "", err
			}
			sb.WriteByte(numberchar[ix])
		}
		return sb.String(), nil
	case "ALPHAUPPER":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			ix, err := random64(len(upperchar))
			if err != nil {
				return "", err
			}
			sb.WriteByte(upperchar[ix])
		}
		return sb.String(), nil
	case "ALPHALOWER":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			ix, err := random64(len(lowerchar))
			if err != nil {
				return "", err
			}
			sb.WriteByte(lowerchar[ix])
		}
		return sb.String(), nil
	case "ALPHANUMERIC":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			ix, err := random64(len(alphanumeric))
			if err != nil {
				return "", err
			}
			sb.WriteByte(alphanumeric[ix])
		}
		return sb.String(), nil
	case "ALPHANUMERICMIXED":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			ix, err := random64(len(alphanumeric_mixed))
			if err != nil {
				return "", err
			}
			sb.WriteByte(alphanumeric_mixed[ix])
		}
		return sb.String(), nil
	case "ALPHASYMBOLS":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			ix, err := random64(len(alphasymbols))
			if err != nil {
				return "", err
			}
			sb.WriteByte(alphasymbols[ix])
		}
		return sb.String(), nil
	case "ALPHASYMBOLSMIXED":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			ix, err := random64(len(alphasymbols_mixed))
			if err != nil {
				return "", err
			}
			sb.WriteByte(alphasymbols_mixed[ix])
		}
		return sb.String(), nil
	case "MIXEDALPHA":
		fallthrough
	case "ALPHA":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			ix, err := random64(len(mixedchar))
			if err != nil {
				return "", err
			}
			sb.WriteByte(mixedchar[ix])
		}
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unknown type of credential to generate")
	}
}
