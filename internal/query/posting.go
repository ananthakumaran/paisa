package query

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Query struct {
	context *gorm.DB
	order   string
}

func Init(db *gorm.DB) *Query {
	return &Query{context: db, order: "ASC"}
}

func (q *Query) Desc() *Query {
	q.order = "DESC"
	return q
}

func (q *Query) Limit(n int) *Query {
	q.context = q.context.Limit(n)
	return q
}

func (q *Query) Clone() *Query {
	return &Query{context: q.context.Session(&gorm.Session{}), order: q.order}
}

func (q *Query) BeforeNMonths(n int) *Query {
	monthStart := utils.BeginningOfMonth(time.Now())
	start := monthStart.AddDate(0, -(n - 1), 0)
	q.context = q.context.Where("date < ?", start)
	return q
}

func (q *Query) UntilToday() *Query {
	q.context = q.context.Where("date < ?", time.Now())
	return q
}

func (q *Query) LastNMonths(n int) *Query {
	monthStart := utils.BeginningOfMonth(time.Now())
	start := monthStart.AddDate(0, -(n - 1), 0)
	end := monthStart.AddDate(0, 1, 0)
	q.context = q.context.Where("date >= ? and date < ?", start, end)
	return q
}

func (q *Query) Commodities(commodities []config.Commodity) *Query {
	q.context = q.context.Where("commodity in ?", lo.Map(commodities, func(c config.Commodity, _ int) string { return c.Name }))
	return q
}

func (q *Query) Status(status string) *Query {
	if status == "cleared" || status == "pending" {
		q.context = q.context.Where("status = ?", status)
	}
	return q
}

func (q *Query) Credit() *Query {
	q.context = q.context.Where("amount > 0")
	return q
}

func (q *Query) AccountPrefix(account string) *Query {
	q.context = q.context.Where("account like ? or account = ?", account+":%", account)
	return q
}

func (q *Query) Like(account string) *Query {
	q.context = q.context.Where("account like ?", account)
	return q
}

func (q *Query) OrLike(account string) *Query {
	q.context = q.context.Or("account like ?", account)
	return q
}

func (q *Query) NotLike(account string) *Query {
	q.context = q.context.Where("account not like ?", account)
	return q
}

func (q *Query) Or(query interface{}, args ...interface{}) *Query {
	q.context = q.context.Or(query, args...)
	return q
}

func (q *Query) Where(query interface{}, args ...interface{}) *Query {
	q.context = q.context.Where(query, args...)
	return q
}

func (q *Query) All() []posting.Posting {
	var postings []posting.Posting
	result := q.context.Order("date " + q.order + ", amount desc").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return postings
}

func (q *Query) First() posting.Posting {
	var posting posting.Posting
	result := q.context.Order("date " + q.order + ", amount desc").First(&posting)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return posting
}
