package checker

import "fmt"

func stringList(list []string, cap int) string {
	var result string
	n := len(list)
	if n < 1 {
		result = "Пусто"
		return result
	}
	if n > cap {
		result = fmt.Sprintf("Слишком много адресов в списке \\(%d\\), не буду их все выводить", n)
		return result
	}
	var s string
	for _, a := range list {
		s += fmt.Sprintf("`%s`\n", a)
	}
	result = s
	return result
}
