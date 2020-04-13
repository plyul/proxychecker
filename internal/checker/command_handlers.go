package checker

import (
	"fmt"
	"time"
)

func handleHelp() string {
	return "/getnum \\- вывести текущее количество адресов в списке\n" +
		"/listall \\- вывести список всех адресов\n" +
		"/checkall \\- запустить проверку всех адресов\n" +
		"/listfailed \\- вывести список адресов, не прошедших проверку\n" +
		"/checkfailed \\- запустить проверку адресов, не прошедших проверку\n" +
		"/stats \\- вывести статистику\n"
}

func (pc *proxyChecker) handleGetNum() string {
	return fmt.Sprintf("%d \\(ограничение на проверку: %d\\)", len(pc.proxyList), pc.maxProxiesToCheck)
}

func (pc *proxyChecker) handleListAll() string {
	return stringList(pc.proxyList, pc.maxAddressesToOutput)
}

func (pc *proxyChecker) handleListFailed() string {
	return stringList(pc.stats.failedAddresses, pc.maxAddressesToOutput)
}

func (pc *proxyChecker) handleCheck(list []string) string {
	pc.locker.Lock()
	if pc.busy {
		return "Проверка в процессе выполнения"
	}
	pc.locker.Unlock()
	if len(list) > 0 {
		go pc.checkProxies(list)
		return "Проверка запущена"
	} else {
		return "Нечего проверять"
	}
}

func (pc *proxyChecker) handleStats() string {
	if !pc.stats.valid {
		return "Проверка ни разу не запускалась, не о чём отчитываться"
	}
	s := fmt.Sprintf("Проверка начата: `%s`\n", pc.stats.startTime.Format("15:04"))
	if pc.stats.checkProgress == 1.0 {
		d := pc.stats.endTime.Sub(pc.stats.startTime)
		mins := d.Round(time.Minute)
		secs := (d - mins).Round(time.Second)
		s += fmt.Sprintf("Длительность: `%02.0f:%02.0f`\n", mins.Minutes(), secs.Seconds())
	} else {
		s += fmt.Sprintf("Прогресс: `%.0f%%`\n", pc.stats.checkProgress*100)
	}
	s += fmt.Sprintf("Всего: `%d`\n", pc.stats.counterTotal)
	s += fmt.Sprintf("Успешно: `%d`\n", pc.stats.counterOk)
	s += fmt.Sprintf("Неуспешно: `%d`\n", pc.stats.counterFail)
	s += "Сломанные прокси:\n"
	s += stringList(pc.stats.failedAddresses, pc.maxAddressesToOutput)
	return s
}
