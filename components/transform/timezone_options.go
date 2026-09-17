package transform

import "github.com/flowgo/flowgo/api/types"

// timezoneSelectOptions 时区下拉（常用 IANA；默认东八区排最前）。
func timezoneSelectOptions() []types.ConfigFieldOption {
	opt := func(value, zh, en string) types.ConfigFieldOption {
		return types.ConfigFieldOption{
			Value:  value,
			Label:  zh,
			Labels: map[string]string{types.LocaleEnUS: en},
		}
	}
	return []types.ConfigFieldOption{
		opt("Asia/Shanghai", "Asia/Shanghai（东八区·中国）", "Asia/Shanghai (UTC+8 China)"),
		opt("Asia/Hong_Kong", "Asia/Hong_Kong（香港）", "Asia/Hong_Kong"),
		opt("Asia/Taipei", "Asia/Taipei（台北）", "Asia/Taipei"),
		opt("Asia/Tokyo", "Asia/Tokyo（东京）", "Asia/Tokyo"),
		opt("Asia/Seoul", "Asia/Seoul（首尔）", "Asia/Seoul"),
		opt("Asia/Singapore", "Asia/Singapore（新加坡）", "Asia/Singapore"),
		opt("Asia/Bangkok", "Asia/Bangkok（曼谷）", "Asia/Bangkok"),
		opt("Asia/Kolkata", "Asia/Kolkata（印度）", "Asia/Kolkata"),
		opt("Asia/Dubai", "Asia/Dubai（迪拜）", "Asia/Dubai"),
		opt("UTC", "UTC（协调世界时）", "UTC"),
		opt("Europe/London", "Europe/London（伦敦）", "Europe/London"),
		opt("Europe/Paris", "Europe/Paris（巴黎）", "Europe/Paris"),
		opt("Europe/Berlin", "Europe/Berlin（柏林）", "Europe/Berlin"),
		opt("Europe/Moscow", "Europe/Moscow（莫斯科）", "Europe/Moscow"),
		opt("America/New_York", "America/New_York（纽约）", "America/New_York"),
		opt("America/Chicago", "America/Chicago（芝加哥）", "America/Chicago"),
		opt("America/Denver", "America/Denver（丹佛）", "America/Denver"),
		opt("America/Los_Angeles", "America/Los_Angeles（洛杉矶）", "America/Los_Angeles"),
		opt("America/Sao_Paulo", "America/Sao_Paulo（圣保罗）", "America/Sao_Paulo"),
		opt("Australia/Sydney", "Australia/Sydney（悉尼）", "Australia/Sydney"),
		opt("Pacific/Auckland", "Pacific/Auckland（奥克兰）", "Pacific/Auckland"),
	}
}
