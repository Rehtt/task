package internal

func StringMust(str string, err error) string {
	if err != nil {
		return ""
	}
	return str
}
