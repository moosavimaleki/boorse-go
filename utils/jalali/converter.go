package jalali

func Gregorian_to_jalali(gy int, gm int, gd int) (int, int, int) {
	var jy, jm, jd, gy2, days int
	var g_d_m = [12]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	if gm > 2 {
		gy2 = gy + 1
	} else {
		gy2 = gy
	}
	days = 355666 + (365 * gy) + ((gy2 + 3) / 4) - ((gy2 + 99) / 100) + ((gy2 + 399) / 400) + gd + g_d_m[gm-1]
	jy = -1595 + (33 * (days / 12053))
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}
	if days < 186 {
		jm = 1 + (days / 31)
		jd = 1 + (days % 31)
	} else {
		jm = 7 + ((days - 186) / 30)
		jd = 1 + ((days - 186) % 30)
	}
	return jy, jm, jd
}

/* Multiple Return Values */
func Jalali_to_gregorian(jy int, jm int, jd int) (int, int, int) {
	var gy, gm, gd, days int
	var sal_a = [13]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	jy += 1595
	days = -355668 + (365 * jy) + ((jy / 33) * 8) + (((jy % 33) + 3) / 4) + jd
	if jm < 7 {
		days += (jm - 1) * 31
	} else {
		days += ((jm - 7) * 30) + 186
	}
	gy = 400 * (days / 146097)
	days %= 146097
	if days > 36524 {
		days--
		gy += 100 * (days / 36524)
		days %= 36524
		if days >= 365 {
			days++
		}
	}
	gy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		gy += (days - 1) / 365
		days = (days - 1) % 365
	}
	gd = days + 1
	if (gy%4 == 0 && gy%100 != 0) || (gy%400 == 0) {
		sal_a[2] = 29
	}
	gm = 0
	for gm < 13 && gd > sal_a[gm] {
		gd -= sal_a[gm]
		gm++
	}
	return gy, gm, gd
}
