const Database = require('better-sqlite3');
const path = require('path');

const dbPath = path.join(__dirname, 'family.db');
const db = new Database(dbPath);

// 开启 WAL 模式和外键
db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

// 建表 (与 server/database/db.go 一致)
db.exec(`
  CREATE TABLE IF NOT EXISTS family (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT,
    webhook_key TEXT DEFAULT '',
    created_at  DATETIME DEFAULT (datetime('now','localtime')),
    updated_at  DATETIME DEFAULT (datetime('now','localtime'))
  );
  CREATE TABLE IF NOT EXISTS person (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    family_id   INTEGER NOT NULL,
    name        TEXT NOT NULL,
    gender      TEXT CHECK(gender IN ('male','female','unknown')) DEFAULT 'unknown',
    birthday    TEXT,
    birthday_type TEXT CHECK(birthday_type IN ('solar','lunar')) DEFAULT 'solar',
    generation  INTEGER,
    photo_url   TEXT,
    phone       TEXT,
    address     TEXT,
    notes       TEXT,
    is_alive    INTEGER DEFAULT 1,
    created_at  DATETIME DEFAULT (datetime('now','localtime')),
    updated_at  DATETIME DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (family_id) REFERENCES family(id) ON DELETE CASCADE
  );
  CREATE TABLE IF NOT EXISTS relation (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id   INTEGER NOT NULL,
    related_id  INTEGER NOT NULL,
    type        TEXT NOT NULL CHECK(type IN ('parent','child','spouse','sibling','in_law','grandparent','grandchild')),
    created_at  DATETIME DEFAULT (datetime('now','localtime')),
    FOREIGN KEY (person_id) REFERENCES person(id) ON DELETE CASCADE,
    FOREIGN KEY (related_id) REFERENCES person(id) ON DELETE CASCADE,
    UNIQUE(person_id, related_id, type)
  );
  CREATE INDEX IF NOT EXISTS idx_person_family ON person(family_id);
  CREATE INDEX IF NOT EXISTS idx_relation_person ON relation(person_id);
  CREATE INDEX IF NOT EXISTS idx_relation_related ON relation(related_id);
  CREATE INDEX IF NOT EXISTS idx_person_birthday ON person(birthday);
`);

// ============ 清空旧数据 ============
db.exec('DELETE FROM relation');
db.exec('DELETE FROM person');
db.exec('DELETE FROM family');
db.exec("DELETE FROM sqlite_sequence");

// ============ 家族 1: 张氏家族 (4代人，完整) ============
const f1 = db.prepare('INSERT INTO family (name, description, webhook_key) VALUES (?, ?, ?)')
  .run('张氏家族', '河南洛阳张氏，始祖张明远，堂号「百忍堂」', '').lastInsertRowid;

// Helper
function insPerson(fid, name, gender, birthday, gen, phone, addr, notes, alive = 1, birthdayType = 'solar') {
  return db.prepare(
    'INSERT INTO person (family_id, name, gender, birthday, birthday_type, generation, phone, address, notes, is_alive) VALUES (?,?,?,?,?,?,?,?,?,?)'
  ).run(fid, name, gender, birthday, birthdayType, gen, phone, addr, notes, alive).lastInsertRowid;
}

function insRelation(pid, rid, type) {
  try {
    db.prepare('INSERT INTO relation (person_id, related_id, type) VALUES (?,?,?)').run(pid, rid, type);
  } catch(e) { /* skip duplicates */ }
}

function addCouple(a, b) {
  insRelation(a, b, 'spouse');
  insRelation(b, a, 'spouse');
}

function addParentChild(parent, child) {
  insRelation(parent, child, 'parent');
  insRelation(child, parent, 'child');
}

function addSiblings(ids) {
  for (let i = 0; i < ids.length; i++) {
    for (let j = i + 1; j < ids.length; j++) {
      insRelation(ids[i], ids[j], 'sibling');
      insRelation(ids[j], ids[i], 'sibling');
    }
  }
}

// 第1代: 祖父母
const zhang_mingyuan = insPerson(f1, '张明远', 'male', '1935-03-15', 1, '13800001001', '河南省洛阳市老城区', '家族始迁祖，退休教师', 0);
const zhang_xiulan  = insPerson(f1, '张秀兰', 'female', '1938-06-15', 1, null, '河南省洛阳市老城区', '农历六月十五', 0, 'lunar');
addCouple(zhang_mingyuan, zhang_xiulan);

// 第2代: 三个子女 + 配偶
const zhang_jianguo = insPerson(f1, '张建国', 'male', '1962-01-10', 2, '13800002001', '河南省洛阳市涧西区', '长子，工程师');
const wang_shufang  = insPerson(f1, '王淑芳', 'female', '1964-02-08', 2, '13800002002', '河南省洛阳市涧西区', '建国之妻，会计，农历二月初八', 1, 'lunar');
const zhang_jianhua = insPerson(f1, '张建华', 'male', '1965-08-03', 2, '13800002003', '北京市朝阳区', '次子，IT创业者');
const li_meiling    = insPerson(f1, '李美玲', 'female', '1968-08-15', 2, '13800002004', '北京市朝阳区', '建华之妻，医生，农历八月十五中秋', 1, 'lunar');
const zhang_jianni  = insPerson(f1, '张建妮', 'female', '1970-04-12', 2, '13800002005', '上海市浦东新区', '长女，大学教授');
const chen_wei      = insPerson(f1, '陈伟', 'male', '1969-09-08', 2, '13800002006', '上海市浦东新区', '建妮之夫，律师');

addCouple(zhang_jianguo, wang_shufang);
addCouple(zhang_jianhua, li_meiling);
addCouple(zhang_jianni, chen_wei);
addSiblings([zhang_jianguo, zhang_jianhua, zhang_jianni]);

// 祖父母 -> 子女
[zhang_jianguo, zhang_jianhua, zhang_jianni].forEach(c => {
  addParentChild(zhang_mingyuan, c);
  addParentChild(zhang_xiulan, c);
});

// 第3代: 5个孙辈
const zhang_haoran  = insPerson(f1, '张浩然', 'male', '1990-06-15', 3, '13800003001', '河南省洛阳市涧西区', '建国长子，银行职员');
const zhang_yuxin   = insPerson(f1, '张雨欣', 'female', '1993-01-20', 3, '13800003002', '河南省洛阳市涧西区', '建国之女，设计师，农历正月二十', 1, 'lunar');
const zhang_zihan   = insPerson(f1, '张子涵', 'male', '1995-12-01', 3, '13800003003', '北京市海淀区', '建华之子，研究生在读');
const chen_jiayi    = insPerson(f1, '陈佳怡', 'female', '1996-08-20', 3, '13800003004', '上海市浦东新区', '建妮之女，留学生');
const chen_jiayang  = insPerson(f1, '陈嘉阳', 'male', '1999-03-10', 3, '13800003005', '上海市浦东新区', '建妮之子，大四学生');

addSiblings([zhang_haoran, zhang_yuxin]);
addSiblings([chen_jiayi, chen_jiayang]);

// 父母 -> 子女
addParentChild(zhang_jianguo, zhang_haoran);
addParentChild(wang_shufang, zhang_haoran);
addParentChild(zhang_jianguo, zhang_yuxin);
addParentChild(wang_shufang, zhang_yuxin);
addParentChild(zhang_jianhua, zhang_zihan);
addParentChild(li_meiling, zhang_zihan);
addParentChild(zhang_jianni, chen_jiayi);
addParentChild(chen_wei, chen_jiayi);
addParentChild(zhang_jianni, chen_jiayang);
addParentChild(chen_wei, chen_jiayang);

// 祖父母 -> 孙辈
[zhang_haoran, zhang_yuxin, zhang_zihan, chen_jiayi, chen_jiayang].forEach(g => {
  addParentChild(zhang_mingyuan, g);
  addParentChild(zhang_xiulan, g);
  insRelation(zhang_mingyuan, g, 'grandparent');
  insRelation(g, zhang_mingyuan, 'grandchild');
  insRelation(zhang_xiulan, g, 'grandparent');
  insRelation(g, zhang_xiulan, 'grandchild');
});

// 第3代配偶
const liu_xiaojing = insPerson(f1, '刘晓静', 'female', '1992-04-05', 3, '13800003006', '河南省洛阳市涧西区', '浩然之妻，护士');
addCouple(zhang_haoran, liu_xiaojing);
addParentChild(liu_xiaojing, zhang_mingyuan); // 姻亲（简化）

// 第4代
const zhang_rui    = insPerson(f1, '张睿', 'male', '2018-09-12', 4, null, '河南省洛阳市涧西区', '浩然之子，小学一年级');
const zhang_yihan  = insPerson(f1, '张艺涵', 'female', '2020-03-25', 4, null, '河南省洛阳市涧西区', '浩然之女，幼儿园');
addSiblings([zhang_rui, zhang_yihan]);
addParentChild(zhang_haoran, zhang_rui);
addParentChild(liu_xiaojing, zhang_rui);
addParentChild(zhang_haoran, zhang_yihan);
addParentChild(liu_xiaojing, zhang_yihan);

// 姻亲关系
insRelation(zhang_jianguo, li_meiling, 'in_law');
insRelation(zhang_jianhua, wang_shufang, 'in_law');

// ============ 家族 2: 李氏家族 (3代人，部分数据) ============
const f2 = db.prepare('INSERT INTO family (name, description, webhook_key) VALUES (?, ?, ?)')
  .run('李氏家族', '山东济南李氏，堂号「陇西堂」', '').lastInsertRowid;

const li_desheng   = insPerson(f2, '李德盛', 'male', '1940-02-14', 1, '13900001001', '山东省济南市历下区', '家族族长', 0);
const liu_guiying  = insPerson(f2, '刘桂英', 'female', '1943-06-30', 1, null, '山东省济南市历下区', null, 0);
addCouple(li_desheng, liu_guiying);

const li_guoqiang  = insPerson(f2, '李国强', 'male', '1968-10-05', 2, '13900002001', '山东省济南市市中区', '长子，公务员');
const zhao_mei     = insPerson(f2, '赵梅', 'female', '1970-12-18', 2, '13900002002', '山东省济南市市中区', '国强之妻，教师');
const li_guohua    = insPerson(f2, '李国华', 'male', '1972-03-22', 2, '13900002003', '广东省深圳市', '次子，程序员');
const huang_lili   = insPerson(f2, '黄丽丽', 'female', '1975-08-15', 2, '13900002004', '广东省深圳市', '国华之妻，产品经理');

addCouple(li_guoqiang, zhao_mei);
addCouple(li_guohua, huang_lili);
addSiblings([li_guoqiang, li_guohua]);
addParentChild(li_desheng, li_guoqiang);
addParentChild(liu_guiying, li_guoqiang);
addParentChild(li_desheng, li_guohua);
addParentChild(liu_guiying, li_guohua);

const li_minghao   = insPerson(f2, '李明浩', 'male', '1998-01-08', 3, '13900003001', '山东省济南市市中区', '国强之子，大三学生');
const li_mingyu    = insPerson(f2, '李明宇', 'male', '2001-07-14', 3, null, '广东省深圳市', '国华之子，高三学生');
addParentChild(li_guoqiang, li_minghao);
addParentChild(zhao_mei, li_minghao);
addParentChild(li_guohua, li_mingyu);
addParentChild(huang_lili, li_mingyu);

// ============ 家族 3: 王氏家族 (只有创始人，空家族) ============
const f3 = db.prepare('INSERT INTO family (name, description, webhook_key) VALUES (?, ?, ?)')
  .run('王氏家族', '广东潮汕王氏，堂号「三槐堂」（待录入成员）', '').lastInsertRowid;

// ============ 验证 ============
const famCount = db.prepare('SELECT COUNT(*) as c FROM family').get().c;
const personCount = db.prepare('SELECT COUNT(*) as c FROM person').get().c;
const relCount = db.prepare('SELECT COUNT(*) as c FROM relation').get().c;

console.log(`✅ 测试数据生成完成`);
console.log(`   家族: ${famCount}`);
console.log(`   成员: ${personCount}`);
console.log(`   关系: ${relCount}`);
console.log(`   数据库: ${dbPath}`);

db.close();
