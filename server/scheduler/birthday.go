package scheduler

import (
	"fmt"
	"log"
	"time"

	"family-tree/model"
	"family-tree/pkg"

	"github.com/robfig/cron/v3"
)

type BirthdayScheduler struct {
	cron       *cron.Cron
	personRepo *model.PersonRepo
	familyRepo *model.FamilyRepo
	wechatURL  string
}

func NewBirthdayScheduler(
	personRepo *model.PersonRepo,
	familyRepo *model.FamilyRepo,
	wechatURL string,
	cronExpr string,
) *BirthdayScheduler {
	c := cron.New(cron.WithLocation(time.Local))

	s := &BirthdayScheduler{
		cron:       c,
		personRepo: personRepo,
		familyRepo: familyRepo,
		wechatURL:  wechatURL,
	}

	c.AddFunc(cronExpr, s.checkAndNotify)

	return s
}

func (s *BirthdayScheduler) Start() {
	log.Println("birthday scheduler started")
	s.cron.Start()
}

func (s *BirthdayScheduler) Stop() {
	s.cron.Stop()
	log.Println("birthday scheduler stopped")
}

func (s *BirthdayScheduler) checkAndNotify() {
	log.Println("checking today's birthdays...")

	// 获取所有配了 webhook 的家族
	families, err := s.familyRepo.ListWithWebhook()
	if err != nil {
		log.Printf("failed to list families with webhook: %v", err)
		return
	}

	if len(families) == 0 {
		log.Println("no families with webhook configured, skip")
		return
	}

	now := time.Now()

	for _, family := range families {
		// 查这个家族今天生日的人
		persons, err := s.personRepo.GetBirthdayTodayByFamily(family.ID)
		if err != nil {
			log.Printf("failed to get birthdays for family %s: %v", family.Name, err)
			continue
		}

		if len(persons) == 0 {
			log.Printf("family %s: no birthdays today", family.Name)
			continue
		}

		// 构建消息
		content := fmt.Sprintf("🎂 **%s - 今日生日提醒**\n\n", family.Name)
		for _, p := range persons {
			age := ""
			if len(p.Birthday) >= 4 {
				birthYear := 0
				fmt.Sscanf(p.Birthday[:4], "%d", &birthYear)
				if birthYear > 0 {
					age = fmt.Sprintf("（%d岁）", now.Year()-birthYear)
				}
			}
			content += fmt.Sprintf("> %s %s\n", p.Name, age)
		}
		content += "\n记得送上祝福！🎉"

		// 用这个家族自己的 webhook 发送
		wechat := pkg.NewWechatWebhook(s.wechatURL, family.WebhookKey)
		if err := wechat.SendMarkdown(content); err != nil {
			log.Printf("failed to send birthday notification for family %s: %v", family.Name, err)
			continue
		}

		log.Printf("birthday notification sent for family %s: %d person(s)", family.Name, len(persons))
	}
}

// TriggerCheck 手动触发检查（用于测试）
func (s *BirthdayScheduler) TriggerCheck() {
	log.Println("manual birthday check triggered")
	s.checkAndNotify()
}
