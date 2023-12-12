package idgen

import (
	"fmt"
	gonanoid "github.com/matoous/go-nanoid"
	"time"
)

const (
	ALPHABETALLCASENUM = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	ALPHABET     = "abcdefghijklmnopqrstuvwxyz"
	ALPHABETNUM  = "abcdefghijklmnopqrstuvwxyz1234567890"
	Id_LEN       = 8
	Id_LEN_SHORT = 4
	Id_LEN_LONG  = 16

	TOKEN_PREF_OPENAPI = "ot"
	TOKEN_PREF_USER    = "ut"
)

func GetIdWithPref(pref string) (x string) {
	rand, _ := gonanoid.Generate(ALPHABET, Id_LEN_SHORT)
	return fmt.Sprintf("%s%d%s", pref, time.Now().UnixMilli(), rand)
}

func GetRandStr() string {
	tmp, _ := gonanoid.Generate(ALPHABET, 8)
	return tmp
}

func GetRandStrByTypeSize(typee string, size int) string {
	tmp, _ := gonanoid.Generate(typee, size)
	return tmp
}

func GetRandStrByTypeSizeWithPref(typee string, size int, pref string) string {
	tmp, _ := gonanoid.Generate(typee, size)
	return fmt.Sprintf("%s%s", pref, tmp)
}
