package grammar

func GetCodeFromBool(b bool) (x int) {
	if b {
		return 1
	} else {
		return 0
	}
}

func GetBoolFromCode(x int) (b bool) {
	if x == 1 {
		return true
	} else {
		return false
	}
}
