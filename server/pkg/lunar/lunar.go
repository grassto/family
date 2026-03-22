// Package lunar provides Chinese Lunar Calendar conversion utilities.
// Converts between solar (Gregorian) and lunar dates.
package lunar

import (
	"fmt"
	"time"
)

// lunarInfo contains lunar calendar data from 1900 to 2100.
// Each entry encodes: days in year (12/13 months), leap month, and days per month.
// Bit format: 0xAAAABBCC
//
//	AAAA = number of days in each month (12/13 bits, 1=30, 0=29)
//	BB   = leap month (0 = no leap, 1-12 = which month is leap)
//	CC   = days in leap month (0=29, 1=30)
var lunarInfo = []uint32{
	0x04bd8, 0x04ae0, 0x0a570, 0x054d5, 0x0d260, 0x0d950, 0x16554, 0x056a0, 0x09ad0, 0x055d2, // 1900-1909
	0x04ae0, 0x0a5b6, 0x0a4d0, 0x0d250, 0x1d255, 0x0b540, 0x0d6a0, 0x0ada2, 0x095b0, 0x14977, // 1910-1919
	0x04970, 0x0a4b0, 0x0b4b5, 0x06a50, 0x06d40, 0x1ab54, 0x02b60, 0x09570, 0x052f2, 0x04970, // 1920-1929
	0x06566, 0x0d4a0, 0x0ea50, 0x16a95, 0x05ad0, 0x02b60, 0x186e3, 0x092e0, 0x1c8d7, 0x0c950, // 1930-1939
	0x0d4a0, 0x1d8a6, 0x0b550, 0x056a0, 0x1a5b4, 0x025d0, 0x092d0, 0x0d2b2, 0x0a950, 0x0b557, // 1940-1949
	0x06ca0, 0x0b550, 0x15355, 0x04da0, 0x0a5b0, 0x14573, 0x052b0, 0x0a9a8, 0x0e950, 0x06aa0, // 1950-1959
	0x0aea6, 0x0ab50, 0x04b60, 0x0aae4, 0x0a570, 0x05260, 0x0f263, 0x0d950, 0x05b57, 0x056a0, // 1960-1969
	0x096d0, 0x04dd5, 0x04ad0, 0x0a4d0, 0x0d4d4, 0x0d250, 0x0d558, 0x0b540, 0x0b6a0, 0x195a6, // 1970-1979
	0x095b0, 0x049b0, 0x0a974, 0x0a4b0, 0x0b27a, 0x06a50, 0x06d40, 0x0af46, 0x0ab60, 0x09570, // 1980-1989
	0x04af5, 0x04970, 0x064b0, 0x074a3, 0x0ea50, 0x06b58, 0x05ac0, 0x0ab60, 0x096d5, 0x092e0, // 1990-1999
	0x0c960, 0x0d954, 0x0d4a0, 0x0da50, 0x07552, 0x056a0, 0x0abb7, 0x025d0, 0x092d0, 0x0cab5, // 2000-2009
	0x0a950, 0x0b4a0, 0x0baa4, 0x0ad50, 0x055d9, 0x04ba0, 0x0a5b0, 0x15176, 0x052b0, 0x0a930, // 2010-2019
	0x07954, 0x06aa0, 0x0ad50, 0x05b52, 0x04b60, 0x0a6e6, 0x0a4e0, 0x0d260, 0x0ea65, 0x0d530, // 2020-2029
	0x05aa0, 0x076a3, 0x096d0, 0x04afb, 0x04ad0, 0x0a4d0, 0x1d0b6, 0x0d250, 0x0d520, 0x0dd45, // 2030-2039
	0x0b5a0, 0x056d0, 0x055b2, 0x049b0, 0x0a577, 0x0a4b0, 0x0aa50, 0x1b255, 0x06d20, 0x0ada0, // 2040-2049
	0x14b63, 0x09370, 0x049f8, 0x04970, 0x064b0, 0x168a6, 0x0ea50, 0x06b20, 0x1a6c4, 0x0aae0, // 2050-2059
	0x092e0, 0x0d2e3, 0x0c960, 0x0d557, 0x0d4a0, 0x0da50, 0x05d55, 0x056a0, 0x0a6d0, 0x055d4, // 2060-2069
	0x052d0, 0x0a9b8, 0x0a950, 0x0b4a0, 0x0b6a6, 0x0ad50, 0x055a0, 0x0aba4, 0x0a5b0, 0x052b0, // 2070-2079
	0x0b273, 0x06930, 0x07337, 0x06aa0, 0x0ad50, 0x14b55, 0x04b60, 0x0a570, 0x054e4, 0x0d160, // 2080-2089
	0x0e968, 0x0d520, 0x0daa0, 0x16aa6, 0x056d0, 0x04ae0, 0x0a9d4, 0x0a2d0, 0x0d150, 0x0f252, // 2090-2099
	0x0d520, // 2100
}

// 天干
var tianGan = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}

// 地支
var diZhi = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// 生肖
var shengXiao = []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}

// 农历月份名
var lunarMonthNames = []string{"正", "二", "三", "四", "五", "六", "七", "八", "九", "十", "冬", "腊"}

// 农历日名
var lunarDayNames = []string{
	"初一", "初二", "初三", "初四", "初五", "初六", "初七", "初八", "初九", "初十",
	"十一", "十二", "十三", "十四", "十五", "十六", "十七", "十八", "十九", "二十",
	"廿一", "廿二", "廿三", "廿四", "廿五", "廿六", "廿七", "廿八", "廿九", "三十",
}

const baseDate = "1900-01-31" // 农历1900年正月初一对应的公历日期

// LunarDate represents a Chinese lunar date.
type LunarDate struct {
	Year   int  `json:"year"`
	Month  int  `json:"month"`
	Day    int  `json:"day"`
	IsLeap bool `json:"is_leap"`
}

// GetLunarYearDays returns the total number of days in a lunar year.
func GetLunarYearDays(year int) int {
	sum := 348 // 12 months * 29 days = 348
	i := 0x8000
	for i > 0x8 {
		if lunarInfo[year-1900]&uint32(i) != 0 {
			sum++
		}
		i >>= 1
	}
	return sum + GetLeapDays(year)
}

// GetLeapMonth returns the leap month number (0 = no leap).
func GetLeapMonth(year int) int {
	return int(lunarInfo[year-1900] & 0xf)
}

// GetLeapDays returns the number of days in the leap month.
func GetLeapDays(year int) int {
	if GetLeapMonth(year) != 0 {
		if lunarInfo[year-1900]&0x10000 != 0 {
			return 30
		}
		return 29
	}
	return 0
}

// GetLunarMonthDays returns the number of days in a specific lunar month.
func GetLunarMonthDays(year int, month int) int {
	if lunarInfo[year-1900]&(0x10000>>uint(month)) != 0 {
		return 30
	}
	return 29
}

// SolarToLunar converts a solar (Gregorian) date to a lunar date.
func SolarToLunar(solarDate time.Time) *LunarDate {
	base, _ := time.Parse("2006-01-02", baseDate)
	offset := int(solarDate.Sub(base).Hours() / 24)

	year := 1900
	for year < 2101 && offset > 0 {
		days := GetLunarYearDays(year)
		offset -= days
		year++
	}
	if offset < 0 {
		year--
		offset += GetLunarYearDays(year)
	}

	leap := GetLeapMonth(year)
	isLeap := false
	month := 1

	for month <= 12 {
		daysInMonth := GetLunarMonthDays(year, month)

		// If there's a leap month after this month, insert it
		if leap > 0 && month == leap+1 && !isLeap {
			month-- // re-process this month as the leap month
			isLeap = true
			daysInMonth = GetLeapDays(year)
		}

		if offset < daysInMonth {
			break
		}
		offset -= daysInMonth

		if isLeap && month == leap {
			isLeap = false
			month++
		} else if isLeap {
			isLeap = false
		} else {
			month++
		}
	}

	return &LunarDate{
		Year:   year,
		Month:  month,
		Day:    offset + 1,
		IsLeap: isLeap,
	}
}

// LunarToSolar converts a lunar date to a solar (Gregorian) date for a given year.
// Returns the corresponding solar date.
func LunarToSolar(lunar *LunarDate) (time.Time, error) {
	base, _ := time.Parse("2006-01-02", baseDate)
	if lunar.Year < 1900 || lunar.Year > 2100 {
		return time.Time{}, fmt.Errorf("lunar year out of range: %d", lunar.Year)
	}

	// Calculate days from base to start of this lunar year
	offset := 0
	for y := 1900; y < lunar.Year; y++ {
		offset += GetLunarYearDays(y)
	}

	// Add days for months before the target month
	leap := GetLeapMonth(lunar.Year)
	month := 1
	for month < lunar.Month {
		offset += GetLunarMonthDays(lunar.Year, month)
		if month == leap {
			offset += GetLeapDays(lunar.Year)
		}
		month++
	}

	// If it's a leap month, add days for the regular month first
	if lunar.IsLeap && leap == lunar.Month {
		offset += GetLunarMonthDays(lunar.Year, lunar.Month)
	}

	// Add days
	offset += lunar.Day - 1

	return base.AddDate(0, 0, offset), nil
}

// GetNextSolarBirthday returns the next solar date for a lunar birthday.
// If the birthday hasn't passed this year, returns this year's date;
// otherwise returns next year's date.
func GetNextSolarBirthday(lunarMonth, lunarDay int, isLeap bool, from time.Time) time.Time {
	// Try this year
	thisYear := &LunarDate{Year: from.Year(), Month: lunarMonth, Day: lunarDay, IsLeap: isLeap}
	solar, err := LunarToSolar(thisYear)
	if err == nil && solar.After(from.AddDate(0, 0, -1)) {
		return solar
	}
	// Try next year
	nextYear := &LunarDate{Year: from.Year() + 1, Month: lunarMonth, Day: lunarDay, IsLeap: isLeap}
	solar, err = LunarToSolar(nextYear)
	if err == nil {
		return solar
	}
	return time.Time{}
}

// Format returns a human-readable string of the lunar date.
func (l *LunarDate) Format() string {
	leap := ""
	if l.IsLeap {
		leap = "闰"
	}
	return fmt.Sprintf("%s%s月%s", leap, lunarMonthNames[l.Month-1], lunarDayNames[l.Day-1])
}

// FormatWithYear returns a full string including the year with zodiac.
func (l *LunarDate) FormatWithYear() string {
	tg := tianGan[(l.Year-4)%10]
	dz := diZhi[(l.Year-4)%12]
	sx := shengXiao[(l.Year-4)%12]
	return fmt.Sprintf("%s%s（%s）年 %s", tg, dz, sx, l.Format())
}

// GetZodiac returns the zodiac animal for the lunar year.
func (l *LunarDate) GetZodiac() string {
	return shengXiao[(l.Year-4)%12]
}

// ParseLunarBirthday parses a lunar birthday string like "06-15" into month and day.
func ParseLunarBirthday(birthday string) (month, day int, err error) {
	n, err := fmt.Sscanf(birthday, "%d-%d", &month, &day)
	if err != nil || n != 2 {
		return 0, 0, fmt.Errorf("invalid lunar birthday format: %s", birthday)
	}
	if month < 1 || month > 12 || day < 1 || day > 30 {
		return 0, 0, fmt.Errorf("invalid lunar birthday: month=%d, day=%d", month, day)
	}
	return month, day, nil
}
