package database

import (
	"database/sql"
	"errors"
)

func ns(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func Seed(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM relation`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM person`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM family`); err != nil {
		return err
	}
	var seqName string
	if err := tx.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='sqlite_sequence'`).Scan(&seqName); err == nil {
		if _, err := tx.Exec(`DELETE FROM sqlite_sequence`); err != nil {
			return err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	insPerson := func(fid int64, name, gender, birthday string, gen int, phone, addr, notes *string, alive int, birthdayType string) (int64, error) {
		r, err := tx.Exec(
			`INSERT INTO person (family_id, name, gender, birthday, generation, phone, address, notes, is_alive) VALUES (?,?,?,?,?,?,?,?,?)`,
			fid, name, gender, birthday, gen, ns(phone), ns(addr), ns(notes), alive,
		)
		if err != nil {
			return 0, err
		}
		return r.LastInsertId()
	}

	insRelation := func(pid, rid int64, relType string) error {
		_, err := tx.Exec(`INSERT OR IGNORE INTO relation (person_id, related_id, type) VALUES (?,?,?)`, pid, rid, relType)
		return err
	}

	addCouple := func(a, b int64) error {
		if err := insRelation(a, b, "spouse"); err != nil {
			return err
		}
		return insRelation(b, a, "spouse")
	}

	addParentChild := func(parent, child int64) error {
		if err := insRelation(parent, child, "parent"); err != nil {
			return err
		}
		return insRelation(child, parent, "child")
	}

	insFamily := func(name, desc, webhook string) (int64, error) {
		r, err := tx.Exec(`INSERT INTO family (name, description, webhook_key) VALUES (?,?,?)`, name, desc, webhook)
		if err != nil {
			return 0, err
		}
		return r.LastInsertId()
	}

	f1, err := insFamily("张氏家族", "河南洛阳张氏，始祖张明远，堂号「百忍堂」", "")
	if err != nil {
		return err
	}

	p := func(s string) *string { return &s }

	zhangMingyuan, err := insPerson(f1, "张明远", "male", "1935-03-15", 1, p("13800001001"), p("河南省洛阳市老城区"), p("家族始迁祖，退休教师"), 0, "solar")
	if err != nil {
		return err
	}
	zhangXiulan, err := insPerson(f1, "张秀兰", "female", "1938-06-15", 1, nil, p("河南省洛阳市老城区"), p("农历六月十五"), 0, "lunar")
	if err != nil {
		return err
	}
	if err := addCouple(zhangMingyuan, zhangXiulan); err != nil {
		return err
	}

	zhangJianguo, err := insPerson(f1, "张建国", "male", "1962-01-10", 2, p("13800002001"), p("河南省洛阳市涧西区"), p("长子，工程师"), 1, "solar")
	if err != nil {
		return err
	}
	wangShufang, err := insPerson(f1, "王淑芳", "female", "1964-02-08", 2, p("13800002002"), p("河南省洛阳市涧西区"), p("建国之妻，会计，农历二月初八"), 1, "lunar")
	if err != nil {
		return err
	}
	zhangJianhua, err := insPerson(f1, "张建华", "male", "1965-08-03", 2, p("13800002003"), p("北京市朝阳区"), p("次子，IT创业者"), 1, "solar")
	if err != nil {
		return err
	}
	liMeiling, err := insPerson(f1, "李美玲", "female", "1968-08-15", 2, p("13800002004"), p("北京市朝阳区"), p("建华之妻，医生，农历八月十五中秋"), 1, "lunar")
	if err != nil {
		return err
	}
	zhangJianni, err := insPerson(f1, "张建妮", "female", "1970-04-12", 2, p("13800002005"), p("上海市浦东新区"), p("长女，大学教授"), 1, "solar")
	if err != nil {
		return err
	}
	chenWei, err := insPerson(f1, "陈伟", "male", "1969-09-08", 2, p("13800002006"), p("上海市浦东新区"), p("建妮之夫，律师"), 1, "solar")
	if err != nil {
		return err
	}

	if err := addCouple(zhangJianguo, wangShufang); err != nil {
		return err
	}
	if err := addCouple(zhangJianhua, liMeiling); err != nil {
		return err
	}
	if err := addCouple(zhangJianni, chenWei); err != nil {
		return err
	}
	for _, pc := range [][2]int64{{zhangMingyuan, zhangJianguo}, {zhangXiulan, zhangJianguo}, {zhangMingyuan, zhangJianhua}, {zhangXiulan, zhangJianhua}, {zhangMingyuan, zhangJianni}, {zhangXiulan, zhangJianni}} {
		if err := addParentChild(pc[0], pc[1]); err != nil {
			return err
		}
	}

	zhangHaoran, err := insPerson(f1, "张浩然", "male", "1990-06-15", 3, p("13800003001"), p("河南省洛阳市涧西区"), p("建国长子，银行职员"), 1, "solar")
	if err != nil {
		return err
	}
	zhangYuxin, err := insPerson(f1, "张雨欣", "female", "1993-01-20", 3, p("13800003002"), p("河南省洛阳市涧西区"), p("建国之女，设计师，农历正月二十"), 1, "lunar")
	if err != nil {
		return err
	}
	zhangZihan, err := insPerson(f1, "张子涵", "male", "1995-12-01", 3, p("13800003003"), p("北京市海淀区"), p("建华之子，研究生在读"), 1, "solar")
	if err != nil {
		return err
	}
	chenJiayi, err := insPerson(f1, "陈佳怡", "female", "1996-08-20", 3, p("13800003004"), p("上海市浦东新区"), p("建妮之女，留学生"), 1, "solar")
	if err != nil {
		return err
	}
	chenJiayang, err := insPerson(f1, "陈嘉阳", "male", "1999-03-10", 3, p("13800003005"), p("上海市浦东新区"), p("建妮之子，大四学生"), 1, "solar")
	if err != nil {
		return err
	}

	for _, pc := range [][2]int64{{zhangJianguo, zhangHaoran}, {wangShufang, zhangHaoran}, {zhangJianguo, zhangYuxin}, {wangShufang, zhangYuxin}, {zhangJianhua, zhangZihan}, {liMeiling, zhangZihan}, {zhangJianni, chenJiayi}, {chenWei, chenJiayi}, {zhangJianni, chenJiayang}, {chenWei, chenJiayang}} {
		if err := addParentChild(pc[0], pc[1]); err != nil {
			return err
		}
	}

	liuXiaojing, err := insPerson(f1, "刘晓静", "female", "1992-04-05", 3, p("13800003006"), p("河南省洛阳市涧西区"), p("浩然之妻，护士"), 1, "solar")
	if err != nil {
		return err
	}
	if err := addCouple(zhangHaoran, liuXiaojing); err != nil {
		return err
	}

	zhangRui, err := insPerson(f1, "张睿", "male", "2018-09-12", 4, nil, p("河南省洛阳市涧西区"), p("浩然之子，小学一年级"), 1, "solar")
	if err != nil {
		return err
	}
	zhangYihan, err := insPerson(f1, "张艺涵", "female", "2020-03-25", 4, nil, p("河南省洛阳市涧西区"), p("浩然之女，幼儿园"), 1, "solar")
	if err != nil {
		return err
	}
	for _, pc := range [][2]int64{{zhangHaoran, zhangRui}, {liuXiaojing, zhangRui}, {zhangHaoran, zhangYihan}, {liuXiaojing, zhangYihan}} {
		if err := addParentChild(pc[0], pc[1]); err != nil {
			return err
		}
	}

	f2, err := insFamily("李氏家族", "山东济南李氏，堂号「陇西堂」", "")
	if err != nil {
		return err
	}

	liDesheng, err := insPerson(f2, "李德盛", "male", "1940-02-14", 1, p("13900001001"), p("山东省济南市历下区"), p("家族族长"), 0, "solar")
	if err != nil {
		return err
	}
	liuGuiying, err := insPerson(f2, "刘桂英", "female", "1943-06-30", 1, nil, p("山东省济南市历下区"), nil, 0, "solar")
	if err != nil {
		return err
	}
	if err := addCouple(liDesheng, liuGuiying); err != nil {
		return err
	}

	liGuoqiang, err := insPerson(f2, "李国强", "male", "1968-10-05", 2, p("13900002001"), p("山东省济南市市中区"), p("长子，公务员"), 1, "solar")
	if err != nil {
		return err
	}
	zhaoMei, err := insPerson(f2, "赵梅", "female", "1970-12-18", 2, p("13900002002"), p("山东省济南市市中区"), p("国强之妻，教师"), 1, "solar")
	if err != nil {
		return err
	}
	liGuohua, err := insPerson(f2, "李国华", "male", "1972-03-22", 2, p("13900002003"), p("广东省深圳市"), p("次子，程序员"), 1, "solar")
	if err != nil {
		return err
	}
	huangLili, err := insPerson(f2, "黄丽丽", "female", "1975-08-15", 2, p("13900002004"), p("广东省深圳市"), p("国华之妻，产品经理"), 1, "solar")
	if err != nil {
		return err
	}

	if err := addCouple(liGuoqiang, zhaoMei); err != nil {
		return err
	}
	if err := addCouple(liGuohua, huangLili); err != nil {
		return err
	}
	for _, pc := range [][2]int64{{liDesheng, liGuoqiang}, {liuGuiying, liGuoqiang}, {liDesheng, liGuohua}, {liuGuiying, liGuohua}} {
		if err := addParentChild(pc[0], pc[1]); err != nil {
			return err
		}
	}

	liMinghao, err := insPerson(f2, "李明浩", "male", "1998-01-08", 3, p("13900003001"), p("山东省济南市市中区"), p("国强之子，大三学生"), 1, "solar")
	if err != nil {
		return err
	}
	liMingyu, err := insPerson(f2, "李明宇", "male", "2001-07-14", 3, nil, p("广东省深圳市"), p("国华之子，高三学生"), 1, "solar")
	if err != nil {
		return err
	}
	for _, pc := range [][2]int64{{liGuoqiang, liMinghao}, {zhaoMei, liMinghao}, {liGuohua, liMingyu}, {huangLili, liMingyu}} {
		if err := addParentChild(pc[0], pc[1]); err != nil {
			return err
		}
	}

	_, err = insFamily("王氏家族", "广东潮汕王氏，堂号「三槐堂」（待录入成员）", "")
	if err != nil {
		return err
	}

	return tx.Commit()
}
