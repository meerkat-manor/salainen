package generate

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type GenerateParameters struct {
	MinimumLength int
	MaximumLength int
	ContentSet    string
}

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

func GenerateCredential(credentialType string, codeLength int) (string, error) {

	seed := time.Now().UnixNano()
	rand.New(rand.NewSource(seed))

	switch strings.ToUpper(credentialType) {
	case "DIGIT":
		iCode := rand.Int63() // Generates a random int64 number
		code := fmt.Sprintf("%d", iCode)
		return code[:codeLength], nil
	case "ALPHAUPPER":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			sb.WriteByte(upperchar[rand.Intn(len(upperchar))])
		}
		return sb.String(), nil
	case "ALPHALOWER":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			sb.WriteByte(lowerchar[rand.Intn(len(lowerchar))])
		}
		return sb.String(), nil
	case "ALPHANUMERIC":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			sb.WriteByte(alphanumeric[rand.Intn(len(alphanumeric))])
		}
		return sb.String(), nil
	case "ALPHANUMERICMIXED":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			sb.WriteByte(alphanumeric_mixed[rand.Intn(len(alphanumeric_mixed))])
		}
		return sb.String(), nil
	case "ALPHASYMBOLS":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			sb.WriteByte(alphasymbols[rand.Intn(len(alphasymbols))])
		}
		return sb.String(), nil
	case "ALPHASYMBOLSMIXED":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			sb.WriteByte(alphasymbols_mixed[rand.Intn(len(alphasymbols_mixed))])
		}
		return sb.String(), nil
	case "MIXEDALPHA":
		fallthrough
	case "ALPHA":
		sb := strings.Builder{}
		sb.Grow(codeLength)
		for i := 0; i < codeLength; i++ {
			sb.WriteByte(mixedchar[rand.Intn(len(mixedchar))])
		}
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unknown type of credential to generate")
	}
}
